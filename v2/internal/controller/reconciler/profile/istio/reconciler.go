package istio

import (
	"context"
	"fmt"
	xpevent "github.com/crossplane/crossplane-runtime/pkg/event"
	v1 "github.com/kubeflow/kubeflow/v2/apis/core/v1"
	"github.com/kubeflow/kubeflow/v2/internal/controller/reconciler/profile/watch"
	"github.com/kubeflow/kubeflow/v2/internal/logging"
	securityv1beta1 "istio.io/api/security/v1beta1"
	apissecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strings"
)

const AUTHZPOLICYISTIO string = "ns-owner-access-istio"

func Setup(mgr ctrl.Manager, o controller.Options) error {

	name := "kubeflow/" + strings.ToLower(v1.ProfileKind) + "/istio"

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1.Profile{}).
		Owns(&apissecurityv1beta1.AuthorizationPolicy{}).
		Watches(
			&source.Kind{Type: &v1.ProfileConfig{}},
			&watch.EnqueueRequestForProfiles{Reader: mgr.GetClient()},
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
	config := &v1.ProfileConfig{}
	if err := r.client.Get(ctx, client.ObjectKey{Name: "default"}, config); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	authz := &apissecurityv1beta1.AuthorizationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AUTHZPOLICYISTIO,
			Namespace: profile.Name,
		},
	}
	res, err := controllerutil.CreateOrPatch(ctx, r.client, authz, func() error {
		if err := controllerutil.SetControllerReference(profile, authz, r.scheme); err != nil {
			return err
		}
		if authz.Annotations == nil {
			authz.Annotations = make(map[string]string)
		}
		authz.Annotations["user"] = profile.Spec.Owner.Name
		authz.Spec.Action = securityv1beta1.AuthorizationPolicy_ALLOW
		authz.Spec.Rules = []*securityv1beta1.Rule{{
			When: []*securityv1beta1.Condition{{
				Key:    fmt.Sprintf("request.headers[%s]", config.Spec.User.IDHeader),
				Values: []string{config.Spec.User.IDPrefix + profile.Spec.Owner.Name},
			}},
		}, {
			When: []*securityv1beta1.Condition{{
				Key:    "source.namespace",
				Values: []string{profile.Name},
			}},
		}, {
			To: []*securityv1beta1.Rule_To{{
				Operation: &securityv1beta1.Operation{
					Paths: []string{
						"/healthz",
						"/metrics",
						"/wait-for-drain",
					},
				},
			}},
		}, {
			From: []*securityv1beta1.Rule_From{{
				Source: &securityv1beta1.Source{
					Principals: []string{
						"cluster.local/ns/kubeflow/sa/notebook-controller-service-account",
					},
				},
			}},
			To: []*securityv1beta1.Rule_To{{
				Operation: &securityv1beta1.Operation{
					Methods: []string{"GET"},
					Paths:   []string{"*/api/kernels"}, // wildcard for the name of the notebook server
				},
			}},
		}}
		return nil
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	switch res {
	case controllerutil.OperationResultCreated:
	case controllerutil.OperationResultUpdated:
	}
	return ctrl.Result{}, nil
}
