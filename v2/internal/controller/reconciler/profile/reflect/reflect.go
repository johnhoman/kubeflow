package reflect

import (
	"github.com/kubeflow/kubeflow/v2/internal/controller/reconciler/profile/reflect/configmap"
	"github.com/kubeflow/kubeflow/v2/internal/controller/reconciler/profile/reflect/secret"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	setupFuncs := []func(mgr ctrl.Manager, o controller.Options) error{
		configmap.Setup,
		secret.Setup,
	}

	for _, fn := range setupFuncs {
		if err := fn(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
