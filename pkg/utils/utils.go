package utils

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	dvv1alpha1 "github.com/lxs137/droidvirt-ctrl/pkg/apis/droidvirt/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GenPVCLabelsForDroidVirtVolume(virtVolume *dvv1alpha1.DroidVirtVolume) map[string]string {
	return map[string]string{
		"droidvirtvolume.droidvirt.io/name":  virtVolume.Name,
		"droidvirtvolume.droidvirt.io/phone": virtVolume.Spec.Device.Phone,
	}
}

func GenVMILabelsForDroidVirtVolume(virtVolume *dvv1alpha1.DroidVirtVolume) map[string]string {
	return map[string]string{
		"droidvirtvolume.droidvirt.io/name":  virtVolume.Name,
		"droidvirtvolume.droidvirt.io/phone": virtVolume.Spec.Device.Phone,
	}
}

func GenVMILabelsForDroidVirt(virt *dvv1alpha1.DroidVirt) map[string]string {
	return map[string]string{
		"droidvirt.droidvirt.io/name": virt.Name,
	}
}

func GenSVCLabelsForDroidVirt(virt *dvv1alpha1.DroidVirt) map[string]string {
	return map[string]string{
		"droidvirt.droidvirt.io/name": virt.Name,
	}
}

func GenMappingLabelsForDroidVirt(virt *dvv1alpha1.DroidVirt) map[string]string {
	return map[string]string{
		"droidvirt.droidvirt.io/name": virt.Name,
	}
}

func GenPodLabelsForVMI(vmi *kubevirtv1.VirtualMachineInstance) map[string]string {
	return map[string]string{
		"kubevirt.io/created-by": string(vmi.UID),
	}
}

func CreateIfNotExists(c client.Client, ifExistKey client.ObjectKey, obj runtime.Object, gvk schema.GroupVersionKind) error {
	found := &unstructured.Unstructured{}
	found.SetGroupVersionKind(gvk)
	err := c.Get(context.Background(), ifExistKey, found)
	if err != nil && errors.IsNotFound(err) {
		return c.Create(context.TODO(), obj)
	} else if err != nil {
		return err
	} else {
		return nil
	}
}

func CreateIfNotExistsService(c client.Client, s *runtime.Scheme, service *corev1.Service) error {
	gvk, err := apiutil.GVKForObject(service, s)
	if err != nil || gvk.Empty() {
		return err
	}
	return CreateIfNotExists(c, types.NamespacedName{
		Namespace: service.Namespace,
		Name:      service.Name,
	}, service, gvk)
}

func CreateIfNotExistsPVC(c client.Client, s *runtime.Scheme, claim *corev1.PersistentVolumeClaim) error {
	gvk, err := apiutil.GVKForObject(claim, s)
	if err != nil || gvk.Empty() {
		return err
	}
	return CreateIfNotExists(c, types.NamespacedName{
		Namespace: claim.Namespace,
		Name:      claim.Name,
	}, claim, gvk)
}

func CreateIfNotExistsVMI(c client.Client, s *runtime.Scheme, vmi *kubevirtv1.VirtualMachineInstance) error {
	gvk, err := apiutil.GVKForObject(vmi, s)
	if err != nil || gvk.Empty() {
		return err
	}
	return CreateIfNotExists(c, types.NamespacedName{
		Namespace: vmi.Namespace,
		Name:      vmi.Name,
	}, vmi, gvk)
}

func DeleteVMI(c client.Client, vmi *kubevirtv1.VirtualMachineInstance) error {
	return c.Delete(context.TODO(), vmi)
}

func CreateIfNotExistsMapping(c client.Client, mapping *unstructured.Unstructured) error {
	return CreateIfNotExists(c, types.NamespacedName{
		Namespace: mapping.GetNamespace(),
		Name:      mapping.GetName(),
	}, mapping, schema.GroupVersionKind{
		Group:   "getambassador.io",
		Kind:    "Mapping",
		Version: "v1",
	})
}
