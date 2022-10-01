package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type MenuLinkType string

type MenuLinkSpec struct {
	Type MenuLinkType `json:"type"`
	Link string       `json:"link"`
	Text string       `json:"text"`
	Icon string       `json:"icon"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="TYPE",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="LINK",type="string",JSONPath=".spec.link"
// +kubebuilder:printcolumn:name="TEXT",type="string",JSONPath=".spec.text"

type MenuLink struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MenuLinkSpec `json:"spec"`
}

// +kubebuilder:object:root=true

type MenuLinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []MenuLink `json:"items,omitempty"`
}
