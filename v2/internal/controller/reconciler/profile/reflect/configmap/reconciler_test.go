package configmap

import (
	"fmt"
	"github.com/kubeflow/kubeflow/v2/apis/core/v1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
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

		kf := &corev1.Namespace{}
		kf.SetName("kubeflow")
		gomega.Expect(k8s.Create(ctx, kf)).Should(gomega.Succeed())
		gomega.Eventually(komega.Get(kf)).Should(gomega.Succeed())

		configMaps := []*corev1.ConfigMap{{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pip-config",
				Namespace: "kubeflow",
				Labels:    map[string]string{labelReflect: "true"},
			},
			Data: map[string]string{
				"pip.conf": fmt.Sprintf("[install]\nuser = true"),
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop-config",
				Namespace: "kubeflow",
			},
			Data: map[string]string{
				"hadoop.conf": `
<property>
  <name>fs.s3a.access.key</name>
  <description>AWS access key ID.
   Omit for IAM role-based or provider-based authentication.</description>
</property>

<property>
  <name>fs.s3a.secret.key</name>
  <description>AWS secret key.
   Omit for IAM role-based or provider-based authentication.</description>
</property>
`,
			},
		}}
		for _, configMap := range configMaps {
			gomega.Expect(k8s.Create(ctx, configMap)).Should(gomega.Succeed())
			gomega.Eventually(komega.Get(configMap)).Should(gomega.Succeed())
		}

		for _, item := range []string{"starlord", "drax", "gamora"} {
			cm := configMaps[0].DeepCopy()
			cm.SetNamespace(item)
			gomega.Eventually(komega.Get(cm)).Should(gomega.Succeed())
			gomega.Expect(cm.Data).Should(gomega.Equal(configMaps[0].Data))
			gomega.Expect(cm.Immutable).Should(gomega.Equal(configMaps[0].Immutable))
		}
		for _, item := range []string{"starlord", "drax", "gamora"} {
			cm := configMaps[1].DeepCopy()
			cm.SetNamespace(item)
			gomega.Consistently(komega.Get(cm)).ShouldNot(gomega.Succeed())
		}
	})
})
