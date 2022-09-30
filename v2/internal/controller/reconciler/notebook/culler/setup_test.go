package culler

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/kubeflow/kubeflow/v2/apis/core/v1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

func TestAPI(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "controller-tests")
}

var (
	ctx      context.Context
	k8s      client.Client
	teardown func() error
)

var _ = ginkgo.BeforeSuite(func() {
	testEnv := &envtest.Environment{
		ErrorIfCRDPathMissing: true,
		CRDDirectoryPaths: []string{
			filepath.Join("../../../../../config/crd/bases"),
		},
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())

	gomega.Expect(v1.AddToScheme(scheme.Scheme))

	cfg, err := testEnv.Start()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 scheme.Scheme,
		LeaderElection:         false,
		HealthProbeBindAddress: "0",
		MetricsBindAddress:     "0",
	})
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(mgr).ShouldNot(gomega.BeNil())
	e := Setup(mgr, controller.Options{}, WithJupyterClient(&alwaysIdle{}))
	gomega.Expect(e).Should(gomega.BeNil())

	k8s, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	komega.SetContext(ctx)
	komega.SetClient(k8s)

	go func() {
		defer ginkgo.GinkgoRecover()
		gomega.Expect(mgr.Start(ctx)).Should(gomega.Succeed())
	}()

	teardown = func() error {
		cancel()
		return testEnv.Stop()
	}
})

var _ = ginkgo.AfterSuite(func() {
	gomega.Expect(teardown()).Should(gomega.Succeed())
})
