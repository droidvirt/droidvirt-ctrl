package droidvirtvolume

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	dvv1alpha1 "github.com/lxs137/droidvirt-ctrl/pkg/apis/droidvirt/v1alpha1"
	"golang.org/x/crypto/ssh"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

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
		return r.client.Update(context.TODO(), volume)
	} else {
		return nil
	}
}

func (r *ReconcileDroidVirtVolume) appendLog(volume *dvv1alpha1.DroidVirtVolume, message string) {
	log := dvv1alpha1.StatusLog{
		Time:       metav1.Now(),
		Message:    message,
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

func waitCloudInitReady(sshPort uint32, sshHost, sshUser, sshPassword string) error {
	config := &ssh.ClientConfig{
		Timeout:         time.Second * 5,
		User:            sshUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.Password(sshPassword)},
	}

	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshHost, sshPort), config)
	if err != nil {
		return fmt.Errorf("create SSH client error: %v", err)
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("create SSH session error: %v", err)
	}
	defer session.Close()

	combo, err := session.CombinedOutput("cloud-init status")
	if err != nil {
		return fmt.Errorf("execute SSH command error: %v", err)
	}

	if strings.Contains(string(combo), "done") {
		return nil
	} else {
		return fmt.Errorf("cloud-init not ready, output: %s", string(combo))
	}
}
