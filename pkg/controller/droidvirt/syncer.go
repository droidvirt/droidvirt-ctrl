package droidvirt

import (
	droidvirtv1alpha1 "github.com/lxs137/droidvirt-ctrl/pkg/apis/droidvirt/v1alpha1"
	"github.com/lxs137/droidvirt-ctrl/pkg/syncer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewProbeDeploySyncer(dv *droidvirtv1alpha1.DroidVirt, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	vmi := newVMIForDroidVirt(dv)
	metaobj := &kubevirtv1.VirtualMachineInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vmi.ObjectMeta.Name,
			Namespace: vmi.ObjectMeta.Namespace,
		},
	}
	return syncer.NewObjectSyncer("kubevirt.io/v1alpha3/virtualmachineinstances", vmi, metaobj, c, scheme, func(existing runtime.Object) error {
		return nil
	})
}
