package droidvirtvolume

import (
	"context"
	"fmt"
	"github.com/lxs137/droidvirt-ctrl/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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

var needRequeueIn30s = reconcile.Result{
	Requeue:      true,
	RequeueAfter: time.Second * 30,
}

var needRequeueIn5s = reconcile.Result{
	Requeue:      true,
	RequeueAfter: time.Second * 5,
}

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
	err = c.Watch(&source.Kind{Type: &dvv1alpha1.DroidVirtVolume{}}, &handler.EnqueueRequestForObject{}, predicate.GenerationChangedPredicate{})
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

type RelatedResource struct {
	VirtVolume *dvv1alpha1.DroidVirtVolume
	PVC        *corev1.PersistentVolumeClaim
	VMI        *kubevirtv1.VirtualMachineInstance
	CloudInit  CloudInitStatus
}

func (r *ReconcileDroidVirtVolume) FetchRelatedResource(request reconcile.Request) (*RelatedResource, error) {
	res := &RelatedResource{}

	virt := &dvv1alpha1.DroidVirtVolume{}
	err := r.client.Get(context.TODO(), request.NamespacedName, virt)
	if errors.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	res.VirtVolume = virt

	pvc, err := r.relatedPVC(res.VirtVolume)
	if err != nil {
		return res, err
	}
	res.PVC = pvc

	k, v := utils.GenPVCLabelToMarkDroidVirtVolumeReady()
	if res.PVC != nil && res.PVC.Labels != nil && res.PVC.Labels[k] == v {
		res.CloudInit = CloudInitDone
		return res, nil
	}

	vmi, err := r.relatedVMI(res.VirtVolume)
	if err != nil {
		return res, err
	}
	res.VMI = vmi

	if res.VMI != nil && (res.VMI.Status.Phase == kubevirtv1.Running && len(res.VMI.Status.Interfaces) > 0) {
		vmiIP := res.VMI.Status.Interfaces[0].IP
		status, err := cloudInitStatus(22, vmiIP, config.CloudInitVMISSHUser, config.CloudInitVMISSHPassword)
		if err != nil {
			log.Error(err, "Get cloud-init status err")
			return res, err
		}
		res.CloudInit = status
	}

	return res, nil
}

type ReconcileAction struct {
	Tag      string
	Expected func(*RelatedResource) bool
	Action   func(*RelatedResource, *ReconcileDroidVirtVolume) (reconcile.Result, error)
}

var actions = []ReconcileAction{
	{
		Tag: "PVC committed",
		Expected: func(res *RelatedResource) bool {
			return res.PVC == nil
		},
		Action: func(res *RelatedResource, r *ReconcileDroidVirtVolume) (reconcile.Result, error) {
			claim, err := r.newPVCForDroidVirtVolume(res.VirtVolume)
			log.Info("PVC Generated", "PVC", claim)
			if err != nil {
				log.Error(err, "Generate PVC error")
				return needRequeueIn5s, err
			}

			if err := utils.CreateIfNotExistsPVC(r.client, r.scheme, claim); err != nil {
				log.Error(err, "Create PVC error")
				return needRequeueIn30s, err
			}
			log.Info("PVC created")

			res.VirtVolume.Status.Phase = dvv1alpha1.VolumePending
			res.VirtVolume.Status.RelatedPVC = claim.Name
			r.appendLog(res.VirtVolume, fmt.Sprintf("PVC is created: %s/%s", claim.Namespace, claim.Name))
			_ = r.syncStatus(res.VirtVolume)
			return reconcile.Result{}, nil
		},
	},
	{
		Tag: "PVC lost",
		Expected: func(res *RelatedResource) bool {
			return res.PVC != nil && res.PVC.Status.Phase == corev1.ClaimLost
		},
		Action: func(res *RelatedResource, r *ReconcileDroidVirtVolume) (reconcile.Result, error) {
			res.VirtVolume.Status.Phase = dvv1alpha1.VolumePVCFailed
			r.appendLog(res.VirtVolume, fmt.Sprintf("PVC lost: %s/%s", res.PVC.Namespace, res.PVC.Name))
			_ = r.syncStatus(res.VirtVolume)
			return reconcile.Result{}, nil
		},
	},
	{
		Tag: "PVC bounded, create VMI",
		Expected: func(res *RelatedResource) bool {
			return res.PVC != nil && res.PVC.Status.Phase == corev1.ClaimBound &&
				res.CloudInit != CloudInitDone &&
				res.VMI == nil
		},
		Action: func(res *RelatedResource, r *ReconcileDroidVirtVolume) (reconcile.Result, error) {
			vmi, err := r.newVMIForDroidVirtVolume(res.VirtVolume, res.PVC.Name)
			log.Info("VMI Generated", "VMI", vmi)
			if err != nil {
				log.Error(err, "Generate VMI error")
				return needRequeueIn5s, err
			}

			if err := utils.CreateIfNotExistsVMI(r.client, r.scheme, vmi); err != nil {
				log.Error(err, "Create VMI error")
				return needRequeueIn30s, err
			}
			log.Info("VMI created")

			res.VirtVolume.Status.Phase = dvv1alpha1.VolumePVCBounded
			r.appendLog(res.VirtVolume, fmt.Sprintf("VMI has created"))
			_ = r.syncStatus(res.VirtVolume)
			return reconcile.Result{}, nil
		},
	},
	{
		Tag: "VMI pending",
		Expected: func(res *RelatedResource) bool {
			return res.PVC != nil && res.PVC.Status.Phase == corev1.ClaimBound &&
				res.CloudInit != CloudInitDone &&
				res.VMI != nil &&
				(res.VMI.Status.Phase == kubevirtv1.VmPhaseUnset ||
					res.VMI.Status.Phase == kubevirtv1.Pending ||
					res.VMI.Status.Phase == kubevirtv1.Scheduling ||
					res.VMI.Status.Phase == kubevirtv1.Scheduled)
		},
		Action: func(res *RelatedResource, r *ReconcileDroidVirtVolume) (reconcile.Result, error) {
			workerPod, err := r.relatedPod(res.VMI)
			if err != nil || workerPod == nil || workerPod.GetDeletionTimestamp() != nil {
				return needRequeueIn5s, err
			}
			if r.isPodBlocking(workerPod) {
				err := utils.DeleteVMI(r.client, res.VMI)
				if err != nil {
					return needRequeueIn5s, err
				}
				r.appendLog(res.VirtVolume, fmt.Sprintf("VMI's pod is blocking, delete failed VMI: %s/%s (%s)", res.VMI.Namespace, res.VMI.Name, res.VMI.UID))
				_ = r.syncStatus(res.VirtVolume)
				return reconcile.Result{}, nil
			} else {
				r.appendLog(res.VirtVolume, fmt.Sprintf("Waiting VMI to be ready, worker pod: %s/%s, current VMI phase: %s", workerPod.Namespace, workerPod.Name, res.VMI.Status.Phase))
				_ = r.syncStatus(res.VirtVolume)
				return needRequeueIn30s, nil
			}
		},
	},
	{
		Tag: "VMI failed",
		Expected: func(res *RelatedResource) bool {
			return res.PVC != nil && res.PVC.Status.Phase == corev1.ClaimBound &&
				res.CloudInit != CloudInitDone &&
				res.VMI != nil &&
				(res.VMI.Status.Phase == kubevirtv1.Succeeded ||
					res.VMI.Status.Phase == kubevirtv1.Failed)
		},
		Action: func(res *RelatedResource, r *ReconcileDroidVirtVolume) (reconcile.Result, error) {
			err := utils.DeleteVMI(r.client, res.VMI)
			if err != nil {
				return needRequeueIn5s, err
			}
			res.VirtVolume.Status.Phase = dvv1alpha1.VolumeInitFailed
			r.appendLog(res.VirtVolume, "VMI failed, delete VMI")
			_ = r.syncStatus(res.VirtVolume)
			return reconcile.Result{}, nil
		},
	},
	{
		Tag: "Cloud-init job running",
		Expected: func(res *RelatedResource) bool {
			return res.PVC != nil && res.PVC.Status.Phase == corev1.ClaimBound &&
				(res.CloudInit == CloudInitUnknown || res.CloudInit == CloudInitRunning) &&
				res.VMI != nil && res.VMI.Status.Phase == kubevirtv1.Running
		},
		Action: func(res *RelatedResource, r *ReconcileDroidVirtVolume) (reconcile.Result, error) {
			res.VirtVolume.Status.Phase = dvv1alpha1.VolumeInitializing
			r.appendLog(res.VirtVolume, fmt.Sprintf("Waiting cloud-init job done, worker pod ip: %s", res.VMI.Status.Interfaces[0].IP))
			_ = r.syncStatus(res.VirtVolume)
			return needRequeueIn30s, nil
		},
	},
	{
		Tag: "Cloud-init job failed",
		Expected: func(res *RelatedResource) bool {
			return res.PVC != nil && res.PVC.Status.Phase == corev1.ClaimBound &&
				res.CloudInit == CloudInitFailed &&
				res.VMI != nil
		},
		Action: func(res *RelatedResource, r *ReconcileDroidVirtVolume) (reconcile.Result, error) {
			err := utils.DeleteVMI(r.client, res.VMI)
			if err != nil {
				return needRequeueIn5s, err
			}
			res.VirtVolume.Status.Phase = dvv1alpha1.VolumeInitFailed
			r.appendLog(res.VirtVolume, "cloud-init job failed, delete VMI")
			_ = r.syncStatus(res.VirtVolume)
			return reconcile.Result{}, nil
		},
	},
	{
		Tag: "Cloud-init job done",
		Expected: func(res *RelatedResource) bool {
			return res.PVC != nil && res.PVC.Status.Phase == corev1.ClaimBound &&
				res.CloudInit == CloudInitDone
		},
		Action: func(res *RelatedResource, r *ReconcileDroidVirtVolume) (reconcile.Result, error) {
			if err := r.markPVCReady(res.PVC); err != nil {
				return needRequeueIn5s, err
			}
			res.VirtVolume.Status.Phase = dvv1alpha1.VolumeReady
			r.appendLog(res.VirtVolume, "cloud-init job done")
			_ = r.syncStatus(res.VirtVolume)
			return reconcile.Result{}, nil
		},
	},
	{
		Tag: "Clean ready VMI",
		Expected: func(res *RelatedResource) bool {
			return res.PVC != nil && res.PVC.Status.Phase == corev1.ClaimBound &&
				res.CloudInit == CloudInitDone &&
				res.VMI != nil
		},
		Action: func(res *RelatedResource, r *ReconcileDroidVirtVolume) (reconcile.Result, error) {
			err := utils.DeleteVMI(r.client, res.VMI)
			if err != nil {
				return needRequeueIn5s, err
			}
			r.appendLog(res.VirtVolume, "clean ready VMI")
			_ = r.syncStatus(res.VirtVolume)
			return reconcile.Result{}, nil
		},
	},
}

// Reconcile reads that state of the cluster for a DroidVirtVolume object and makes changes based on the state read
// and what is in the DroidVirtVolume.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileDroidVirtVolume) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	res, err := r.FetchRelatedResource(request)
	if err != nil || res == nil {
		return needRequeueIn5s, err
	}

	reqLogger.Info("Reconciling DroidVirtVolume", "Phase", res.VirtVolume.Status.Phase)

	var processedTags []string = nil
	combineResult := reconcile.Result{}
	for _, action := range actions {
		if action.Expected(res) {
			processedTags = append(processedTags, action.Tag)
			result, err := action.Action(res, r)
			if err != nil {
				reqLogger.Error(err, "ReconcileAction error", "ActionTag", action.Tag)
			} else {
				reqLogger.Info("ReconcileAction success", "ActionTag", action.Tag)
			}
			combineResult.Requeue = combineResult.Requeue || result.Requeue
			if combineResult.RequeueAfter == 0 || combineResult.RequeueAfter > result.RequeueAfter {
				combineResult.RequeueAfter = result.RequeueAfter
			}
		}
	}
	reqLogger.Info("ReconcileAction", "ProcessedTags", processedTags)

	return combineResult, nil
}
