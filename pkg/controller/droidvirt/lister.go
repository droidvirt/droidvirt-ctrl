package droidvirt

import (
	"context"
	"fmt"
	dvv1alpha1 "github.com/lxs137/droidvirt-ctrl/pkg/apis/droidvirt/v1alpha1"
	"github.com/lxs137/droidvirt-ctrl/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ReconcileDroidVirt) findDataVolume(virt *dvv1alpha1.DroidVirt) (*dvv1alpha1.DroidVirtVolume, error) {
	volumeSource := virt.Spec.DataVolumeSource
	listOpts := &client.ListOptions{
		Namespace:     virt.Namespace,
		FieldSelector: fields.OneTermEqualSelector("spec.device.phone", volumeSource.Phone),
	}

	dvvs := &dvv1alpha1.DroidVirtVolumeList{}
	err := r.client.List(context.TODO(), dvvs, listOpts)
	if err != nil {
		return nil, err
	}

	var dvv *dvv1alpha1.DroidVirtVolume = nil
	for _, item := range dvvs.Items {
		if item.Status.Phase == dvv1alpha1.VolumeReady {
			if dvv == nil {
				dvv = &item
			} else if dvv.CreationTimestamp.Before(&item.CreationTimestamp) {
				dvv = &item
			}
		}
	}
	return dvv, nil
}

func (r *ReconcileDroidVirt) findReadyDataVolumePVC(virt *dvv1alpha1.DroidVirt) (*corev1.PersistentVolumeClaim, error) {
	dvv, err := r.findDataVolume(virt)
	if dvv == nil || err != nil {
		return nil, fmt.Errorf("can not find ready DroidVirtVolume, err: %v", err)
	}

	claimLabels := utils.GenPVCLabelsForDroidVirtVolume(dvv)
	listOpts := &client.ListOptions{
		Namespace:     dvv.Namespace,
		LabelSelector: labels.SelectorFromSet(claimLabels),
	}

	claims := &corev1.PersistentVolumeClaimList{}
	err = r.client.List(context.TODO(), claims, listOpts)
	if err != nil {
		return nil, err
	}

	var relatedPVC *corev1.PersistentVolumeClaim = nil
	for _, item := range claims.Items {
		if !metav1.IsControlledBy(&item, dvv) || item.Status.Phase != corev1.ClaimBound {
			continue
		}
		if relatedPVC == nil {
			relatedPVC = &item
		} else if relatedPVC.CreationTimestamp.Before(&item.CreationTimestamp) {
			relatedPVC = &item
		}
	}
	return relatedPVC, nil
}

func (r *ReconcileDroidVirt) relatedVMI(virt *dvv1alpha1.DroidVirt) (*kubevirtv1.VirtualMachineInstance, error) {
	vmiLabels := utils.GenVMILabelsForDroidVirt(virt)
	listOpts := &client.ListOptions{
		Namespace:     virt.Namespace,
		LabelSelector: labels.SelectorFromSet(vmiLabels),
	}

	vmis := &kubevirtv1.VirtualMachineInstanceList{}
	err := r.client.List(context.TODO(), vmis, listOpts)
	if err != nil {
		return nil, err
	}

	var relatedVMI *kubevirtv1.VirtualMachineInstance = nil
	for _, item := range vmis.Items {
		if !metav1.IsControlledBy(&item, virt) {
			continue
		}
		if relatedVMI == nil {
			relatedVMI = &item
		} else if relatedVMI.CreationTimestamp.Before(&item.CreationTimestamp) {
			relatedVMI = &item
		}
	}
	return relatedVMI, nil
}

func (r *ReconcileDroidVirt) relatedPod(vmi *kubevirtv1.VirtualMachineInstance) (*corev1.Pod, error) {
	podSelector := utils.GenPodLabelsForVMI(vmi)
	listOpts := &client.ListOptions{
		Namespace:     vmi.Namespace,
		LabelSelector: labels.SelectorFromSet(podSelector),
	}

	pods := &corev1.PodList{}
	err := r.client.List(context.TODO(), pods, listOpts)
	if err != nil {
		return nil, err
	}

	var relatedPod *corev1.Pod = nil
	for _, item := range pods.Items {
		if !metav1.IsControlledBy(&item, vmi) {
			continue
		}
		if relatedPod == nil {
			relatedPod = &item
		} else if relatedPod.CreationTimestamp.Before(&item.CreationTimestamp) {
			relatedPod = &item
		}
	}
	return relatedPod, nil
}

func (r *ReconcileDroidVirt) relatedSVC(virt *dvv1alpha1.DroidVirt) (*corev1.Service, error) {
	svcLabels := utils.GenSVCLabelsForDroidVirt(virt)
	listOpts := &client.ListOptions{
		Namespace:     virt.Namespace,
		LabelSelector: labels.SelectorFromSet(svcLabels),
	}

	svcs := &corev1.ServiceList{}
	err := r.client.List(context.TODO(), svcs, listOpts)
	if err != nil {
		return nil, err
	}

	var relatedSVC *corev1.Service = nil
	for _, item := range svcs.Items {
		if !metav1.IsControlledBy(&item, virt) {
			continue
		}
		if relatedSVC == nil {
			relatedSVC = &item
		} else if relatedSVC.CreationTimestamp.Before(&item.CreationTimestamp) {
			relatedSVC = &item
		}
	}
	return relatedSVC, nil
}

func (r *ReconcileDroidVirt) relatedMapping(virt *dvv1alpha1.DroidVirt) (*unstructured.Unstructured, error) {
	svcLabels := utils.GenMappingLabelsForDroidVirt(virt)
	listOpts := &client.ListOptions{
		Namespace:     virt.Namespace,
		LabelSelector: labels.SelectorFromSet(svcLabels),
	}


	mappings := &unstructured.UnstructuredList{}
	mappings.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "getambassador.io",
		Kind:    "Mapping",
		Version: "v1",
	})
	err := r.client.List(context.TODO(), mappings, listOpts)
	if err != nil {
		return nil, err
	}

	var relatedMapping *unstructured.Unstructured = nil
	for _, item := range mappings.Items {
		if !metav1.IsControlledBy(&item, virt) {
			continue
		}
		if relatedMapping == nil {
			relatedMapping = &item
		}
	}
	return relatedMapping, nil
}
