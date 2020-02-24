#!/bin/bash
## $1 => cluster name
## $2 => region
## $3 => aws_profile
## $4 => master node instance profile ARN

##Split cluster name by "." delimeter to avoid naming syntax issues
cluster=$(echo $1 | cut -d. -f1)

## Execute CFN using awscli command
policyList=`cat allowed_policies.txt`
echo $policyList
aws cloudformation update-stack --stack-name iam-manager-$cluster-cfn --use-previous-template --parameters ParameterKey=DeploymentType,ParameterValue=kiam ParameterKey=ParamK8sTrustRole,ParameterValue=$4 ParameterKey=ParamK8sClusterName,ParameterValue=$1 ParameterKey=AllowedPolicyList,ParameterValue=$policyList --capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND --on-failure DELETE --region $2 --profile $3

## install config map
kubectl apply -f iam-manager/iammanager.keikoproj.io_iamroles-configmap.yaml

## add kiam annotation to deployment spec
kubens iam-manager-system
kubectl patch deployment/iam-manager-controller-manager -p '{"spec":{"template":{"metadata":{"annotations":{"iam.amazonaws.com/role": "k8s-iam-manager-'$1'-role"}}}}}' --ns iam-manager-system

## Test and verify
