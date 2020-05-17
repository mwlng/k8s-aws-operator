package controller

import (
	"github.com/mwlng/k8s-aws-operator/pkg/controller/ssmconfig"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, ssmconfig.Add)
}
