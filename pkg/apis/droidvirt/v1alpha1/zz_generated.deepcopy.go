// +build !ignore_autogenerated

// Code generated by operator-sdk. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloneProvisionInfo) DeepCopyInto(out *CloneProvisionInfo) {
	*out = *in
	if in.ImageMeta != nil {
		in, out := &in.ImageMeta, &out.ImageMeta
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloneProvisionInfo.
func (in *CloneProvisionInfo) DeepCopy() *CloneProvisionInfo {
	if in == nil {
		return nil
	}
	out := new(CloneProvisionInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudInitVMSpec) DeepCopyInto(out *CloudInitVMSpec) {
	*out = *in
	if in.InitForEmptyDisk != nil {
		in, out := &in.InitForEmptyDisk, &out.InitForEmptyDisk
		*out = new(InitForEmptyDiskVMSpec)
		**out = **in
	}
	if in.MigrateFromVMDK != nil {
		in, out := &in.MigrateFromVMDK, &out.MigrateFromVMDK
		*out = new(MigrateFromVMDKVMSpec)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudInitVMSpec.
func (in *CloudInitVMSpec) DeepCopy() *CloudInitVMSpec {
	if in == nil {
		return nil
	}
	out := new(CloudInitVMSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeviceInfo) DeepCopyInto(out *DeviceInfo) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeviceInfo.
func (in *DeviceInfo) DeepCopy() *DeviceInfo {
	if in == nil {
		return nil
	}
	out := new(DeviceInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DroidVirt) DeepCopyInto(out *DroidVirt) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DroidVirt.
func (in *DroidVirt) DeepCopy() *DroidVirt {
	if in == nil {
		return nil
	}
	out := new(DroidVirt)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DroidVirt) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DroidVirtList) DeepCopyInto(out *DroidVirtList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DroidVirt, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DroidVirtList.
func (in *DroidVirtList) DeepCopy() *DroidVirtList {
	if in == nil {
		return nil
	}
	out := new(DroidVirtList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DroidVirtList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DroidVirtSpec) DeepCopyInto(out *DroidVirtSpec) {
	*out = *in
	out.OSPVCSource = in.OSPVCSource
	out.DataVolumeSource = in.DataVolumeSource
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DroidVirtSpec.
func (in *DroidVirtSpec) DeepCopy() *DroidVirtSpec {
	if in == nil {
		return nil
	}
	out := new(DroidVirtSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DroidVirtStatus) DeepCopyInto(out *DroidVirtStatus) {
	*out = *in
	if in.Logs != nil {
		in, out := &in.Logs, &out.Logs
		*out = make([]StatusLog, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Gateway != nil {
		in, out := &in.Gateway, &out.Gateway
		*out = new(VirtGateway)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DroidVirtStatus.
func (in *DroidVirtStatus) DeepCopy() *DroidVirtStatus {
	if in == nil {
		return nil
	}
	out := new(DroidVirtStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DroidVirtVolume) DeepCopyInto(out *DroidVirtVolume) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DroidVirtVolume.
func (in *DroidVirtVolume) DeepCopy() *DroidVirtVolume {
	if in == nil {
		return nil
	}
	out := new(DroidVirtVolume)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DroidVirtVolume) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DroidVirtVolumeList) DeepCopyInto(out *DroidVirtVolumeList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DroidVirtVolume, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DroidVirtVolumeList.
func (in *DroidVirtVolumeList) DeepCopy() *DroidVirtVolumeList {
	if in == nil {
		return nil
	}
	out := new(DroidVirtVolumeList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DroidVirtVolumeList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DroidVirtVolumeSelector) DeepCopyInto(out *DroidVirtVolumeSelector) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DroidVirtVolumeSelector.
func (in *DroidVirtVolumeSelector) DeepCopy() *DroidVirtVolumeSelector {
	if in == nil {
		return nil
	}
	out := new(DroidVirtVolumeSelector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DroidVirtVolumeSpec) DeepCopyInto(out *DroidVirtVolumeSpec) {
	*out = *in
	out.Device = in.Device
	in.ClaimSpec.DeepCopyInto(&out.ClaimSpec)
	if in.RBDClone != nil {
		in, out := &in.RBDClone, &out.RBDClone
		*out = new(CloneProvisionInfo)
		(*in).DeepCopyInto(*out)
	}
	in.CloudInitVMSpec.DeepCopyInto(&out.CloudInitVMSpec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DroidVirtVolumeSpec.
func (in *DroidVirtVolumeSpec) DeepCopy() *DroidVirtVolumeSpec {
	if in == nil {
		return nil
	}
	out := new(DroidVirtVolumeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DroidVirtVolumeStatus) DeepCopyInto(out *DroidVirtVolumeStatus) {
	*out = *in
	if in.Logs != nil {
		in, out := &in.Logs, &out.Logs
		*out = make([]StatusLog, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DroidVirtVolumeStatus.
func (in *DroidVirtVolumeStatus) DeepCopy() *DroidVirtVolumeStatus {
	if in == nil {
		return nil
	}
	out := new(DroidVirtVolumeStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InitForEmptyDiskVMSpec) DeepCopyInto(out *InitForEmptyDiskVMSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InitForEmptyDiskVMSpec.
func (in *InitForEmptyDiskVMSpec) DeepCopy() *InitForEmptyDiskVMSpec {
	if in == nil {
		return nil
	}
	out := new(InitForEmptyDiskVMSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrateFromVMDKVMSpec) DeepCopyInto(out *MigrateFromVMDKVMSpec) {
	*out = *in
	in.MigrationSource.DeepCopyInto(&out.MigrationSource)
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrateFromVMDKVMSpec.
func (in *MigrateFromVMDKVMSpec) DeepCopy() *MigrateFromVMDKVMSpec {
	if in == nil {
		return nil
	}
	out := new(MigrateFromVMDKVMSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StatusLog) DeepCopyInto(out *StatusLog) {
	*out = *in
	in.Time.DeepCopyInto(&out.Time)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StatusLog.
func (in *StatusLog) DeepCopy() *StatusLog {
	if in == nil {
		return nil
	}
	out := new(StatusLog)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VirtGateway) DeepCopyInto(out *VirtGateway) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VirtGateway.
func (in *VirtGateway) DeepCopy() *VirtGateway {
	if in == nil {
		return nil
	}
	out := new(VirtGateway)
	in.DeepCopyInto(out)
	return out
}
