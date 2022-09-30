package v1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (in *NotebookCuller) Default() {
	if in.Options.Duration == "" {
		// default is 24 hours
		in.Options.Duration = "1440m"
	}
	if in.Options.Interval == "" {
		// check every minute
		in.Options.Interval = "1m"
	}
}

func (in *NotebookCuller) ValidateCreate() error {
	if in.Name != "default" {
		return apierrors.NewInvalid(
			GroupVersion.WithKind(NotebookCullerKind).GroupKind(),
			in.Name,
			field.ErrorList{field.Invalid(
				field.NewPath("metadata").Child("name"),
				in.Name,
				"metadata.name must be 'default'",
			)},
		)
	}
	return nil
}

func (in *NotebookCuller) ValidateUpdate(old runtime.Object) error { return nil }
func (in *NotebookCuller) ValidateDelete() error                   { return nil }

var (
	_ admission.Validator = &NotebookCuller{}
	_ admission.Defaulter = &NotebookCuller{}
)
