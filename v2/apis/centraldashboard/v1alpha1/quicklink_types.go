package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type QuickLinkType string

type QuickLinkSpec struct {
	Link string `json:"link"`
	Text string `json:"text"`
	Desc string `json:"desc"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

type QuickLink struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec QuickLinkSpec `json:"spec"`
}

// +kubebuilder:object:root=true

type QuickLinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []QuickLink `json:"items,omitempty"`
}
