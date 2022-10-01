package secret

import (
	"context"
	"fmt"
	xpevent "github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/kubeflow/kubeflow/v2/apis/core/v1"
	"github.com/kubeflow/kubeflow/v2/internal/controller/reconciler/profile/watch"
	"github.com/kubeflow/kubeflow/v2/internal/logging"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	reasonCreatedSecret = xpevent.Reason("CreatedSecret")
	reasonUpdatedSecret = xpevent.Reason("UpdatedSecret")
)

func Setup(mgr ctrl.Manager, o controller.Options) error {

	name := "kubeflow/reflect/secret"

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1.Profile{}).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
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
	r.logger.Debug("using profile", "name", profile.Name)

	secretList := &corev1.SecretList{}
	if err := r.client.List(ctx, secretList, client.MatchingLabels{labelReflect: "true"}, client.InNamespace("kubeflow")); err != nil {
		r.logger.Debug("failed to list secrets", "error", err.Error())
		return ctrl.Result{}, err
	}

	r.logger.Debug("found", "count", len(secretList.Items))

	for _, item := range secretList.Items {
		secret := &corev1.Secret{}
		secret.SetName(item.Name)
		secret.SetNamespace(profile.Name)
		res, err := controllerutil.CreateOrPatch(ctx, r.client, secret, func() error {
			if err := controllerutil.SetControllerReference(profile, secret, r.scheme); err != nil {
				return err
			}
			secret.Data = item.Data
			secret.StringData = item.StringData
			secret.Labels = item.Labels
			secret.Annotations = item.Annotations
			secret.Immutable = item.Immutable
			return nil
		})
		if err != nil {
			return ctrl.Result{}, err
		}
		switch res {
		case controllerutil.OperationResultCreated:
			r.record.Event(profile, xpevent.Normal(reasonCreatedSecret, "reflected", "from", item.Name))
		case controllerutil.OperationResultUpdated:
			r.record.Event(profile, xpevent.Normal(reasonUpdatedSecret, "reflected", "from", item.Name))
		}
	}
	return ctrl.Result{}, nil
}

var (
	labelReflect = fmt.Sprintf("profile.%s/reflect", v1.Group)
)
