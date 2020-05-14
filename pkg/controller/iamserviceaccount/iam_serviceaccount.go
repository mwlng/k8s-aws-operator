package iamserviceaccount

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/lithammer/shortuuid"
)

type IamServiceAccount struct {
	sa *corev1.ServiceAccount
}

func NewIamServiceAccount(sa *corev1.ServiceAccount) *IamServiceAccount {
	return &IamServiceAccount{
		sa: sa,
	}
}

func (i *IamServiceAccount) isIamServiceAccount() bool {
	if t, ok := i.sa.Labels["type"]; ok && t == "iam" {
		return true
	}
	return false
}

func (i *IamServiceAccount) validate() error {
	if id, ok := i.sa.Annotations["accountId"]; !ok && strings.TrimSpace(id) == "" {
		return nil
	}

	if id, ok := i.sa.Annotations["eksId"]; !ok && strings.TrimSpace(id) == "" {
		return nil
	}

	if policies, ok := i.sa.Annotations["attachedPolicies"]; !ok && strings.TrimSpace(policies) == "" {
		return nil
	}

	return nil
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
			ret = append(ret, strings.TrimSpace(p))
		}
	}
	return ret
}

func (i *IamServiceAccount) GenerateRoleName() string {
	return fmt.Sprintf("eks-%s-%s-%s-%s",
		i.GetClusterName(),
		i.GetNamespace(),
		i.GetName(),
		shortuuid.New())
}
