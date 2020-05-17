package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SsmConfigSpec defines the desired state of SsmConfig
type SsmConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Env     string   `json:"env"`
	SsmKeys []string `json:"ssmKeys"`
}

// SsmConfigStatus defines the observed state of SsmConfig
type SsmConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SsmConfig is the Schema for the ssmconfigs API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=ssmconfigs,scope=Namespaced
type SsmConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SsmConfigSpec   `json:"spec,omitempty"`
	Status SsmConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SsmConfigList contains a list of SsmConfig
type SsmConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SsmConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SsmConfig{}, &SsmConfigList{})
}
