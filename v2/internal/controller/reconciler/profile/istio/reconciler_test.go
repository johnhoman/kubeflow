package istio

import (
	"github.com/kubeflow/kubeflow/v2/apis/core/v1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = ginkgo.Describe("Reconciler", func() {

	ginkgo.It("creates configmaps in all profiles", func() {
		profiles := []client.Object{
			&v1.Profile{
				ObjectMeta: metav1.ObjectMeta{
					Name: "starlord",
				},
				Spec: v1.ProfileSpec{
					Owner: rbacv1.Subject{
						Kind: "User",
						Name: "starlord@guardians.com",
					},
				},
			},
			&v1.Profile{
				ObjectMeta: metav1.ObjectMeta{
					Name: "drax",
				},
				Spec: v1.ProfileSpec{
					Owner: rbacv1.Subject{
						Kind: "User",
						Name: "drax@guardians.com",
					},
				},
			},
			&v1.Profile{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gamora",
				},
				Spec: v1.ProfileSpec{
					Owner: rbacv1.Subject{
						Kind: "User",
						Name: "gamora@guardians.com",
					},
				},
			},
		}

		for _, item := range profiles {
			gomega.Expect(k8s.Create(ctx, item)).Should(gomega.Succeed())
			ns := &corev1.Namespace{}
			ns.SetName(item.GetName())
			gomega.Expect(k8s.Create(ctx, ns)).Should(gomega.Succeed())
			gomega.Eventually(komega.Get(ns)).Should(gomega.Succeed())
		}

		config := &v1.ProfileConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
			Spec: v1.ProfileConfigSpec{
				User: v1.ProfileConfigUserSpec{
					IDHeader: "kubeflow-userid",
				},
			},
		}
		gomega.Expect(k8s.Create(ctx, config)).Should(gomega.Succeed())
		gomega.Eventually(komega.Get(config)).Should(gomega.Succeed())

		for _, item := range profiles {
			vs := &securityv1beta1.AuthorizationPolicy{}
			vs.SetName(AUTHZPOLICYISTIO)
			vs.SetNamespace(item.GetName())
			gomega.Eventually(komega.Get(vs)).Should(gomega.Succeed())
		}
	})
})
