package istio

import (
	"context"
	"fmt"
	v1 "github.com/kubeflow/kubeflow/v2/apis/core/v1"
	securityv1beta1 "istio.io/api/security/v1beta1"
	apissecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const AUTHZPOLICYISTIO string = "ns-owner-access-istio"

func Setup(mgr ctrl.Manager, o controller.Options) error {
	return nil
}

type Reconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	profile := &v1.Profile{}
	config := &v1.ProfileConfig{}

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
				Key:    fmt.Sprintf("reqest.headers[%s]", config.Spec.User.IDHeader),
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
