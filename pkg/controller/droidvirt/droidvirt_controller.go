package droidvirt

import (
	"context"
	"fmt"
	"github.com/lxs137/droidvirt-ctrl/pkg/utils"
	"time"

	dvv1alpha1 "github.com/lxs137/droidvirt-ctrl/pkg/apis/droidvirt/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
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
	err = c.Watch(&source.Kind{Type: &dvv1alpha1.DroidVirt{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	watchResources := []runtime.Object{
		&kubevirtv1.VirtualMachineInstance{},
		&corev1.PersistentVolumeClaim{},
		&corev1.Service{},
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

// Reconcile reads that state of the cluster for a DroidVirt object and makes changes based on the state read
// and what is in the DroidVirt.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileDroidVirt) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	// Fetch the DroidVirt instance
	virt := &dvv1alpha1.DroidVirt{}
	err := r.client.Get(context.TODO(), request.NamespacedName, virt)
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

	reqLogger.Info("Reconciling DroidVirt", "DroidVirt.Spec", virt.Spec, "DroidVirt.Status", virt.Status)
	switch virt.Status.Phase {
	case "":
		vmi, err := r.newVMIForDroidVirt(virt)
		if err != nil {
			_ = r.appendLogAndSync(virt, fmt.Sprintf("generate VMI spec error: %v", err))
			return reconcile.Result{}, fmt.Errorf("generate VMI spec error: %v", err)
		}

		// Create VMI if not found
		if err := utils.CreateIfNotExistsVMI(r.client, r.scheme, vmi); err != nil {
			_ = r.appendLogAndSync(virt, fmt.Sprintf("creating VMI error, retry after 30s: %v", err))
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: time.Second * 30,
			}, err
		}
		reqLogger.Info("VMI created")

		virt.Status.Phase = dvv1alpha1.VirtualMachinePending
		virt.Status.RelatedVMI = vmi.Name
		r.appendLog(virt, fmt.Sprintf("VMI is created: %s/%s", vmi.Namespace, vmi.Name))
		_ = r.syncStatus(virt)
		return reconcile.Result{}, nil
	case dvv1alpha1.VirtualMachinePending:
		vmi, err := r.relatedVMI(virt)
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
			virt.Status.Phase = dvv1alpha1.VirtualMachineRunning
			r.appendLog(virt, fmt.Sprintf("VMI is running: %v", vmi.ObjectMeta))
			_ = r.syncStatus(virt)
			return reconcile.Result{}, nil
		case kubevirtv1.Succeeded, kubevirtv1.Failed, kubevirtv1.Unknown:
			virt.Status.Phase = dvv1alpha1.VirtualMachineDown
			r.appendLog(virt, fmt.Sprintf("VMI is %s: %v", vmi.Status.Phase, vmi.ObjectMeta))
			_ = r.syncStatus(virt)
			return reconcile.Result{}, fmt.Errorf("VMI is down")
		}
	case dvv1alpha1.VirtualMachineRunning:
		vmi, err := r.relatedVMI(virt)
		if vmi == nil {
			reqLogger.Error(err, "Related VMI not found")
			virt.Status.Phase = dvv1alpha1.VirtualMachineDown
			r.appendLog(virt, fmt.Sprintf("VMI is down"))
			_ = r.syncStatus(virt)
			return reconcile.Result{}, fmt.Errorf("VMI is down")
		}

		service, err := r.newServiceForDroidVirt(virt)
		if err != nil {
			_ = r.appendLogAndSync(virt, fmt.Sprintf("generate Service spec error: %v", err))
			return reconcile.Result{}, fmt.Errorf("generate Service spec error: %v", err)
		}

		// Create Service if not found
		if err := utils.CreateIfNotExistsService(r.client, r.scheme, service); err != nil {
			_ = r.appendLogAndSync(virt, fmt.Sprintf("creating Service error, retry after 30s: %v", err))
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: time.Second * 30,
			}, err
		}
		reqLogger.Info("Service created")

		gatewaySpec := virt.Status.Gateway
		if gatewaySpec == nil {
			gatewaySpec = &dvv1alpha1.VirtGateway{
				VncWebsocketUrl: fmt.Sprintf("/droid/%s/ws", virt.UID),
			}
		}
		mapping, err := r.newGatewayMappingForService(gatewaySpec, virt, service)
		if err != nil {
			_ = r.appendLogAndSync(virt, fmt.Sprintf("generate Mapping spec error: %v", err))
			return reconcile.Result{}, fmt.Errorf("generate Mapping spec error: %v", err)
		}

		// Create Mapping if not found
		if err := utils.CreateIfNotExistsMapping(r.client, mapping); err != nil {
			_ = r.appendLogAndSync(virt, fmt.Sprintf("creating Mapping error, retry after 30s: %v", err))
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: time.Second * 30,
			}, err
		}
		reqLogger.Info("Mapping created")

		virt.Status.Phase = dvv1alpha1.GatewayServiceReady
		virt.Status.Gateway = gatewaySpec
		r.appendLog(virt, fmt.Sprintf("ambassador.io/mapping is ready: %v", mapping))
		_ = r.syncStatus(virt)
		return reconcile.Result{}, nil
	case dvv1alpha1.VirtualMachineDown:
		vmi, err := r.relatedVMI(virt)
		if vmi != nil {
			err = r.client.Delete(context.TODO(), vmi)
			if err != nil {
				reqLogger.Error(err, "Down VMI can't be deleted: %v", vmi.ObjectMeta)
				return reconcile.Result{
					Requeue:      true,
					RequeueAfter: time.Second * 30,
				}, err
			}
		}
		reqLogger.Info("Down VMI has been deleted")

		virt.Status.Phase = ""
		virt.Status.RelatedVMI = ""
		r.appendLog(virt, fmt.Sprintf("Down VMI has been deleted"))
		_ = r.syncStatus(virt)
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}
