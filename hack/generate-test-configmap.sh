#!/bin/bash
# Script to generate a test ConfigMap for IAM Manager testing

mkdir -p hack/test

cat > hack/test/test-configmap.yaml << EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: iamroles-v1alpha1-configmap
  namespace: dev
data:
  iam.policy.action.prefix.whitelist: "s3:ListBucket,sts:,ec2:Describe,acm:Describe,acm:List,acm:Get,route53:Get,route53:List,route53:Create,route53:Delete,route53:Change,kms:Decrypt,kms:Encrypt,kms:ReEncrypt,kms:GenerateDataKey,kms:DescribeKey,dynamodb:,secretsmanager:GetSecretValue,es:,sqs:SendMessage,sqs:ReceiveMessage,sqs:DeleteMessage,SNS:Publish,sqs:GetQueueAttributes,sqs:GetQueueUrl"
  iam.policy.resource.blacklist: "policy-resource,kops"
  iam.policy.s3.restricted.resource: "s3-resource"
  aws.accountId: "123456789012"
  aws.region: "us-west-2"
  aws.MasterRole: "masters.cluster.k8s.local"
  iam.managed.policies: "shared.policy"
  iam.managed.permission.boundary.policy: "iam-manager-permission-boundary"
  webhook.enabled: "true"
  iam.role.max.limit.per.namespace: "10"
  controller.desired.frequency: "30"
  k8s.cluster.name: "test-cluster"
  irsaEnabled: "true" 
  k8s.cluster.oidc.issuer.url: "https://oidc.eks.us-west-2.amazonaws.com/test"
  iam.default.trust.policy: "default-policy"
EOF

echo "Test ConfigMap generated at hack/test/test-configmap.yaml"
