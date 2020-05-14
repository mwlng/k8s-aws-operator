package iamserviceaccount

import (
	"github.com/mwlng/aws-go-clients/clients"
	"github.com/mwlng/aws-go-clients/service"
)

var (
	iamClient *clients.IAMClient
)

func init() {
	svc := service.Service{
		Region: "us-east-1",
		//Profile: "default",
	}
	sess := svc.NewSession()
	iamClient = clients.NewClient("iam", sess).(*clients.IAMClient)
}
