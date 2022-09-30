package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type NotebookCullerOptions struct {
	Interval string `json:"interval,omitempty"`
	Duration string `json:"duration,omitempty"`
}

// +kubebuilder:object:root=true

type NotebookCuller struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Options NotebookCullerOptions `json:"options,omitempty"`
}

// +kubebuilder:object:root=true

type NotebookCullerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []NotebookCuller `json:"items,omitempty"`
}
