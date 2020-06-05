package droidvirt

import (
	"context"
	droidvirtv1alpha1 "github.com/lxs137/droidvirt-ctrl/pkg/apis/droidvirt/v1alpha1"
	"github.com/lxs137/droidvirt-ctrl/pkg/syncer"
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
	err = c.Watch(&source.Kind{Type: &droidvirtv1alpha1.DroidVirt{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	watchResources := []runtime.Object{
		&kubevirtv1.VirtualMachine{},
		&corev1.Service{},
	}

	for _, subResource := range watchResources {
		err = c.Watch(&source.Kind{Type: subResource}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &droidvirtv1alpha1.DroidVirt{},
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
	client client.Client
	scheme *runtime.Scheme
	recorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a DroidVirt object and makes changes based on the state read
// and what is in the DroidVirt.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileDroidVirt) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling DroidVirt")

	// Fetch the DroidVirt instance
	instance := &droidvirtv1alpha1.DroidVirt{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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

	log.Info("Reconciling Resource %s, '%+v'", instance.Name, instance.Spec)
	// if the CRD is terminating, stop reconcile
	if instance.ObjectMeta.DeletionTimestamp != nil {
		return reconcile.Result{}, nil
	}
	syncers := make([]syncer.Interface, 0)
	syncers = append(syncers, NewProbeDeploySyncer(instance, r.client, r.scheme))
	if err := r.sync(syncers); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, r.updateStatus(instance)
}

func (r *ReconcileDroidVirt) sync(syncers []syncer.Interface) error {
	for _, s := range syncers {
		if err := syncer.Sync(context.TODO(), s, r.recorder); err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileDroidVirt) updateStatus(dv *droidvirtv1alpha1.DroidVirt) error {
	return nil
}
