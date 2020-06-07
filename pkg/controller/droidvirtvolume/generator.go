package droidvirtvolume

import (
	"encoding/json"
	"fmt"
	"strings"

	dvv1alpha1 "github.com/lxs137/droidvirt-ctrl/pkg/apis/droidvirt/v1alpha1"
	"github.com/lxs137/droidvirt-ctrl/pkg/config"
	"github.com/lxs137/droidvirt-ctrl/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	CloudInitUserDataTpl = `Content-Type: multipart/mixed; boundary="//"
MIME-Version: 1.0

--//
Content-Type: text/cloud-config; charset="utf-8"
MIME-Version: 1.0
Content-Transfer-Encoding: 7bit
Content-Disposition: attachment; filename="cloud-config"

CLOUD_CONFIG

--//
Content-Type: text/x-shellscript; charset="utf-8"
MIME-Version: 1.0
Content-Transfer-Encoding: 7bit
Content-Disposition: attachment; filename="shell-script"

SHELL_SCRIPT

--//
`
	CloudInitCloudConfigTpl = `#cloud-config
users:
- default
- name: SSH_USER
  plain_text_passwd: 'SSH_PASSWORD'
  shell: /bin/bash
  lock_passwd: False
  sudo: "ALL=(ALL) NOPASSWD:ALL"
cloud_config_modules:
- [ runcmd, always ]
cloud_final_modules:
- [ scripts-user, always ]
MOUNTS_CONFIG
`
	InitForEmptyDiskShellTpl = `#!/bin/bash
set -e

dhclient -v -cf /etc/dhcp/dhclient.conf enp1s0

cp_set_owner () {
  local src_file=$1
  local dest_file=$2
  local file_owner=$3

  rm -rf "$dest_file"
  cp -r -p "$src_file" "$dest_file"
  chown -R "$file_owner" "$dest_file"
}

node /opt/device_info/main.js PHONE_NO /opt/device_info/hongmo_device_info.json /opt/device_info/wifi_info.json

cp_set_owner "/opt/device_info/hongmo_device_info.json" "/mnt/dest/media/0/tencent/hongmo_device_info.json" "1023:1023"

cp_set_owner "/opt/device_info/wifi_info.json" "/mnt/dest/media/0/tencent/wifi_info.json" "1023:1023"

`
	MigrateFromVMDKShellTpl = `#!/bin/bash
set -e

dhclient -v -cf /etc/dhcp/dhclient.conf enp1s0

cp_set_owner () {
  local src_file=$1
  local dest_file=$2
  local file_owner=$3

  rm -rf "$dest_file"
  cp -r -p "$src_file" "$dest_file"
  chown -R "$file_owner" "$dest_file"
}

cp_set_owner "/mnt/src/android-6.0-r3/data/data/com.tencent.mm" "/mnt/dest/app/com.tencent.mm" "10042:10042"
chown -h 0:0 /mnt/dest/app/com.tencent.mm/lib

cp_set_owner "/mnt/src/android-6.0-r3/data/media/0/tencent/wifi_info.json" "/mnt/dest/media/0/tencent/wifi_info.json" "1023:1023"

cp_set_owner "/mnt/src/android-6.0-r3/data/media/0/tencent/hongmo_device_info.json" "/mnt/dest/media/0/tencent/hongmo_device_info.json" "1023:1023"
`
)

func (r *ReconcileDroidVirtVolume) newPVCForDroidVirtVolume(virtVolume *dvv1alpha1.DroidVirtVolume) (*corev1.PersistentVolumeClaim, error) {
	annotations := make(map[string]string)
	if virtVolume.Spec.RBDClone != nil {
		cloneInfoBytes, err := json.Marshal(virtVolume.Spec.RBDClone)
		if err != nil {
			return nil, err
		} else {
			annotations[config.RBDClonedInfoAnnotation] = string(cloneInfoBytes)
		}
	}

	claim := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:        virtVolume.Name + "-propagation-pvc",
			Namespace:   virtVolume.Namespace,
			Labels:      utils.GenPVCLabelsForDroidVirtVolume(virtVolume),
			Annotations: annotations,
		},
		Spec: *virtVolume.Spec.ClaimSpec.DeepCopy(),
	}

	if err := controllerutil.SetControllerReference(virtVolume, claim, r.scheme); err != nil {
		return nil, err
	}

	return claim, nil
}

func (r *ReconcileDroidVirtVolume) newVMIForDroidVirtVolume(virtVolume *dvv1alpha1.DroidVirtVolume, claim string) (*kubevirtv1.VirtualMachineInstance, error) {
	volumes := []kubevirtv1.Volume{
		{
			Name: "os-disk",
			VolumeSource: kubevirtv1.VolumeSource{
				Ephemeral: &kubevirtv1.EphemeralVolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: config.CloudInitOSPVCName,
					},
				},
			},
		},
		{
			Name: "dest-disk",
			VolumeSource: kubevirtv1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: claim,
				},
			},
		},
	}
	bootOrder := uint(1)
	vmiDisks := []kubevirtv1.Disk{
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
			Name:   "dest-disk",
			Cache:  "none",
			Serial: config.CloudInitVMIDestDiskSerial,
			DiskDevice: kubevirtv1.DiskDevice{
				Disk: &kubevirtv1.DiskTarget{
					Bus: "virtio",
				},
			},
		},
	}

	if virtVolume.Spec.InitForEmptyDisk != nil {
		userData := virtVolume.Spec.InitForEmptyDisk.UserData
		if userData == "" {
			diskMountInfo := map[string]string{
				config.CloudInitVMIDestDiskSerial: "/mnt/dest",
			}
			cloudConfig := genCloudInitMountsConfig(diskMountInfo)
			shellScript := strings.Replace(InitForEmptyDiskShellTpl, "PHONE_NO", virtVolume.Spec.Device.Phone, 1)
			userData = mergeCloudInitUserData(cloudConfig, shellScript)
		}

		volumes = append(volumes, kubevirtv1.Volume{
			Name: "cloudinit-disk",
			VolumeSource: kubevirtv1.VolumeSource{
				CloudInitNoCloud: &kubevirtv1.CloudInitNoCloudSource{
					UserData: userData,
				},
			},
		})
		vmiDisks = append(vmiDisks, kubevirtv1.Disk{
			Name:  "cloudinit-disk",
			Cache: "none",
			DiskDevice: kubevirtv1.DiskDevice{
				Disk: &kubevirtv1.DiskTarget{
					Bus: "virtio",
				},
			},
		})
	} else if virtVolume.Spec.MigrateFromVMDK != nil {
		userData := virtVolume.Spec.MigrateFromVMDK.UserData
		if userData == "" {
			diskMountInfo := map[string]string{
				config.CloudInitVMISrcDiskSerial:  "/mnt/src",
				config.CloudInitVMIDestDiskSerial: "/mnt/dest",
			}
			cloudConfig := genCloudInitMountsConfig(diskMountInfo)
			userData = mergeCloudInitUserData(cloudConfig, MigrateFromVMDKShellTpl)
		}

		volumes = append(volumes, kubevirtv1.Volume{
			Name:         "src-disk",
			VolumeSource: virtVolume.Spec.MigrateFromVMDK.MigrationSource,
		}, kubevirtv1.Volume{
			Name: "cloudinit-disk",
			VolumeSource: kubevirtv1.VolumeSource{
				CloudInitNoCloud: &kubevirtv1.CloudInitNoCloudSource{
					UserData: userData,
				},
			},
		})
		vmiDisks = append(vmiDisks, kubevirtv1.Disk{
			Name:   "src-disk",
			Cache:  "none",
			Serial: config.CloudInitVMISrcDiskSerial,
			DiskDevice: kubevirtv1.DiskDevice{
				Disk: &kubevirtv1.DiskTarget{
					Bus: "virtio",
				},
			},
		}, kubevirtv1.Disk{
			Name:  "cloudinit-disk",
			Cache: "none",
			DiskDevice: kubevirtv1.DiskDevice{
				Disk: &kubevirtv1.DiskTarget{
					Bus: "virtio",
				},
			},
		})
	}

	vmi := &kubevirtv1.VirtualMachineInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      virtVolume.Name + "-propagation-vmi",
			Namespace: virtVolume.Namespace,
			Labels:    utils.GenVMILabelsForDroidVirtVolume(virtVolume),
			Annotations: map[string]string{
				"hooks.kubevirt.io/hookSidecars": "[{\"image\": \"registry.cn-shanghai.aliyuncs.com/droidvirt/hook-sidecar:base\"}]",
				"vnc.droidvirt.io/port":          "5900",
				"disk.droidvirt.io/names":        "os-disk,",
				"disk.droidvirt.io/driverType":   "qcow2",
			},
		},
		Spec: kubevirtv1.VirtualMachineInstanceSpec{
			Domain: kubevirtv1.DomainSpec{
				CPU: &kubevirtv1.CPU{
					Model:   "host-model",
					Cores:   1,
					Sockets: 2,
					Threads: 2,
				},
				Resources: kubevirtv1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("2Gi"),
					},
				},
				Devices: kubevirtv1.Devices{
					Disks: vmiDisks,
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
			Volumes: volumes,
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

	if virtVolume.Spec.MigrateFromVMDK != nil && len(virtVolume.Spec.MigrateFromVMDK.NodeSelector) > 0 {
		vmi.Spec.NodeSelector = virtVolume.Spec.MigrateFromVMDK.NodeSelector
	}

	if err := controllerutil.SetControllerReference(virtVolume, vmi, r.scheme); err != nil {
		return nil, err
	}

	return vmi, nil
}

func mergeCloudInitUserData(cloudConfig, shellScript string) string {
	userData := strings.Replace(CloudInitUserDataTpl, "CLOUD_CONFIG", cloudConfig, 1)
	userData = strings.Replace(userData, "SHELL_SCRIPT", shellScript, 1)
	return userData
}

func genCloudInitMountsConfig(serialToMountpoint map[string]string) string {
	mountCmds := []string{}
	cmdTpl := `[ /dev/disk/by-id/virtio-SERIAL-part1, MOUNTPOINT, "ext4", "defaults,nofail,discard", "0", "0" ]`
	for k, v := range serialToMountpoint {
		cmd := strings.Replace(cmdTpl, "SERIAL", k, 1)
		cmd = strings.Replace(cmd, "MOUNTPOINT", v, 1)
		mountCmds = append(mountCmds, cmd)
	}
	mountConfig := "mounts:\n"
	for _, item := range mountCmds {
		mountConfig += fmt.Sprintf("- %s\n", item)
	}

	// set ssh user and password
	cloudConfig := strings.Replace(CloudInitCloudConfigTpl, "SSH_USER", config.CloudInitVMISSHUser, 1)
	cloudConfig = strings.Replace(cloudConfig, "SSH_PASSWORD", config.CloudInitVMISSHPassword, 1)

	return strings.Replace(cloudConfig, "MOUNTS_CONFIG", mountConfig, 1)
}
