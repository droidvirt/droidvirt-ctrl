package droidvirtvolume

import (
	"context"
	"fmt"
	"github.com/droidvirt/droidvirt-ctrl/pkg/utils"
	"reflect"
	"strings"
	"time"

	dvv1alpha1 "github.com/droidvirt/droidvirt-ctrl/pkg/apis/droidvirt/v1alpha1"
	"golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileDroidVirtVolume) markPVCReady(pvc *corev1.PersistentVolumeClaim) error {
	currPVC := &corev1.PersistentVolumeClaim{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      pvc.Name,
		Namespace: pvc.Namespace,
	}, currPVC)
	if err != nil {
		return err
	}

	k, v := utils.GenPVCLabelToMarkDroidVirtVolumeReady()
	if currPVC.Labels[k] == v {
		return nil
	}

	currPVC.Labels[k] = v
	return r.client.Update(context.TODO(), currPVC)
}

func (r *ReconcileDroidVirtVolume) syncStatus(volume *dvv1alpha1.DroidVirtVolume) error {
	oldVolume := &dvv1alpha1.DroidVirtVolume{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      volume.Name,
		Namespace: volume.Namespace,
	}, oldVolume)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(oldVolume.Status, volume.Status) {
		return r.client.Status().Update(context.TODO(), volume)
	} else {
		return nil
	}
}

func (r *ReconcileDroidVirtVolume) appendLog(volume *dvv1alpha1.DroidVirtVolume, message string) {
	log := dvv1alpha1.StatusLog{
		Time:    metav1.Now(),
		Message: message,
	}

	if volume.Status.Logs == nil {
		volume.Status.Logs = []dvv1alpha1.StatusLog{log}
	} else {
		volume.Status.Logs = append(volume.Status.Logs, log)
	}
}

func (r *ReconcileDroidVirtVolume) appendLogAndSync(volume *dvv1alpha1.DroidVirtVolume, message string) error {
	r.appendLog(volume, message)
	return r.syncStatus(volume)
}

type CloudInitStatus string

const (
	CloudInitRunning CloudInitStatus = "running"
	CloudInitDone    CloudInitStatus = "done"
	CloudInitFailed  CloudInitStatus = "failed"
	CloudInitUnknown CloudInitStatus = "unknown"
)

func cloudInitStatus(sshPort uint32, sshHost, sshUser, sshPassword string) (CloudInitStatus, error) {
	config := &ssh.ClientConfig{
		Timeout:         time.Second * 5,
		User:            sshUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.Password(sshPassword)},
	}

	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshHost, sshPort), config)
	if err != nil {
		return CloudInitUnknown, fmt.Errorf("create SSH client error: %v", err)
	}
	defer sshClient.Close()

	statusOutput, _ := executeSSHCmd(sshClient, "cloud-init status")

	if strings.Contains(string(statusOutput), "status: running") {
		return CloudInitRunning, nil
	} else if strings.Contains(string(statusOutput), "status: done") {
		return CloudInitDone, nil
	} else if strings.Contains(string(statusOutput), "status: error") {
		result, _ := executeSSHCmd(sshClient, "cat /var/lib/cloud/data/result.json")
		return CloudInitFailed, fmt.Errorf("cloud-init error, output: %s", string(result))
	} else {
		return CloudInitUnknown, fmt.Errorf("cloud-init unknown output: %s", string(statusOutput))
	}
}

func executeSSHCmd(client *ssh.Client, cmd string) ([]byte, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("create SSH session error: %v", err)
	}
	defer session.Close()
	return session.Output(cmd)
}

func (r *ReconcileDroidVirtVolume) isPodBlocking(po *corev1.Pod) bool {
	if po.Status.Phase != corev1.PodPending {
		return false
	}

	// 300 seconds
	now := metav1.Now().Second()
	if now-po.GetCreationTimestamp().Second() < 300 {
		return false
	}

	for _, item := range po.Status.Conditions {
		if item.Type == corev1.PodScheduled && item.Status == corev1.ConditionFalse {
			return false
		}
	}
	return true
}
