// Package v1 contains API Schema definitions for the kubeflow.org v1 API group
// +kubebuilder:object:generate=true
// +groupName=kubeflow.org
package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

const (
	Group   = "kubeflow.org"
	Version = "v1"
)

var (
	GroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	SchemeBuilder = scheme.Builder{GroupVersion: GroupVersion}

	AddToScheme = SchemeBuilder.AddToScheme

	NotebookCullerKind = reflect.TypeOf(&NotebookCuller{}).Elem().Name()
	NotebookKind       = reflect.TypeOf(&Notebook{}).Elem().Name()
)

func init() {
	SchemeBuilder.Register(
		&Notebook{},
		&NotebookList{},
		&NotebookCuller{},
		&NotebookCullerList{},
	)
}
