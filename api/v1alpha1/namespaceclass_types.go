// +k8s:deepcopy-gen=package
// +groupName=core.resource-wrapper.io

package v1alpha1

import (
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +kubebuilder:resource:scope=Cluster
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NamespaceClass is the Schema for the namespaceclasses API
type NamespaceClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NamespaceClassSpec   `json:"spec,omitempty"`
	Status NamespaceClassStatus `json:"status,omitempty"`
}

// NamespaceClassSpec defines the desired state of NamespaceClass
type NamespaceClassSpec struct {
	// NetworkPolicyTemplate defines the NetworkPolicy to be created
	NetworkPolicyTemplate *networkingv1.NetworkPolicySpec `json:"networkPolicyTemplate,omitempty"`

	// ServiceAccountTemplate defines the ServiceAccount to be created
	ServiceAccountTemplate *ServiceAccountTemplate `json:"serviceAccountTemplate,omitempty"`
}

// ServiceAccountTemplate defines the configuration for ServiceAccount
type ServiceAccountTemplate struct {

	// AutomountServiceAccountToken indicates whether a service account token should be automatically mounted
	AutomountServiceAccountToken *bool `json:"automountServiceAccountToken,omitempty"`
}

// NamespaceClassStatus defines the observed state of NamespaceClass
type NamespaceClassStatus struct {
	// ObservedGeneration is the generation of the resource that was most recently applied to the cluster
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	LastAppliedTime metav1.Time `json:"lastAppliedTime,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NamespaceClassList contains a list of NamespaceClass
type NamespaceClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NamespaceClass `json:"items"`
}
