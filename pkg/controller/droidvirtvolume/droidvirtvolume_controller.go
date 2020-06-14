package droidvirtvolume

import (
	"context"
	"fmt"
	"github.com/lxs137/droidvirt-ctrl/pkg/utils"
	"time"

	dvv1alpha1 "github.com/lxs137/droidvirt-ctrl/pkg/apis/droidvirt/v1alpha1"
	"github.com/lxs137/droidvirt-ctrl/pkg/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("droidvirtvolume")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new DroidVirtVolume Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDroidVirtVolume{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("droidvirtvolume-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource DroidVirtVolume
	err = c.Watch(&source.Kind{Type: &dvv1alpha1.DroidVirtVolume{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	watchResources := []runtime.Object{
		&kubevirtv1.VirtualMachineInstance{},
		&corev1.PersistentVolumeClaim{},
	}

	for _, subResource := range watchResources {
		err = c.Watch(&source.Kind{Type: subResource}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &dvv1alpha1.DroidVirtVolume{},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileDroidVirtVolume implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileDroidVirtVolume{}

// ReconcileDroidVirtVolume reconciles a DroidVirtVolume object
type ReconcileDroidVirtVolume struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a DroidVirtVolume object and makes changes based on the state read
// and what is in the DroidVirtVolume.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileDroidVirtVolume) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	// Fetch the DroidVirtVolume instance
	virtVolume := &dvv1alpha1.DroidVirtVolume{}
	err := r.client.Get(context.TODO(), request.NamespacedName, virtVolume)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	reqLogger.Info("Reconciling DroidVirtVolume", "DroidVirtVolume.Spec", virtVolume.Spec, "DroidVirtVolume.Status", virtVolume.Status)
	switch virtVolume.Status.Phase {
	case "":
		// Generate PVC spec
		claim, err := r.newPVCForDroidVirtVolume(virtVolume)
		reqLogger.Info("Generate PVC", "PVC", claim)
		if err != nil {
			_ = r.appendLogAndSync(virtVolume, fmt.Sprintf("generate PVC spec error: %v", err))
			return reconcile.Result{}, fmt.Errorf("generate PVC spec error: %v", err)
		}

		// Create PVC if not found
		if err := utils.CreateIfNotExistsPVC(r.client, r.scheme, claim); err != nil {
			_ = r.appendLogAndSync(virtVolume, fmt.Sprintf("creating PVC error, retry after 30s: %v", err))
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: time.Second * 30,
			}, err
		}
		reqLogger.Info("PVC created")

		virtVolume.Status.Phase = dvv1alpha1.VolumePending
		virtVolume.Status.RelatedPVC = claim.Name
		r.appendLog(virtVolume, fmt.Sprintf("PVC is created: %s/%s", claim.Namespace, claim.Name))
		_ = r.syncStatus(virtVolume)
		return reconcile.Result{}, nil
	case dvv1alpha1.VolumePending:
		claim, err := r.relatedPVC(virtVolume)
		if err != nil || claim == nil {
			reqLogger.Error(err, "Related PVC not found")
			return reconcile.Result{}, err
		}

		switch claim.Status.Phase {
		case corev1.ClaimPending:
			reqLogger.Info("Related PVC is pending, waiting...")
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: time.Second * 30,
			}, nil
		case corev1.ClaimBound:
			virtVolume.Status.Phase = dvv1alpha1.VolumePVCBounded
			r.appendLog(virtVolume, fmt.Sprintf("PVC is bounded: %v", claim.ObjectMeta))
			_ = r.syncStatus(virtVolume)
			return reconcile.Result{}, nil
		case corev1.ClaimLost:
			virtVolume.Status.Phase = dvv1alpha1.VolumePVCFailed
			r.appendLog(virtVolume, fmt.Sprintf("PVC is lost: %v", claim.ObjectMeta))
			_ = r.syncStatus(virtVolume)
			return reconcile.Result{}, fmt.Errorf("related PVC is lost")
		}
	case dvv1alpha1.VolumePVCBounded:
		claim, err := r.relatedPVC(virtVolume)
		if claim == nil {
			reqLogger.Error(err, "Related PVC not found")
			return reconcile.Result{}, err
		}

		vmi, err := r.newVMIForDroidVirtVolume(virtVolume, claim.Name)
		reqLogger.Info("Generate VMI", "VMI", vmi)
		if err != nil {
			_ = r.appendLogAndSync(virtVolume, fmt.Sprintf("generate VMI spec error: %v", err))
			return reconcile.Result{}, fmt.Errorf("generate VMI spec error: %v", err)
		}

		// Create VMI if not found
		if err := utils.CreateIfNotExistsVMI(r.client, r.scheme, vmi); err != nil {
			_ = r.appendLogAndSync(virtVolume, fmt.Sprintf("creating VMI error, retry after 30s: %v", err))
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: time.Second * 30,
			}, err
		}
		reqLogger.Info("VMI created")

		virtVolume.Status.Phase = dvv1alpha1.VolumeInitializing
		r.appendLog(virtVolume, fmt.Sprintf("VMI is created: %s/%s", vmi.Name, vmi.Namespace))
		_ = r.syncStatus(virtVolume)
		return reconcile.Result{}, nil
	case dvv1alpha1.VolumeInitializing:
		vmi, err := r.relatedVMI(virtVolume)
		if err != nil || vmi == nil {
			reqLogger.Error(err, "Related VMI not found")
			return reconcile.Result{}, err
		}

		switch vmi.Status.Phase {
		case kubevirtv1.Pending, kubevirtv1.Scheduled, kubevirtv1.Scheduling:
			reqLogger.Info("Related VMI is pending, waiting...")
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: time.Second * 30,
			}, nil
		case kubevirtv1.Running:
			reqLogger.Info("Related VMI is running, waiting cloud-init job done...")
			if len(vmi.Status.Interfaces) > 0 {
				vmiIP := vmi.Status.Interfaces[0].IP
				reqLogger.Info("cloud-init VMI's info", "ip", vmiIP)
				status, err := cloudInitStatus(22, vmiIP, config.CloudInitVMISSHUser, config.CloudInitVMISSHPassword)
				switch status {
				case CloudInitUnknown, CloudInitRunning:
					reqLogger.Error(err, "cloud-init SSH check failed, retry after 30s")
					return reconcile.Result{
						Requeue:      true,
						RequeueAfter: time.Second * 30,
					}, err
				case CloudInitDone:
					virtVolume.Status.Phase = dvv1alpha1.VolumeReady
					r.appendLog(virtVolume, fmt.Sprintf("VMI is running, and cloud-init is ready: %v", vmi.ObjectMeta))
					_ = r.syncStatus(virtVolume)
					return reconcile.Result{}, nil
				case CloudInitFailed:
					virtVolume.Status.Phase = dvv1alpha1.VolumeInitFailed
					r.appendLog(virtVolume, fmt.Sprintf("cloud-init error output: %v", err))
					_ = r.syncStatus(virtVolume)
					return reconcile.Result{}, err
				}
			}
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: time.Second * 30,
			}, nil
		case kubevirtv1.Succeeded, kubevirtv1.Failed, kubevirtv1.Unknown:
			virtVolume.Status.Phase = dvv1alpha1.VolumeInitFailed
			r.appendLog(virtVolume, fmt.Sprintf("VMI is %s: %v", vmi.Status.Phase, vmi.ObjectMeta))
			_ = r.syncStatus(virtVolume)
			return reconcile.Result{}, fmt.Errorf("VMI has some trouble")
		}
	case dvv1alpha1.VolumeReady, dvv1alpha1.VolumeInitFailed:
		claim, err := r.relatedPVC(virtVolume)
		if claim == nil {
			reqLogger.Error(err, "Related Claim has lost, recreate it")
			virtVolume.Status.Phase = ""
			virtVolume.Status.RelatedPVC = ""
			_ = r.syncStatus(virtVolume)
			return reconcile.Result{}, nil
		}

		vmi, _ := r.relatedVMI(virtVolume)
		if vmi != nil {
			reqLogger.Info("Cloud-init VMI's job is complete, delete it", "VMI.Spec", vmi.Spec, "VMI.status", vmi.Status)
			//return reconcile.Result{}, nil
			return reconcile.Result{}, r.client.Delete(context.TODO(), vmi)
		}

	}

	return reconcile.Result{}, nil
}
