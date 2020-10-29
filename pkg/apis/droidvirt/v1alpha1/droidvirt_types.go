package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DroidVirtSpec defines the desired state of DroidVirt
// +k8s:openapi-gen=true
type DroidVirtSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	VM               *VMSpec                                  `json:"vm,omitempty"`
	OSPVCSource      corev1.PersistentVolumeClaimVolumeSource `json:"osPVC"`
	DataVolumeSource DroidVirtVolumeSelector                  `json:"dataVolume"`
}

type VMSpec struct {
	// Resources describes the Compute Resources required by this vmi.
	Resources kubevirtv1.ResourceRequirements `json:"resources,omitempty"`
	// CPU allow specified the detailed CPU topology inside the vmi.
	// +optional
	CPU *kubevirtv1.CPU `json:"cpu,omitempty"`
	// Memory allow specifying the VMI memory features.
	// +optional
	Memory *kubevirtv1.Memory `json:"memory,omitempty"`
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
	Phase DroidVirtPhase `json:"phase,omitempty"`

	DataPVC string `json:"dataPVC,omitempty"`

	Gateway *VirtGateway `json:"gateway,omitempty"`

	RelatedPod string                `json:"relatedPod,omitempty"`
	Interfaces []VMINetworkInterface `json:"interfaces,omitempty"`

	Logs []StatusLog `json:"logs,omitempty"`
}

type VMINetworkInterface struct {
	IP  string `json:"ipAddress,omitempty"`
	MAC string `json:"mac,omitempty"`
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
