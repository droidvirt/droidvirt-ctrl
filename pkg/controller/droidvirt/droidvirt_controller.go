package droidvirt

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	dvv1alpha1 "github.com/lxs137/droidvirt-ctrl/pkg/apis/droidvirt/v1alpha1"
	"github.com/lxs137/droidvirt-ctrl/pkg/utils"
)

var log = logf.Log.WithName("droidvirt-controller")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new DroidVirt Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDroidVirt{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("droidvirt-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource DroidVirt
	err = c.Watch(&source.Kind{Type: &dvv1alpha1.DroidVirt{}}, &handler.EnqueueRequestForObject{}, predicate.GenerationChangedPredicate{})
	if err != nil {
		return err
	}

	mapping := &unstructured.Unstructured{}
	mapping.SetGroupVersionKind(GatewayMappingGVK())
	watchResources := []runtime.Object{
		&kubevirtv1.VirtualMachineInstance{},
		&corev1.Service{},
		mapping,
	}

	for _, subResource := range watchResources {
		err = c.Watch(&source.Kind{Type: subResource}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &dvv1alpha1.DroidVirt{},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileDroidVirt implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileDroidVirt{}

// ReconcileDroidVirt reconciles a DroidVirt object
type ReconcileDroidVirt struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

type RelatedResource struct {
	Virt    *dvv1alpha1.DroidVirt
	VMI     *kubevirtv1.VirtualMachineInstance
	SVC     *corev1.Service
	Mapping *unstructured.Unstructured
}

func (r *ReconcileDroidVirt) FetchRelatedResource(request reconcile.Request) (*RelatedResource, error) {
	res := &RelatedResource{}

	virt := &dvv1alpha1.DroidVirt{}
	err := r.client.Get(context.TODO(), request.NamespacedName, virt)
	if errors.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	res.Virt = virt

	vmi, err := r.relatedVMI(res.Virt)
	if err != nil {
		return res, err
	}
	res.VMI = vmi

	svc, err := r.relatedSVC(res.Virt)
	if err != nil {
		return res, err
	}
	res.SVC = svc

	mapping, err := r.relatedMapping(res.Virt)
	if err != nil {
		return res, err
	}
	res.Mapping = mapping

	return res, nil
}

type ReconcileAction struct {
	Tag      string
	Expected func(*RelatedResource) bool
	Action   func(*RelatedResource, *ReconcileDroidVirt) (reconcile.Result, error)
}

var actions = []ReconcileAction{
	{
		Tag: "VirtualMachineInstance committed",
		Expected: func(res *RelatedResource) bool {
			return res.VMI == nil
		},
		Action: func(res *RelatedResource, r *ReconcileDroidVirt) (reconcile.Result, error) {
			dataPVC, err := r.findReadyDataVolumePVC(res.Virt)
			if dataPVC == nil || err != nil {
				return reconcile.Result{}, fmt.Errorf("can not find PVC of data DroidVirtVolume, err: %v", err)
			}
			if res.Virt.Status.DataPVC != dataPVC.Name {
				res.Virt.Status.DataPVC = dataPVC.Name
				_ = r.syncStatus(res.Virt)
			}

			vmi, err := r.newVMIForDroidVirt(res.Virt, dataPVC)
			if err != nil {
				return reconcile.Result{}, fmt.Errorf("generate VMI spec error: %v", err)
			}

			// Create VMI if not found
			if err := utils.CreateIfNotExistsVMI(r.client, r.scheme, vmi); err != nil {
				return reconcile.Result{
					Requeue:      true,
					RequeueAfter: time.Second * 30,
				}, err
			}
			return reconcile.Result{}, nil
		},
	},
	{
		Tag: "VirtualMachineInstance pending",
		Expected: func(res *RelatedResource) bool {
			return res.VMI != nil &&
				(res.VMI.Status.Phase == kubevirtv1.VmPhaseUnset ||
				res.VMI.Status.Phase == kubevirtv1.Pending ||
				res.VMI.Status.Phase == kubevirtv1.Scheduling ||
				res.VMI.Status.Phase == kubevirtv1.Scheduled)
		},
		Action: func(res *RelatedResource, r *ReconcileDroidVirt) (reconcile.Result, error) {
			workerPod, err := r.relatedPod(res.VMI)
			if err != nil || workerPod == nil || workerPod.GetDeletionTimestamp() != nil {
				return reconcile.Result{
					Requeue:      true,
					RequeueAfter: time.Second * 3,
				}, err
			}
			if r.isPodBlocking(workerPod) {
				err := utils.DeleteVMI(r.client, res.VMI)
				if err != nil {
					return reconcile.Result{
						Requeue:      true,
						RequeueAfter: time.Second * 3,
					}, err
				}
				res.Virt.Status.Phase = ""
				res.Virt.Status.RelatedPod = ""
				res.Virt.Status.Interfaces = nil
				r.appendLog(res.Virt, fmt.Sprintf("VMI's pod is blocking, delete failed VMI: %s/%s (%s)", res.VMI.Namespace, res.VMI.Name, res.VMI.UID))
				_ = r.syncStatus(res.Virt)
				return reconcile.Result{}, nil
			} else {
				res.Virt.Status.Phase = dvv1alpha1.VirtualMachinePending
				res.Virt.Status.RelatedPod = workerPod.Name
				r.appendLog(res.Virt, fmt.Sprintf("VMI is created, worker pod: %s/%s", workerPod.Namespace, workerPod.Name))
				_ = r.syncStatus(res.Virt)
				// TODO VMI's pod may block on Pending, caused by ceph
				// kubelet log is "rbd: map failed exit status 1, rbd output: rbd: sysfs write failed\nIn some cases useful info is found in syslog - try \"dmesg | tail\".\nrbd: map failed: (1) Operation not permitted\n"
				return reconcile.Result{
					Requeue:      true,
					RequeueAfter: time.Second * 30,
				}, nil
			}
		},
	},
	{
		Tag: "VirtualMachineInstance running",
		Expected: func(res *RelatedResource) bool {
			return res.VMI != nil &&
				(res.VMI.Status.Phase == kubevirtv1.Running &&
					len(res.VMI.Status.Interfaces) > 0)
		},
		Action: func(res *RelatedResource, r *ReconcileDroidVirt) (reconcile.Result, error) {
			virtInterfaces := []dvv1alpha1.VMINetworkInterface{}
			for _, item := range res.VMI.Status.Interfaces {
				virtInterfaces = append(virtInterfaces, dvv1alpha1.VMINetworkInterface{
					IP:  item.IP,
					MAC: item.MAC,
				})
			}
			res.Virt.Status.Phase = dvv1alpha1.VirtualMachineRunning
			res.Virt.Status.Interfaces = virtInterfaces
			r.appendLog(res.Virt, fmt.Sprintf("VMI is running: %s/%s (%s)", res.VMI.Namespace, res.VMI.Name, res.VMI.UID))
			_ = r.syncStatus(res.Virt)
			return reconcile.Result{}, nil
		},
	},
	{
		Tag: "VirtualMachineInstance failed",
		Expected: func(res *RelatedResource) bool {
			return res.VMI != nil &&
				(res.VMI.Status.Phase == kubevirtv1.Succeeded ||
					res.VMI.Status.Phase == kubevirtv1.Failed)
		},
		Action: func(res *RelatedResource, r *ReconcileDroidVirt) (reconcile.Result, error) {
			err := utils.DeleteVMI(r.client, res.VMI)
			if err != nil {
				return reconcile.Result{
					Requeue:      true,
					RequeueAfter: time.Second * 3,
				}, err
			}
			res.Virt.Status.Phase = ""
			res.Virt.Status.RelatedPod = ""
			res.Virt.Status.Interfaces = nil
			r.appendLog(res.Virt, fmt.Sprintf("delete failed VMI: %s/%s (%s)", res.VMI.Namespace, res.VMI.Name, res.VMI.UID))
			_ = r.syncStatus(res.Virt)
			return reconcile.Result{}, nil
		},
	},
	{
		Tag: "gateway mapping Service committed",
		Expected: func(res *RelatedResource) bool {
			return res.Virt.Status.Phase == dvv1alpha1.VirtualMachineRunning &&
				len(res.Virt.Status.Interfaces) > 0
		},
		Action: func(res *RelatedResource, r *ReconcileDroidVirt) (result reconcile.Result, e error) {
			service, err := r.newServiceForDroidVirt(res.Virt)
			if err != nil {
				return reconcile.Result{}, fmt.Errorf("generate Service spec error: %v", err)
			}

			// Create Service if not found
			if err := utils.CreateIfNotExistsService(r.client, r.scheme, service); err != nil {
				return reconcile.Result{
					Requeue:      true,
					RequeueAfter: time.Second * 30,
				}, err
			}

			gatewaySpec := res.Virt.Status.Gateway
			if gatewaySpec == nil {
				gatewaySpec = &dvv1alpha1.VirtGateway{
					VncWebsocketUrl: fmt.Sprintf("/droid/%s/ws", res.Virt.UID),
				}
			}
			mapping, err := r.newGatewayMappingForService(gatewaySpec, res.Virt, service)
			if err != nil {
				return reconcile.Result{}, fmt.Errorf("generate Mapping spec error: %v", err)
			}

			// Create Mapping if not found
			if err := utils.CreateIfNotExistsMapping(r.client, mapping); err != nil {
				return reconcile.Result{
					Requeue:      true,
					RequeueAfter: time.Second * 30,
				}, err
			}

			return reconcile.Result{}, nil
		},
	},
	{
		Tag: "gateway mapping Service ready",
		Expected: func(res *RelatedResource) bool {
			return res.Virt.Status.Phase == dvv1alpha1.VirtualMachineRunning &&
				len(res.Virt.Status.Interfaces) > 0 &&
				res.SVC != nil &&
				res.Mapping != nil
		},
		Action: func(res *RelatedResource, r *ReconcileDroidVirt) (result reconcile.Result, e error) {
			res.Virt.Status.Phase = dvv1alpha1.GatewayServiceReady
			res.Virt.Status.Gateway = &dvv1alpha1.VirtGateway{
				VncWebsocketUrl: fmt.Sprintf("/droid/%s/ws", res.Virt.UID),
			}
			r.appendLog(res.Virt, fmt.Sprintf("gateway mapping Service is ready: %s, %s", res.SVC.GetName(), res.Mapping.GetName()))
			_ = r.syncStatus(res.Virt)
			return reconcile.Result{}, nil
		},
	},
}

// Reconcile reads that state of the cluster for a DroidVirt object and makes changes based on the state read
// and what is in the DroidVirt.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileDroidVirt) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	res, err := r.FetchRelatedResource(request)
	if err != nil || res == nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Reconciling DroidVirt", "Phase", res.Virt.Status.Phase)

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
