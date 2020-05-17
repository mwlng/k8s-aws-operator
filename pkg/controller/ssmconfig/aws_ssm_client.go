package ssmconfig

import (
	"github.com/mwlng/aws-go-clients/clients"
	"github.com/mwlng/aws-go-clients/service"
)

var (
	ssmClient *clients.SSMClient
)

func init() {
	svc := service.Service{
		Region: "us-east-1",
		//Profile: "default",
	}
	sess := svc.NewSession()

	ssmClient = clients.NewClient("ssm", sess).(*clients.SSMClient)
}
