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
  iam.policy.action.prefix.whitelist: "s3:,sts:,ec2:Describe"
  iam.policy.resource.blacklist: "kops"
  iam.policy.s3.restricted.resource: "*"
  aws.accountId: "123456789012"
  aws.region: "us-west-2"
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
