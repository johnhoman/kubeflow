package profile

import (
	"github.com/brianvoe/gofakeit"
	"github.com/kubeflow/kubeflow/v2/apis/core/v1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"
	"strings"
)

var _ = ginkgo.Describe("Reconciler", func() {
	ginkgo.It("Should create a namespace", func() {
		profile := &v1.Profile{
			ObjectMeta: metav1.ObjectMeta{
				Name: strings.ToLower(gofakeit.FirstName()),
			},
			Spec: v1.ProfileSpec{
				Owner: rbacv1.Subject{
					Kind: "User",
					Name: gofakeit.Email(),
				},
			},
		}
		gomega.Expect(k8s.Create(ctx, profile)).Should(gomega.Succeed())
		gomega.Eventually(komega.Get(profile)).Should(gomega.Succeed())
		namespace := &corev1.Namespace{}
		namespace.SetName(profile.Name)
		gomega.Eventually(komega.Get(namespace)).Should(gomega.Succeed())
	})
})
