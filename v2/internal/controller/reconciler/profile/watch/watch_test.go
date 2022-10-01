package watch

import (
	v1 "github.com/kubeflow/kubeflow/v2/apis/core/v1"
	"github.com/onsi/gomega"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestEnqueueRequestForProfiles_add(t *testing.T) {

	g := gomega.NewWithT(t)
	g.Expect(v1.AddToScheme(scheme.Scheme)).Should(gomega.Succeed())

	profiles := []client.Object{
		&v1.Profile{
			ObjectMeta: metav1.ObjectMeta{
				Name: "starlord",
			},
			Spec: v1.ProfileSpec{
				Owner: rbacv1.Subject{
					Name: "starlord@guardians.com",
					Kind: "User",
				},
			},
		},
		&v1.Profile{
			ObjectMeta: metav1.ObjectMeta{
				Name: "drax",
			},
			Spec: v1.ProfileSpec{
				Owner: rbacv1.Subject{
					Name: "drax@guardians.com",
					Kind: "User",
				},
			},
		},
	}

	reader := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(profiles...).
		Build().(client.Reader)

	q := &Q{}
	handler := &EnqueueRequestForProfiles{Reader: reader}
	handler.add(q)
	for _, item := range profiles {
		req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(item)}
		g.Expect(*q).Should(gomega.ContainElement(req))
	}
}

type Q []ctrl.Request

func (q *Q) Add(v interface{}) {
	req, ok := v.(ctrl.Request)
	if !ok {
		return
	}
	*q = append(*q, req)
}
