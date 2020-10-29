package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DroidVirtVolumeSpec defines the desired state of DroidVirtVolume
type DroidVirtVolumeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Device          DeviceInfo                       `json:"device"`
	ClaimSpec       corev1.PersistentVolumeClaimSpec `json:"claimSpec"`
	RBDClone        *CloneProvisionInfo              `json:"rbdClone,omitempty"`
	CloudInitVMSpec `json:",inline"`
}

type DeviceInfo struct {
	Phone string `json:"phone"`
	// wifi_info.json
	// hongmo_device_info.json
	Wifi         string `json:"wifi,omitempty"`
	HongmoDevice string `json:"hongmoDevice,omitempty"`
}

type CloneProvisionInfo struct {
	SrcImage  string            `json:"srcImage"`
	SrcPool   string            `json:"srcPool,omitempty"`
	DestPool  string            `json:"destPool,omitempty"`
	ImageMeta map[string]string `json:"imageMeta,omitempty"`
}

type CloudInitVMSpec struct {
	InitForEmptyDisk *InitForEmptyDiskVMSpec `json:"initForEmptyDisk,omitempty"`
	MigrateFromVMDK  *MigrateFromVMDKVMSpec  `json:"migrateFromVMDK,omitempty"`
}

type InitForEmptyDiskVMSpec struct {
	UserData string `json:"userData,omitempty"`
}

type MigrateFromVMDKVMSpec struct {
	UserData        string                  `json:"userData,omitempty"`
	MigrationSource kubevirtv1.VolumeSource `json:"migrationSource"`
	NodeSelector    map[string]string       `json:"nodeSelector,omitempty"`
}

// DroidVirtVolumeStatus defines the observed state of DroidVirtVolume
type DroidVirtVolumeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	RelatedPVC string               `json:"relatedPVC,omitempty"`
	Phase      DroidVirtVolumePhase `json:"phase,omitempty"`
	Logs       []StatusLog          `json:"logs,omitempty"`
}

type DroidVirtVolumePhase string

const (
	// waiting PVC to be ready
	VolumePending DroidVirtVolumePhase = "Pending"
	// related PVC is bounded
	VolumePVCBounded DroidVirtVolumePhase = "PVCBounded"
	// PVC is not created successfully
	VolumePVCFailed DroidVirtVolumePhase = "PVCFailed"
	// cloud-init do some initializing job, such as: mkpart, mkfs
	VolumeInitializing DroidVirtVolumePhase = "Initializing"
	// cloud-int VM may failed
	VolumeInitFailed DroidVirtVolumePhase = "InitFailed"
	// is ready to be used
	VolumeReady DroidVirtVolumePhase = "Ready"
)

type StatusLog struct {
	Time    metav1.Time `json:"time,omitempty"`
	Message string      `json:"message,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DroidVirtVolume is the Schema for the droidvirtvolumes API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=droidvirtvolumes,scope=Namespaced
type DroidVirtVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DroidVirtVolumeSpec   `json:"spec,omitempty"`
	Status DroidVirtVolumeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DroidVirtVolumeList contains a list of DroidVirtVolume
type DroidVirtVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DroidVirtVolume `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DroidVirtVolume{}, &DroidVirtVolumeList{})
}
