package culler

import (
	"context"
	"fmt"
	xpevent "github.com/crossplane/crossplane-runtime/pkg/event"
	v1 "github.com/kubeflow/kubeflow/v2/apis/core/v1"
	"github.com/kubeflow/kubeflow/v2/internal/logging"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/pointer"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strings"
	"time"
)

const (
	// AnnotationStop is set to tell the notebook controller
	// not to scale up the notebook server
	AnnotationStop           = "kubeflow-resource-stopped"
	annotationLatestActivity = "notebooks.kubeflow.org/last-activity"
	annotationIgnoreCuller   = "culler.notebooks.kubeflow.org/ignore"

	kernelExecutorStateIdle     = "idle"
	kernelExecutorStateBusy     = "busy"
	kernelExecutorStateStarting = "starting"
)

func Setup(mgr ctrl.Manager, o controller.Options, opts ...ReconcilerOption) error {
	name := "kubeflow/" + strings.ToLower(v1.NotebookCullerKind)

	options := []ReconcilerOption{
		WithLogger(logging.NewLogrLogger(mgr.GetLogger().WithValues("controller", name))),
		WithEventRecorder(xpevent.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		WithJupyterClient(&Client{client: http.DefaultClient}),
	}
	options = append(options, opts...)

	c := ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1.Notebook{}).
		Watches(
			&source.Kind{Type: &appsv1.StatefulSet{}},
			&handler.EnqueueRequestForOwner{OwnerType: &v1.Notebook{}, IsController: true},
			builder.WithPredicates(stsPredicate),
		).
		Watches(
			&source.Kind{Type: &v1.NotebookCuller{}},
			&EnqueueRequestForNotebooks{reader: mgr.GetClient()},
		).
		WithOptions(o).
		Complete(NewReconciler(mgr, options...))
	return c
}

type ReconcilerOption func(r *Reconciler)

func WithLogger(l logging.Logger) ReconcilerOption {
	return func(r *Reconciler) {
		r.logger = l
	}
}

func WithEventRecorder(rec xpevent.Recorder) ReconcilerOption {
	return func(r *Reconciler) {
		r.record = rec
	}
}

func WithJupyterClient(c jupyter) ReconcilerOption {
	return func(r *Reconciler) {
		r.jp = c
	}
}

func NewReconciler(mgr ctrl.Manager, opts ...ReconcilerOption) *Reconciler {
	r := &Reconciler{
		client: mgr.GetClient(),
		logger: logging.NewNopLogger(),
		record: xpevent.NewNopRecorder(),
		jp:     &Client{client: http.DefaultClient},
	}

	for _, fn := range opts {
		fn(r)
	}

	return r
}

type Reconciler struct {
	client client.Client
	logger logging.Logger
	record xpevent.Recorder
	jp     jupyter
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	notebook := &v1.Notebook{}
	if err := r.client.Get(ctx, req.NamespacedName, notebook); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !notebook.GetDeletionTimestamp().IsZero() {
		return ctrl.Result{}, nil
	}

	annotations := notebook.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}
	if annotations[annotationIgnoreCuller] == "true" {
		return ctrl.Result{}, nil
	}

	try := map[string]client.ObjectKey{
		"notebook": {
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		"namespace": {
			Name:      "default",
			Namespace: req.Namespace,
		},
		"global": {
			Name:      "default",
			Namespace: "kubeflow",
		},
	}

	found := false
	config := &v1.NotebookCuller{}
	for k, key := range try {
		if err := r.client.Get(ctx, key, config); err != nil {
			if !apierrors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
			continue
		}
		found = true
		r.logger.Debug(fmt.Sprintf("using %s cull config", k))
	}
	if !found {
		r.logger.Debug("culling disabled. No config found")
		return ctrl.Result{}, nil
	}

	// check the replica count -- if the replicas are 0 then no need
	// to cull it
	sts := &appsv1.StatefulSet{}
	if err := r.client.Get(ctx, req.NamespacedName, sts); err != nil {
		return ctrl.Result{}, err
	}
	// This should never be nil, it should have a default of
	// 1 if not specified
	if pointer.Int32Deref(sts.Spec.Replicas, 0) == 0 {
		// Nothing to cull
		return ctrl.Result{}, nil
	}

	// How does it mean to culling a notebook
	// Why do I care about the latest activity?
	// latest activity < time.Now() - cull period
	// - I don't necessarily need annotations for this, nor
	//   do I care about if the notebook is idle.

	// If the notebook is idle, how long has it been idle
	tm, err := r.jp.LatestActivity(notebook)
	if err != nil {
		return ctrl.Result{}, err
	}

	dur, err := time.ParseDuration(config.Options.Duration)
	if err != nil {
		// This should be covered by a validating webhook, but it can
		// fail
		// The config is bad, requeue won't help
		return ctrl.Result{}, nil
	}

	// latest plus the cull duration is < now
	if (*tm).Add(dur).Before(time.Now()) {
		// Cull the notebook
		patch := client.MergeFrom(sts.DeepCopy())
		sts.Spec.Replicas = pointer.Int32(0)
		if sts.Annotations == nil {
			sts.Annotations = make(map[string]string)
		}
		sts.Annotations[AnnotationStop] = "true"
		if err := r.client.Patch(ctx, sts, patch); err != nil {
			return ctrl.Result{}, err
		}
	}

	inv, err := time.ParseDuration(config.Options.Interval)
	if err != nil {
		// the config is bad, nothing to do
		return ctrl.Result{}, nil
	}

	return ctrl.Result{RequeueAfter: inv / 2}, nil
}
