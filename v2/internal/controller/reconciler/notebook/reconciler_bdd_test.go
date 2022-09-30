/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package notebook

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	corev1beta1 "github.com/kubeflow/kubeflow/v2/apis/core/v1"
)

var _ = Describe("Notebook controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		Name      = "test-notebook"
		Namespace = "default"
		timeout   = time.Second * 10
		interval  = time.Millisecond * 250
	)

	Context("When validating the notebook controller", func() {
		It("Should create replicas", func() {
			By("By creating a new Notebook")
			ctx := context.Background()
			notebook := &corev1beta1.Notebook{
				ObjectMeta: metav1.ObjectMeta{
					Name:      Name,
					Namespace: Namespace,
				},
				Spec: corev1beta1.NotebookSpec{
					Template: corev1beta1.NotebookTemplateSpec{
						Spec: v1.PodSpec{Containers: []v1.Container{{
							Name:  "busybox",
							Image: "busybox",
						}}}},
				}}
			Expect(k8sClient.Create(ctx, notebook)).Should(Succeed())

			notebookLookupKey := types.NamespacedName{Name: Name, Namespace: Namespace}
			createdNotebook := &corev1beta1.Notebook{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, notebookLookupKey, createdNotebook)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			/*
				Checking for the underlying statefulset.
				The satefulset controllers aren't running within envtest, when env test's aren't pointing to the live cluster.
				Only the API server is running within envtest. So cannot check actual pods / replicas.
			*/
			By("By checking that the Notebook has statefulset")
			Eventually(func() (bool, error) {
				sts := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{
					Name:      Name,
					Namespace: Namespace,
				}}
				err := k8sClient.Get(ctx, notebookLookupKey, sts)
				if err != nil {
					return false, err
				}
				return true, nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})
