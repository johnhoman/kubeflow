package culler

import (
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/google/uuid"
	"github.com/kubeflow/kubeflow/v2/apis/core/v1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = ginkgo.Describe("Reconciler", func() {
	ginkgo.It("should cull the notebook", func() {
		ns := &corev1.Namespace{}
		ns.SetName("namespace-" + uuid.New().String()[:8])
		gomega.Expect(k8s.Create(ctx, ns)).Should(gomega.Succeed())
		gomega.Eventually(komega.Get(ns)).Should(gomega.Succeed())

		notebook := &v1.Notebook{
			ObjectMeta: metav1.ObjectMeta{
				Name:      gofakeit.DomainName(),
				Namespace: ns.GetName(),
			},
			Spec: v1.NotebookSpec{
				Template: v1.NotebookTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  "http-server",
							Image: "python:3.9",
						}},
					},
				},
			},
		}
		gomega.Expect(k8s.Create(ctx, notebook)).Should(gomega.Succeed())
		gomega.Eventually(komega.Get(notebook)).Should(gomega.Succeed())

		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      notebook.Name,
				Namespace: notebook.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(notebook, appsv1.SchemeGroupVersion.WithKind("StatefulSet")),
				},
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: pointer.Int32(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"notebook-name": notebook.Name},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"notebook-name": notebook.Name,
						},
					},
					Spec: notebook.Spec.Template.Spec,
				},
			},
		}
		gomega.Expect(k8s.Create(ctx, sts)).Should(gomega.Succeed())
		gomega.Eventually(komega.Get(sts)).Should(gomega.Succeed())

		config := &v1.NotebookCuller{
			ObjectMeta: metav1.ObjectMeta{
				Name:      notebook.Name,
				Namespace: notebook.Namespace,
			},
			Options: v1.NotebookCullerOptions{
				Duration: "2s",
				Interval: "0.5s",
			},
		}
		gomega.Expect(k8s.Create(ctx, config)).Should(gomega.Succeed())
		gomega.Eventually(komega.Get(config)).Should(gomega.Succeed())

		sts = &appsv1.StatefulSet{}
		gomega.Eventually(func() map[string]string {
			err := k8s.Get(ctx, client.ObjectKeyFromObject(notebook), sts)
			if err != nil {
				return nil
			}
			return sts.Annotations
		}).WithTimeout(time.Second * 3).Should(gomega.HaveKey(AnnotationStop))
		gomega.Expect(*sts.Spec.Replicas).Should(gomega.Equal(int32(0)))
	})
	ginkgo.It("should not cull the notebook", func() {
		ns := &corev1.Namespace{}
		ns.SetName("namespace-" + uuid.New().String()[:8])
		gomega.Expect(k8s.Create(ctx, ns)).Should(gomega.Succeed())
		gomega.Eventually(komega.Get(ns)).Should(gomega.Succeed())

		notebook := &v1.Notebook{
			ObjectMeta: metav1.ObjectMeta{
				Name:      gofakeit.DomainName(),
				Namespace: ns.GetName(),
			},
			Spec: v1.NotebookSpec{
				Template: v1.NotebookTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  "http-server",
							Image: "python:3.9",
						}},
					},
				},
			},
		}
		gomega.Expect(k8s.Create(ctx, notebook)).Should(gomega.Succeed())
		gomega.Eventually(komega.Get(notebook)).Should(gomega.Succeed())

		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      notebook.Name,
				Namespace: notebook.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(notebook, appsv1.SchemeGroupVersion.WithKind("StatefulSet")),
				},
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: pointer.Int32(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"notebook-name": notebook.Name},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"notebook-name": notebook.Name,
						},
					},
					Spec: notebook.Spec.Template.Spec,
				},
			},
		}
		gomega.Expect(k8s.Create(ctx, sts)).Should(gomega.Succeed())
		gomega.Eventually(komega.Get(sts)).Should(gomega.Succeed())

		config := &v1.NotebookCuller{
			ObjectMeta: metav1.ObjectMeta{
				Name:      notebook.Name,
				Namespace: notebook.Namespace,
			},
			Options: v1.NotebookCullerOptions{
				// The always idle jupyter client has last activity set to now - 1 hour,
				// so the duration must be at least an hour to not cull it
				Duration: "1h1m",
				Interval: "5m",
			},
		}
		gomega.Expect(k8s.Create(ctx, config)).Should(gomega.Succeed())
		gomega.Eventually(komega.Get(config)).Should(gomega.Succeed())

		sts = &appsv1.StatefulSet{}
		gomega.Consistently(func() map[string]string {
			err := k8s.Get(ctx, client.ObjectKeyFromObject(notebook), sts)
			if err != nil {
				return nil
			}
			return sts.Annotations
		}).WithTimeout(time.Second * 5).ShouldNot(gomega.HaveKey(AnnotationStop))
		gomega.Expect(*sts.Spec.Replicas).Should(gomega.Equal(int32(1)))
	})
})
