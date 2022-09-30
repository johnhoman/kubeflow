package controller

import (
	"github.com/kubeflow/kubeflow/v2/internal/controller/reconciler/notebook"
	"github.com/kubeflow/kubeflow/v2/internal/controller/reconciler/notebook/culler"
	"github.com/kubeflow/kubeflow/v2/internal/feature"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type Options struct {
	controller.Options
	*feature.Flags
}

func Setup(mgr ctrl.Manager, o Options) error {

	if o.Enabled(feature.NotebookController) {
		if err := notebook.Setup(mgr, o.Options); err != nil {
			return err
		}
	}
	if o.Enabled(feature.NotebookController) && o.Enabled(feature.NotebookCuller) {
		if err := culler.Setup(mgr, o.Options); err != nil {
			return err
		}
	}

	return nil
}
