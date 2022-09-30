package controller

import (
	"github.com/kubeflow/kubeflow/v2/internal/controller/reconciler/notebook"
	"github.com/kubeflow/kubeflow/v2/internal/controller/reconciler/notebook/culler"
	"github.com/kubeflow/kubeflow/v2/internal/controller/reconciler/profile"
	profileistio "github.com/kubeflow/kubeflow/v2/internal/controller/reconciler/profile/istio"
	reflector "github.com/kubeflow/kubeflow/v2/internal/controller/reconciler/profile/reflect"
	"github.com/kubeflow/kubeflow/v2/internal/feature"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type Options struct {
	controller.Options
	*feature.Flags
}

func Setup(mgr ctrl.Manager, o Options) error {

	setupFuncs := []func(mgr ctrl.Manager, o controller.Options) error{
		notebook.Setup,
		profile.Setup,
	}

	for _, fn := range setupFuncs {
		if err := fn(mgr, o.Options); err != nil {
			return err
		}
	}

	if o.Enabled(feature.NotebookCulling) {
		if err := culler.Setup(mgr, o.Options); err != nil {
			return err
		}
	}
	if o.Enabled(feature.Istio) {
		if err := profileistio.Setup(mgr, o.Options); err != nil {
			return err
		}
	}
	if o.Enabled(feature.Reflection) {
		if err := reflector.Setup(mgr, o.Options); err != nil {
			return err
		}
	}

	return nil
}
