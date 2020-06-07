package droidvirt

import (
	"fmt"

	dvv1alpha1 "github.com/lxs137/droidvirt-ctrl/pkg/apis/droidvirt/v1alpha1"
	"github.com/lxs137/droidvirt-ctrl/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileDroidVirt) newVMIForDroidVirt(virt *dvv1alpha1.DroidVirt, dataPVC *corev1.PersistentVolumeClaim) (*kubevirtv1.VirtualMachineInstance, error) {
	bootOrder := uint(1)
	vmi := &kubevirtv1.VirtualMachineInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      virt.Name + "-vmi",
			Namespace: virt.Namespace,
			Labels:    utils.GenVMILabelsForDroidVirt(virt),
			Annotations: map[string]string{
				"hooks.kubevirt.io/hookSidecars":  "[{\"image\": \"registry.cn-shanghai.aliyuncs.com/droidvirt/hook-sidecar:base\"}]",
				"vnc.droidvirt.io/port":           "5900",
				"websocket.vnc.droidvirt.io/port": "5901",
			},
		},
		Spec: kubevirtv1.VirtualMachineInstanceSpec{
			Domain: kubevirtv1.DomainSpec{
				CPU: &kubevirtv1.CPU{
					Model:   "host-model",
					Cores:   1,
					Sockets: 1,
					Threads: 1,
				},
				Resources: kubevirtv1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("2Gi"),
					},
				},
				Devices: kubevirtv1.Devices{
					Disks: []kubevirtv1.Disk{
						{
							Name:      "os-disk",
							BootOrder: &bootOrder,
							Cache:     "none",
							DiskDevice: kubevirtv1.DiskDevice{
								Disk: &kubevirtv1.DiskTarget{
									Bus: "virtio",
								},
							},
						},
						{
							Name:  "data-disk",
							Cache: "none",
							DiskDevice: kubevirtv1.DiskDevice{
								Disk: &kubevirtv1.DiskTarget{
									Bus: "virtio",
								},
							},
						},
					},
					Interfaces: []kubevirtv1.Interface{
						{
							Name:  "default",
							Model: "virtio",
							InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
								Masquerade: &kubevirtv1.InterfaceMasquerade{},
							},
							Ports: []kubevirtv1.Port{
								{
									Name: "ssh",
									Port: 22,
								},
							},
						},
					},
				},
			},
			Volumes: []kubevirtv1.Volume{
				{
					Name: "os-disk",
					VolumeSource: kubevirtv1.VolumeSource{
						Ephemeral: &kubevirtv1.EphemeralVolumeSource{
							PersistentVolumeClaim: &virt.Spec.OSPVCSource,
						},
					},
				},
				{
					Name: "data-disk",
					VolumeSource: kubevirtv1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: dataPVC.Name,
						},
					},
				},
			},
			Networks: []kubevirtv1.Network{
				{
					Name: "default",
					NetworkSource: kubevirtv1.NetworkSource{
						Pod: &kubevirtv1.PodNetwork{
							VMNetworkCIDR: "192.168.100.0/24",
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(virt, vmi, r.scheme); err != nil {
		return nil, err
	}

	return vmi, nil
}

func (r *ReconcileDroidVirt) newServiceForDroidVirt(virt *dvv1alpha1.DroidVirt) (*corev1.Service, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      virt.Name + "-svc",
			Namespace: virt.Namespace,
			Labels:    utils.GenSVCLabelsForDroidVirt(virt),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: utils.GenVMILabelsForDroidVirt(virt),
			Ports: []corev1.ServicePort{
				{
					Name:       "vnc",
					Protocol:   corev1.ProtocolTCP,
					Port:       5900,
					TargetPort: intstr.FromInt(5900),
				},
				{
					Name:       "vnc-ws",
					Protocol:   corev1.ProtocolTCP,
					Port:       5901,
					TargetPort: intstr.FromInt(5901),
				},
				{
					Name:       "ssh",
					Protocol:   corev1.ProtocolTCP,
					Port:       22,
					TargetPort: intstr.FromInt(22),
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(virt, service, r.scheme); err != nil {
		return nil, err
	}

	return service, nil
}

func (r *ReconcileDroidVirt) newGatewayMappingForService(gatewaySpec *dvv1alpha1.VirtGateway, virt *dvv1alpha1.DroidVirt, service *corev1.Service) (*unstructured.Unstructured, error) {
	u := &unstructured.Unstructured{}
	u.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      service.Name + "-mapping",
			"namespace": service.Namespace,
		},
		"spec": map[string]interface{}{
			"prefix":        gatewaySpec.VncWebsocketUrl,
			"service":       fmt.Sprintf("%s.%s.svc.cluster.local:5901", service.Name, service.Namespace),
			"use_websocket": true,
		},
	}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "getambassador.io",
		Kind:    "Mapping",
		Version: "v1",
	})

	if err := controllerutil.SetControllerReference(virt, u, r.scheme); err != nil {
		return nil, err
	}

	return u, nil
}
