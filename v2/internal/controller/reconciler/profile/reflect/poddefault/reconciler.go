package poddefault

import (
	"context"
	"fmt"
	xpevent "github.com/crossplane/crossplane-runtime/pkg/event"
	v1 "github.com/kubeflow/kubeflow/v2/apis/core/v1"
	"github.com/kubeflow/kubeflow/v2/internal/controller/reconciler/profile/watch"
	"github.com/kubeflow/kubeflow/v2/internal/logging"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	reasonPodDefaultCreated xpevent.Reason = "CreatedPodDefault"
	reasonPodDefaultUpdated xpevent.Reason = "UpdatedPodDefault"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {

	const name = "kubeflow/reflect/configmap"

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1.Profile{}).
		Watches(
			&source.Kind{Type: &v1.PodDefault{}},
			&watch.EnqueueRequestForProfiles{Reader: mgr.GetClient()},
			builder.WithPredicates(
				watch.LabelPredicate(labelReflect, "true"),
				watch.InNamespace("kubeflow"),
			),
		).
		WithOptions(o).
		Complete(NewReconciler(mgr,
			WithLogger(logging.NewLogrLogger(mgr.GetLogger().WithValues("controller", name))),
			WithEventRecorder(xpevent.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

type ReconcilerOption func(r *Reconciler)

func WithLogger(log logging.Logger) ReconcilerOption {
	return func(r *Reconciler) {
		r.logger = log
	}
}

func WithEventRecorder(recorder xpevent.Recorder) ReconcilerOption {
	return func(r *Reconciler) {
		r.record = recorder
	}
}

func NewReconciler(mgr ctrl.Manager, opts ...ReconcilerOption) *Reconciler {
	r := &Reconciler{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		logger: logging.NewNopLogger(),
		record: xpevent.NewNopRecorder(),
	}
	for _, fn := range opts {
		fn(r)
	}
	return r
}

type Reconciler struct {
	client client.Client
	scheme *runtime.Scheme
	logger logging.Logger
	record xpevent.Recorder
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	profile := &v1.Profile{}
	if err := r.client.Get(ctx, req.NamespacedName, profile); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	podDefaultList := &v1.PodDefaultList{}
	if err := r.client.List(ctx, podDefaultList, client.MatchingLabels{labelReflect: "true"}, client.InNamespace("kubeflow")); err != nil {
		return ctrl.Result{}, err
	}

	for _, item := range podDefaultList.Items {
		podDefault := &v1.PodDefault{}
		podDefault.SetName(item.Name)
		podDefault.SetNamespace(profile.Name)
		res, err := controllerutil.CreateOrPatch(ctx, r.client, podDefault, func() error {
			if err := controllerutil.SetControllerReference(profile, podDefault, r.scheme); err != nil {
				return err
			}
			podDefault.Annotations = item.Annotations
			podDefault.Labels = item.Labels
			podDefault.Spec = item.Spec
			return nil
		})
		if err != nil {
			return ctrl.Result{}, err
		}
		switch res {
		case controllerutil.OperationResultCreated:
			r.record.Event(profile, xpevent.Normal(reasonPodDefaultCreated, "reflected", "from", item.Name))
		case controllerutil.OperationResultUpdated:
			r.record.Event(profile, xpevent.Normal(reasonPodDefaultUpdated, "reflected", "from", item.Name))
		}
	}
	return ctrl.Result{}, nil
}

var labelReflect = fmt.Sprintf("profile.%s/reflect", v1.Group)
