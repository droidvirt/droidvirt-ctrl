package droidvirt

import (
	"context"
	"reflect"

	dvv1alpha1 "github.com/lxs137/droidvirt-ctrl/pkg/apis/droidvirt/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileDroidVirt) syncStatus(virt *dvv1alpha1.DroidVirt) error {
	oldVirt := &dvv1alpha1.DroidVirt{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      virt.Name,
		Namespace: virt.Namespace,
	}, oldVirt)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(oldVirt.Status, virt.Status) {
		return r.client.Status().Update(context.TODO(), virt)
	} else {
		return nil
	}
}

func (r *ReconcileDroidVirt) appendLog(virt *dvv1alpha1.DroidVirt, message string) {
	log := dvv1alpha1.StatusLog{
		Time:    metav1.Now(),
		Message: message,
	}

	if virt.Status.Logs == nil {
		virt.Status.Logs = []dvv1alpha1.StatusLog{log}
	} else {
		virt.Status.Logs = append(virt.Status.Logs, log)
	}
}

func (r *ReconcileDroidVirt) appendLogAndSync(virt *dvv1alpha1.DroidVirt, message string) error {
	r.appendLog(virt, message)
	return r.syncStatus(virt)
}
