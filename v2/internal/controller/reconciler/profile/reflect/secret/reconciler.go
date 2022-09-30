package secret

import (
	"context"
	"fmt"
	xpevent "github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/kubeflow/kubeflow/v2/apis/core/v1"
	"github.com/kubeflow/kubeflow/v2/internal/controller/reconciler/profile/watch"
	"github.com/kubeflow/kubeflow/v2/internal/logging"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {

	name := "kubeflow/"

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		// For(
		//   builder.WithPredicates(watch.LabelPredicate(labelReflect, "enabled")),
		// ).
		Watches(
			&source.Kind{Type: &v1.Profile{}},
			&watch.EnqueueRequestForProfiles{Reader: mgr.GetClient()},
		).
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
	return ctrl.Result{}, nil
}

var (
	labelReflect = fmt.Sprintf("profile.%s/reflect", v1.Group)
)

