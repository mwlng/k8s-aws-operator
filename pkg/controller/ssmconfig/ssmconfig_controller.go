package ssmconfig

import (
	"context"
	"path/filepath"

	ssmv1alpha1 "github.com/mwlng/k8s-aws-operator/pkg/apis/ssm/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_ssmconfig")

// Add creates a new SsmConfig Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSsmConfig{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("ssmconfig-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource SsmConfig
	err = c.Watch(&source.Kind{Type: &ssmv1alpha1.SsmConfig{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Secrets and requeue the owner SsmConfig
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &ssmv1alpha1.SsmConfig{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileSsmConfig implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSsmConfig{}

// ReconcileSsmConfig reconciles a SsmConfig object
type ReconcileSsmConfig struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a SsmConfig object and makes changes based on the state read
// and what is in the SsmConfig.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSsmConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling SsmConfig")

	// Fetch the SsmConfig instance
	instance := &ssmv1alpha1.SsmConfig{}
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

	// Define a new Secret object
	secret := newSecretForCR(instance)

	// Set SsmConfig instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, secret, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	found := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
		err = r.client.Create(context.TODO(), secret)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Secret created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ensureLatestSecret(secret, found)
	if err != nil {
		return reconcile.Result{}, err
	}

	// The secret already exists - don't requeue
	//reqLogger.Info("Skip reconcile: The secret already exists", "Secret.Namespace", found.Namespace, "Secret.Name", found.Name)
	return reconcile.Result{}, nil
}

func (r *ReconcileSsmConfig) ensureLatestSecret(newSecret *corev1.Secret, foundSecret *corev1.Secret) error {
	if len(foundSecret.Data) != len(newSecret.Data) {
		err := r.client.Update(context.TODO(), newSecret)
		if err != nil {
			return err
		}
	} else {
		latest := true
		for key, val := range newSecret.Data {
			if v, ok := foundSecret.Data[key]; ok {
				if string(v) == string(val) {
					continue
				} else {
					latest = false
					break
				}
			} else {
				latest = false
				break
			}
		}
		if !latest {
			err := r.client.Update(context.TODO(), newSecret)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func newSecretForCR(cr *ssmv1alpha1.SsmConfig) *corev1.Secret {
	var data map[string][]byte = map[string][]byte{}
	for _, k := range cr.Spec.SsmKeys {
		baseKey := filepath.Base(k)
		value := ssmClient.GetParameter(k)
		data[baseKey] = []byte(value)
	}

	labels := map[string]string{
		"app": cr.Name,
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-secret",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Type: "Opaque",
		Data: data,
	}
}
