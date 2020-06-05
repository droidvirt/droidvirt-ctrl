package droidvirt

import (
	droidvirtv1alpha1 "github.com/lxs137/droidvirt-ctrl/pkg/apis/droidvirt/v1alpha1"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
)

func newVMIForDroidVirt(spec *droidvirtv1alpha1.DroidVirt) *kubevirtv1.VirtualMachineInstance {
	return nil
}
