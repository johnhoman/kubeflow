package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (in *Profile) Default() {}

func (in *Profile) ValidateCreate() error                   { return nil }
func (in *Profile) ValidateUpdate(old runtime.Object) error { return nil }
func (in *Profile) ValidateDelete() error                   { return nil }

var (
	_ admission.Validator = &Profile{}
	_ admission.Defaulter = &Profile{}
)
