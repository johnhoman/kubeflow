package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type DocumentationItemSpec struct {
	Link string `json:"link"`
	Text string `json:"text"`
	Icon string `json:"icon"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

type DocumentationItem struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DocumentationItemSpec `json:"spec"`
}

// +kubebuilder:object:root=true

type DocumentationItemList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []DocumentationItem `json:"items,omitempty"`
}
