package droidvirtvolume

import (
	"context"
	"github.com/droidvirt/droidvirt-ctrl/pkg/utils"

	dvv1alpha1 "github.com/droidvirt/droidvirt-ctrl/pkg/apis/droidvirt/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ReconcileDroidVirtVolume) relatedVMI(volume *dvv1alpha1.DroidVirtVolume) (*kubevirtv1.VirtualMachineInstance, error) {
	vmiLabels := utils.GenVMILabelsForDroidVirtVolume(volume)
	listOpts := &client.ListOptions{
		Namespace:     volume.Namespace,
		LabelSelector: labels.SelectorFromSet(vmiLabels),
	}

	vmis := &kubevirtv1.VirtualMachineInstanceList{}
	err := r.client.List(context.TODO(), vmis, listOpts)
	if err != nil {
		return nil, err
	}

	var relatedVMI *kubevirtv1.VirtualMachineInstance = nil
	for _, item := range vmis.Items {
		if !metav1.IsControlledBy(&item, volume) {
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

func (r *ReconcileDroidVirtVolume) relatedPVC(volume *dvv1alpha1.DroidVirtVolume) (*corev1.PersistentVolumeClaim, error) {
	claimLabels := utils.GenPVCLabelsForDroidVirtVolume(volume)
	listOpts := &client.ListOptions{
		Namespace:     volume.Namespace,
		LabelSelector: labels.SelectorFromSet(claimLabels),
	}

	claims := &corev1.PersistentVolumeClaimList{}
	err := r.client.List(context.TODO(), claims, listOpts)
	if err != nil {
		return nil, err
	}

	var relatedPVC *corev1.PersistentVolumeClaim = nil
	for _, item := range claims.Items {
		if !metav1.IsControlledBy(&item, volume) {
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

func (r *ReconcileDroidVirtVolume) relatedPod(vmi *kubevirtv1.VirtualMachineInstance) (*corev1.Pod, error) {
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
