module github.com/mwlng/k8s-aws-operator

go 1.13

require (
	github.com/lithammer/shortuuid v3.0.0+incompatible
	github.com/mwlng/aws-go-clients v0.0.0-20200516152156-4dab960990d5
	github.com/operator-framework/operator-sdk v0.17.1-0.20200508235800-4e2c562a3d29
	github.com/redhat-cop/operator-utils v0.2.4
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)
