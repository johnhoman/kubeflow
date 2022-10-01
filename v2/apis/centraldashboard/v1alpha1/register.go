// Package v1 contains API Schema definitions for the kubeflow.org v1 API group
// +kubebuilder:object:generate=true
// +groupName=centraldashboard.kubeflow.org

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

const (
	Group   = "centraldashboard.kubeflow.org"
	Version = "v1alpha1"
)

var (
	GroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	SchemeBuilder = scheme.Builder{GroupVersion: GroupVersion}

	AddToScheme = SchemeBuilder.AddToScheme
)

func init() {
	SchemeBuilder.Register(
		&DocumentationItem{},
		&DocumentationItemList{},
		&MenuLink{},
		&MenuLinkList{},
		&QuickLink{},
		&QuickLinkList{},
	)
}
