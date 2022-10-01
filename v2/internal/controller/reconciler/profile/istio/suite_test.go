/*
Copyright 2022.

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

package istio

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/kubeflow/kubeflow/v2/apis/core/v1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg     *rest.Config
	k8s     client.Client
	testEnv *envtest.Environment
	ctx     context.Context
	cancel  context.CancelFunc
)

func TestAPIs(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Controller Suite")
}

var _ = ginkgo.BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(ginkgo.GinkgoWriter), zap.UseDevMode(true)))

	ginkgo.By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("../../../../../config/crd/bases"),
			filepath.Join("../../../../../config/crd/thirdparty/istio.io"),
		},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(cfg).NotTo(gomega.BeNil())

	gomega.Expect(v1.AddToScheme(scheme.Scheme)).NotTo(gomega.HaveOccurred())
	gomega.Expect(securityv1beta1.AddToScheme(scheme.Scheme)).Should(gomega.Succeed())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 scheme.Scheme,
		LeaderElection:         false,
		MetricsBindAddress:     "0",
		HealthProbeBindAddress: "0",
	})
	gomega.Expect(err).Should(gomega.BeNil())
	gomega.Expect(mgr).ShouldNot(gomega.BeNil())

	gomega.Expect(Setup(mgr, controller.Options{})).Should(gomega.Succeed())

	ctx, cancel = context.WithCancel(context.Background())

	//+kubebuilder:scaffold:scheme

	k8s, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(k8s).NotTo(gomega.BeNil())

	komega.SetContext(ctx)
	komega.SetClient(k8s)

	go func() {
		defer ginkgo.GinkgoRecover()
		gomega.Expect(mgr.Start(ctx)).Should(gomega.Succeed())
	}()

})

var _ = ginkgo.AfterSuite(func() {
	ginkgo.By("tearing down the test environment")
	cancel()
	err := testEnv.Stop()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
})
