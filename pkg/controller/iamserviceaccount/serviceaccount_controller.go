package iamserviceaccount

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redhat-cop/operator-utils/pkg/util"
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

var (
	log            = logf.Log.WithName("controller_iamserviceaccount")
	controllerName = "iamserviceaccount-controller"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ServiceAccount Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileServiceAccount{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ServiceAccount
	err = c.Watch(&source.Kind{Type: &corev1.ServiceAccount{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &corev1.ServiceAccount{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileServiceAccount implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileServiceAccount{}

// ReconcileServiceAccount reconciles a ServiceAccount object
type ReconcileServiceAccount struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ServiceAccount object and makes changes based on the state read
// and what is in the ServiceAccount.Annotations etc.
// TODO(user): Modify this Reconcile function to implement your Controller logic.
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileServiceAccount) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ServiceAccount")

	// Fetch the ServiceAccount instance
	instance := &corev1.ServiceAccount{}
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

	isa := NewIamServiceAccount(instance)
	if isa.isIamServiceAccount() {
		if util.IsBeingDeleted(instance) {
			if !util.HasFinalizer(instance, controllerName) {
				return reconcile.Result{}, nil
			}
			err := cleanUp(instance)
			if err != nil {
				reqLogger.Info("Unable to delete iam service account instance",
					"Instance.Namespace", instance.Namespace, "Instance.Name", instance.Name)
				return reconcile.Result{}, err
			}
			util.RemoveFinalizer(instance, controllerName)
			err = r.client.Update(context.TODO(), instance)
			if err != nil {
				reqLogger.Info("Unable to update iam service account instance",
					"Instance.Namespace", instance.Namespace, "Instance.Name", instance.Name)
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}

		err := isa.validate()
		if err != nil {
			return reconcile.Result{Requeue: false}, err
		}

		iamRoleName := isa.GenerateRoleName()
		instance.SetFinalizers([]string{iamRoleName})

		// Define a new configmap object
		configMap := newConfigMapForSA(isa, iamRoleName)

		// Set ServiceAccount instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, configMap, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		// Check if this ConfigMap already exists
		found := &corev1.ConfigMap{}
		err = r.client.Get(context.TODO(),
			types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			err = applyChanges(isa)
			if err != nil {
				return reconcile.Result{}, err
			}
			reqLogger.Info("Creating a new ConfigMap", configMap.Namespace, "ConfigMap.Name", configMap.Name)
			err = r.client.Create(context.TODO(), configMap)
			if err != nil {
				return reconcile.Result{}, err
			}
			// ConfigMap created successfully - don't requeue
			return reconcile.Result{}, nil
		} else if err != nil {
			return reconcile.Result{}, err
		}

		// ConfigMap already exists
		reqLogger.Info("Reconcile update: ConfigMap already exists",
			"ConfigMap.Namespace", found.Namespace, "ConfigMap.Name", found.Name)
		err = applyChanges(isa)
		if err != nil {
			return reconcile.Result{}, err
		}
		err = r.client.Update(context.TODO(), configMap)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func newConfigMapForSA(isa *IamServiceAccount, iamRoleName string) *corev1.ConfigMap {
	var data map[string]string = map[string]string{}
	attachedPolicies := isa.GetAttachedPolicies()
	data["iam_role"] = iamRoleName
	jsonPolicies, _ := json.Marshal(attachedPolicies)
	data["iam_policies"] = string(jsonPolicies)

	labels := map[string]string{}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      isa.GetName(),
			Namespace: isa.GetNamespace(),
			Labels:    labels,
		},
		Data: data,
	}
}

func applyChanges(isa *IamServiceAccount) error {
	roleName := isa.GenerateRoleName()
	policies := isa.GetAttachedPolicies()
	role := iamClient.GetRole(&roleName)
	if role != nil {
		err := updateIamRoleWithPolicies(roleName, policies)
		if err != nil {
			return err
		}
	} else {
		oidcProvider := fmt.Sprintf("arn:aws:iam::%s:oidc-provider", isa.GetAccountId())
		oidcURI := fmt.Sprintf("oidc.eks.us-east-1.amazonaws.com/id/%s", isa.GetEksId())
		saID := fmt.Sprintf("system:serviceaccount:%s:%s", isa.GetNamespace(), isa.GetName())
		err := createIamRoleWithPolicies(roleName, policies, oidcProvider, oidcURI, saID)
		if err != nil {
			return err
		}
	}
	return nil
}

func cleanUp(instance *corev1.ServiceAccount) error {
	finalizers := instance.GetFinalizers()
	for _, f := range finalizers {
		err := deleteIamRole(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func createIamRoleWithPolicies(roleName string, policies []string,
	oidcProvider string, oidcURI string, saID string) error {
	path := "/"
	apd, _ := json.Marshal(&assumePolicyDocument{
		Version: "2012-10-17",
		Statement: []statement{
			{
				Effect: "Allow",
				Principal: principal{
					Federated: fmt.Sprintf("%s/%s", oidcProvider, oidcURI),
				},
				Action: "sts:AssumeRoleWithWebIdentity",
				Condition: condition{
					StringEquals: map[string]string{
						fmt.Sprintf("%s:aud", oidcURI): "sts.amazonaws.com",
						fmt.Sprintf("%s:sub", oidcURI): saID,
					},
				},
			},
		},
	})

	document := string(apd)
	if _, err := iamClient.CreateRole(&roleName, &path, &document); err != nil {
		return err
	}
	for _, p := range policies {
		err := iamClient.AttachRolePolicy(&roleName, &p)
		if err != nil {
			policies := listAttachedPolicies(roleName)
			for _, p := range policies {
				if err := iamClient.DetachRolePolicy(&roleName, &p); err != nil {
					break
				}
			}
			iamClient.DeleteRole(&roleName)
			return err
		}
	}
	return nil
}

func listAttachedPolicies(roleName string) []string {
	ret := []string{}
	for _, p := range iamClient.ListAttachedRolePolicies(&roleName) {
		ret = append(ret, *p.PolicyArn)
	}
	return ret
}

func updateIamRoleWithPolicies(roleName string, newPolicies []string) error {
	oddPolicies := listAttachedPolicies(roleName)
	new, _, missing := diff(oddPolicies, newPolicies)
	for _, n := range new {
		err := iamClient.AttachRolePolicy(&roleName, &n)
		if err != nil {
			return err
		}
	}

	for _, m := range missing {
		err := iamClient.AttachRolePolicy(&roleName, &m)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteIamRole(roleName string) error {
	role := iamClient.GetRole(&roleName)
	if role != nil {
		policies := listAttachedPolicies(roleName)
		for _, p := range policies {
			err := iamClient.DetachRolePolicy(&roleName, &p)
			if err != nil {
				return err
			}
		}
		err := iamClient.DeleteRole(&roleName)
		if err != nil {
			return err
		}
	}
	return nil
}

func diff(src []string, tgt []string) ([]string, []string, []string) {
	new := []string{}
	common := []string{}
	missing := []string{}
	lookupMap := map[string]int{}

	for _, t := range tgt {
		lookupMap[t] = 1
	}

	for _, s := range src {
		if _, ok := lookupMap[s]; !ok {
			missing = append(missing, s)
		} else {
			lookupMap[s]++
		}
	}

	for k, v := range lookupMap {
		if v == 1 {
			new = append(new, k)
		} else {
			common = append(common, k)
		}
	}
	return new, common, missing
}
