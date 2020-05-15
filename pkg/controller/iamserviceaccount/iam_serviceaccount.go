package iamserviceaccount

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/lithammer/shortuuid"
	"github.com/redhat-cop/operator-utils/pkg/util"
)

type InvalidIamServiceAccountError struct {
	Err error
}

func (e *InvalidIamServiceAccountError) Error() string {
	return e.Err.Error()
}

type IamServiceAccount struct {
	sa       *corev1.ServiceAccount
	roleName string
}

func NewIamServiceAccount(sa *corev1.ServiceAccount) *IamServiceAccount {
	return &IamServiceAccount{
		sa:       sa,
		roleName: "",
	}
}

func (i *IamServiceAccount) IsIamServiceAccount() bool {
	if t, ok := i.sa.Labels["type"]; ok && t == "iam" {
		return true
	}
	return false
}

func (i *IamServiceAccount) Validate() error {
	if id, ok := i.sa.Annotations["accountId"]; !ok || strings.TrimSpace(id) == "" {
		return &InvalidIamServiceAccountError{
			Err: fmt.Errorf("Missing or empty annotation field: %s", "accountId"),
		}
	}

	if id, ok := i.sa.Annotations["eksId"]; !ok || strings.TrimSpace(id) == "" {
		return &InvalidIamServiceAccountError{
			Err: fmt.Errorf("Missing or empty annotation field: %s", "eksId"),
		}
	}

	if policies, ok := i.sa.Annotations["attachedPolicies"]; !ok || strings.TrimSpace(policies) == "" {
		return &InvalidIamServiceAccountError{
			Err: fmt.Errorf("Missing or empty annotation field: %s", "attachedPolicies"),
		}
	}

	return nil
}

func (i *IamServiceAccount) GetInstance() *corev1.ServiceAccount {
	return i.sa
}

func (i *IamServiceAccount) GetAnnotations() map[string]string {
	return i.sa.Annotations
}

func (i *IamServiceAccount) GetName() string {
	return i.sa.Name
}

func (i *IamServiceAccount) GetNamespace() string {
	return i.sa.Namespace
}

func (i *IamServiceAccount) GetClusterName() string {
	clusterName := "unknown"
	if name, ok := i.sa.Labels["cluster"]; ok {
		clusterName = name
	}
	return clusterName
}

func (i *IamServiceAccount) GetAccountId() string {
	if id, ok := i.sa.Annotations["accountId"]; ok {
		return strings.TrimSpace(id)
	}
	return ""
}

func (i *IamServiceAccount) GetEksId() string {
	if id, ok := i.sa.Annotations["eksId"]; ok {
		return strings.TrimSpace(id)
	}
	return ""
}

func (i *IamServiceAccount) GetAttachedPolicies() []string {
	ret := []string{}
	if policies, ok := i.sa.Annotations["attachedPolicies"]; ok {
		attachedPolicies := strings.Split(policies, ",")
		for _, p := range attachedPolicies {
			policy := strings.TrimSpace(p)
			if len(policy) != 0 {
				ret = append(ret, policy)
			}
		}
	}
	return ret
}

func (i *IamServiceAccount) GetRoleName() string {
	if len(strings.TrimSpace(i.roleName)) == 0 {
		if roleName, ok := i.sa.Annotations["roleName"]; ok && len(strings.TrimSpace(roleName)) != 0 {
			i.roleName = roleName
		} else {
			i.roleName = i.generateRoleName()
		}
	}
	return i.roleName
}

func (i *IamServiceAccount) generateRoleName() string {
	sid := shortuuid.New()
	roleName := fmt.Sprintf("eks-%s-%s-%s-%s",
		i.GetClusterName(),
		i.GetNamespace(),
		i.GetName(),
		sid)
	if len(roleName) > 64 {
		roleName = fmt.Sprintf("eks-%s-%s-%s",
			i.GetClusterName(),
			i.GetNamespace(),
			sid)
	}
	return roleName
}

func (i *IamServiceAccount) IsBeingDeleted() bool {
	return util.IsBeingDeleted(i.sa)
}

func (i *IamServiceAccount) HasFinalizer(finalizerName string) bool {
	return util.HasFinalizer(i.sa, finalizerName)
}

func (i *IamServiceAccount) SetFinalizers(finalizers []string) {
	i.sa.SetFinalizers(finalizers)
}

func (i *IamServiceAccount) GetFinalizers() []string {
	return i.sa.GetFinalizers()
}

func (i *IamServiceAccount) RemoveFinalizer(finalizerName string) {
	util.RemoveFinalizer(i.sa, finalizerName)
}
