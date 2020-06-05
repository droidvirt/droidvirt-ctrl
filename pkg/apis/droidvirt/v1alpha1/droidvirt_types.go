package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DroidVirtSpec defines the desired state of DroidVirt
// +k8s:openapi-gen=true
type DroidVirtSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Phone   string                    `json:"string"`
	Apps    map[string]AndroidAppSpec `json:"apps,omitempty"`
	Ingress VirtIngressSpec           `json:"ingress,omitempty"`
}

type VirtIngressSpec struct {
	VncWebsocket string `json:"vncWebsocket,omitempty"`
	SSH          string `json:"ssh,omitempty"`
}

type AndroidAppSpec struct {
	PackageID string            `json:"string"`
	Info      map[string]string `json:"info,omitempty"`
}

// DroidVirtStatus defines the observed state of DroidVirt
// +k8s:openapi-gen=true
type DroidVirtStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Apps map[string]AndroidAppStatus `json:"apps,omitempty"`
}

type AndroidAppStatus struct {
	UpdateTime *metav1.Time      `json:"updateTime,omitempty"`
	Info       map[string]string `json:"info,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DroidVirt is the Schema for the droidvirts API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type DroidVirt struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DroidVirtSpec   `json:"spec,omitempty"`
	Status DroidVirtStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DroidVirtList contains a list of DroidVirt
type DroidVirtList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DroidVirt `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DroidVirt{}, &DroidVirtList{})
}
