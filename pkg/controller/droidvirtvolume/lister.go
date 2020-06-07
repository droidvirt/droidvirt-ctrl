package droidvirtvolume

import (
	"context"
	"github.com/lxs137/droidvirt-ctrl/pkg/utils"

	dvv1alpha1 "github.com/lxs137/droidvirt-ctrl/pkg/apis/droidvirt/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ReconcileDroidVirtVolume) relatedVMI(volume *dvv1alpha1.DroidVirtVolume) (*kubevirtv1.VirtualMachineInstance, error) {
	vmiLabels := utils.GenVMILabelsForDroidVirtVolume(volume)
	labelSelector := labels.NewSelector()
	for k, v := range vmiLabels {
		require, err := labels.NewRequirement(k, selection.Equals, []string{v})
		if err != nil {
			continue
		} else {
			labelSelector.Add(*require)
		}
	}
	listOpts := &client.ListOptions{
		Namespace:     volume.Namespace,
		LabelSelector: labelSelector,
	}

	vmis := &kubevirtv1.VirtualMachineInstanceList{}
	err := r.client.List(context.TODO(), listOpts, vmis)
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
	labelSelector := labels.NewSelector()
	for k, v := range claimLabels {
		require, err := labels.NewRequirement(k, selection.Equals, []string{v})
		if err != nil {
			continue
		} else {
			labelSelector.Add(*require)
		}
	}
	listOpts := &client.ListOptions{
		Namespace:     volume.Namespace,
		LabelSelector: labelSelector,
	}

	claims := &corev1.PersistentVolumeClaimList{}
	err := r.client.List(context.TODO(), listOpts, claims)
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
