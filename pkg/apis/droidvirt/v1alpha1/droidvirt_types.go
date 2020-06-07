package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
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
	OSPVCSource      corev1.PersistentVolumeClaimVolumeSource `json:"osPVC"`
	DataVolumeSource DroidVirtVolumeSelector                  `json:"dataVolume"`
}

type DroidVirtVolumeSelector struct {
	Phone string `json:"phone"`
}

type VirtGateway struct {
	VncWebsocketUrl string `json:"vncWebsocketUrl,omitempty"`
}

type DroidVirtPhase string

const (
	// VMI object has been created
	VirtualMachinePending DroidVirtPhase = "VMPending"
	// VMI's phase is "Running"
	VirtualMachineRunning DroidVirtPhase = "VMRunning"
	// Ambassador's gateway Mapping service is ready
	GatewayServiceReady DroidVirtPhase = "ServiceReady"
	// VMI is shutdown
	VirtualMachineDown DroidVirtPhase = "VMDown"
)

// DroidVirtStatus defines the observed state of DroidVirt
// +k8s:openapi-gen=true
type DroidVirtStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	RelatedVMI string         `json:"relatedVMI,omitempty"`
	Phase      DroidVirtPhase `json:"phase,omitempty"`
	Logs       []StatusLog    `json:"logs,omitempty"`
	Gateway    *VirtGateway   `json:"gateway,omitempty"`
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
