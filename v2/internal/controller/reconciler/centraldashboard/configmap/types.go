package configmap

import "github.com/kubeflow/kubeflow/v2/apis/centraldashboard/v1alpha1"

type Links struct {
	MenuLinks          []v1alpha1.MenuLinkSpec          `json:"menuLinks,omitempty"`
	QuickLinks         []v1alpha1.QuickLinkSpec         `json:"quickLinks,omitempty"`
	DocumentationItems []v1alpha1.DocumentationItemSpec `json:"documentationItems,omitempty"`
}
