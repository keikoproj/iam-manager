#!/bin/bash
## $1 => cluster name
## $2 => region
## $3 => aws_profile

## install cert-manager
kubectl apply -f cert-manager/cert-manager-v0.12.0.yaml --validate=false

##Split cluster name by "." delimeter to avoid naming syntax issues
cluster=$(echo $1 | cut -d. -f1)

## Execute CFN using awscli command
policyList=`cat allowed_policies.txt`
echo $policyList
aws cloudformation create-stack --stack-name iam-manager-$cluster-cfn --template-body file://iam-manager-cfn.yaml --parameters ParameterKey=ParamK8sClusterName,ParameterValue=$1 ParameterKey=AllowedPolicyList,ParameterValue=$policyList --capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND --on-failure DELETE --region $2 --profile $3

## wget iam-manager
kubectl apply -f iam-manager/iam-manager.yaml


## install config map
kubectl apply -f iam-manager/iammanager.keikoproj.io_iamroles-configmap.yaml

## Test and verify
