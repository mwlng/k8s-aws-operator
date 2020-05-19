## Overview

I guess almost all people knows the Kubernetes(k8s) is containers(Pods) management and orchestration platform. But if you dive deep, you will find in general k8s can do more than just managing containers, actually it can manage any type of resources, and make them working together as long as you follow k8s API resource model paradigm.

k8s-aws-operator is a k8s operator which will integrate AWS services into k8s clusters like EKS running on the AWS. Currently, it integrated with AWS SSM and IAM services.

## k8s AWS SSM integration

In this use case, we can dynamically import application secrets and parameters from AWS SSM parameter store into k8s namespaced secret resource during application deployment. So application can use secrets directly from k8s without extra hassle.

### SsmConfig custom resource definition(CRD)
```
apiVersion: ssm.aws.605.tv/v1alpha1
kind: SsmConfig  
metadata:
  name: airflow
  namespace: airflow
spec:
  ssmKeys:
      - "/dev-qa/airflow/AIRFLOW__SENTRY__CLIENT_SECRET"
      - "/dev-qa/airflow/FERNET_KEY" 
      ...
```

### How to use SsmConfig to create corresponding k8s secret in the cluster
* Use above SsmConfig resource definition to create an application specific SsmConfig resource definition file, or embed it into  the deployment manifest.
* Run below command as part of CD process.

```
kubectl apply -f deployment.yaml
``` 
Note: For good practice, each application should create its own namespace and put it into the metadata.namespace field. So that secret will be created in the application namespace.

### How to use k8s secret in your application/Pod

[Using secrets](https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets)

[Using Secrets as environment variables](https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-environment-variables)
)

### How it works underhood
All the magic of creating k8s secret is fully controlled by k8s-aws-operator. It will capture create/delete/update events from SsmConfig resource and reconcile its states correspondingly with k8s secret object.

## k8s AWS IAM integration
Recently, AWS SDKs has added a new credential provider that calls sts:AssumeRoleWithWebIdentity, exchanging the Kubernetes-issued OpenID Connect (OIDC) token for AWS role credentials. This OIDC federation access allows you to assume IAM roles via the Secure Token Service (STS), enabling authentication with an OIDC provider, receiving a JSON Web Token (JWT), which in turn can be used to assume an IAM role.

Kubernetes, on the other hand, can issue so-called projected service account tokens, which happen to be valid OIDC JWTs for pods.

By combining OIDC identity provider and Kubernetes service account, IAM Roles for Service Accounts (IRSA), and it's annotations, you can now use IAM roles at the pod level.

### k8s(AWS EKS) IAM service account to AWS IAM role mapping

Create or define k8s IAM service account in your application deployment manifest file (Ex: deployment.yaml)

```
apiVersion: v1
kind: ServiceAccount
metadata:
  name: airflow-iam-serviceaccount
  namespace: airflow
  labels:
    type: "iam" # <----- Label it with type "iam" (case sensitive) to make it as iam service account.
  annotations: # <----- Provide annotations with list of iam policies, so they will be added to the corresponding IAM role.
    attachedPolicies: >
      arn:aws:iam::524458042105:policy/605-SSM-Dev-ReadOnly,
      arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess
```
* Run below command as part of CD process.

```
kubectl apply -f deployment.yaml
```

Note: For good practice, each application should create its own namespace and put it into the metadata.namespace field. So that iam service account will be created in the dedicated application namespace.

### How it works underhood
All the magics of operating k8s IAM service account and it's corresponding IAM role is fully controlled by k8s-aws-operator. It will capture create/delete/update events from k8s iam service account resource and reconcile its states with k8s AWS IAM role resources correspondingly.