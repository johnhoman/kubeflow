package configmap

import (
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

		podDefaults := []*v1.PodDefault{{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pip-config",
				Namespace: "kubeflow",
				Labels:    map[string]string{labelReflect: "true"},
			},
			Spec: v1.PodDefaultSpec{
				Volumes: []corev1.Volume{{
					Name: "pip-config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "pip-config",
							},
						},
					},
				}},
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop-config",
				Namespace: "kubeflow",
			},
			Spec: v1.PodDefaultSpec{
				Volumes: []corev1.Volume{{
					Name: "pip-config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "pip-config",
							},
						},
					},
				}},
			},
		}}
		for _, podDefault := range podDefaults {
			gomega.Expect(k8s.Create(ctx, podDefault)).Should(gomega.Succeed())
			gomega.Eventually(komega.Get(podDefault)).Should(gomega.Succeed())
		}

		for _, item := range []string{"starlord", "drax", "gamora"} {
			pd := podDefaults[0].DeepCopy()
			pd.SetNamespace(item)
			gomega.Eventually(komega.Get(pd)).Should(gomega.Succeed())
			gomega.Expect(pd.Spec).Should(gomega.Equal(podDefaults[0].Spec))
		}
		for _, item := range []string{"starlord", "drax", "gamora"} {
			pd := podDefaults[1].DeepCopy()
			pd.SetNamespace(item)
			gomega.Consistently(komega.Get(pd)).ShouldNot(gomega.Succeed())
		}
	})
})
