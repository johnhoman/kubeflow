package configmap

import (
	"github.com/kubeflow/kubeflow/v2/apis/centraldashboard/v1alpha1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = ginkgo.Describe("Reconciler", func() {

	ginkgo.It("creates configmaps in all profiles", func() {
		kf := &corev1.Namespace{}
		kf.SetName("kubeflow")
		gomega.Expect(k8s.Create(ctx, kf)).Should(gomega.Succeed())
		gomega.Eventually(komega.Get(kf)).Should(gomega.Succeed())

		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configmapName,
				Namespace: "kubeflow",
			},
			Data: map[string]string{},
		}
		gomega.Expect(k8s.Create(ctx, configMap)).Should(gomega.Succeed())
		gomega.Eventually(komega.Get(configMap)).Should(gomega.Succeed())

		links := []*v1alpha1.MenuLink{{
			ObjectMeta: metav1.ObjectMeta{
				Name: "notebooks",
			},
			Spec: v1alpha1.MenuLinkSpec{
				Type: "item",
				Link: "/jupyter/",
				Text: "Notebooks",
				Icon: "book",
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name: "kfserving",
			},
			Spec: v1alpha1.MenuLinkSpec{
				Type: "item",
				Link: "/models/",
				Text: "Models",
				Icon: "kubeflow:models",
			},
		}}

		for _, item := range links {
			gomega.Expect(k8s.Create(ctx, item)).Should(gomega.Succeed())
			gomega.Eventually(komega.Get(item)).Should(gomega.Succeed())
		}

		gomega.Eventually(func() map[string]string {
			err := k8s.Get(ctx, client.ObjectKeyFromObject(configMap), configMap)
			if err != nil {
				return nil
			}
			return configMap.Data
		}).Should(gomega.HaveKey("links"))
	})
})
