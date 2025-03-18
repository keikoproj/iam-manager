#!/bin/bash
set -e

# Script to install iam-manager on a Kubernetes cluster
# 
# Usage:
#   ./install.sh [cluster_name] [aws_region] [aws_profile]
#
# Arguments:
#   cluster_name - Name of the Kubernetes cluster
#   aws_region   - AWS region where resources will be created
#   aws_profile  - AWS profile to use for authentication

# Display usage information
function usage() {
  echo "Usage: $0 <cluster_name> <aws_region> <aws_profile>"
  echo ""
  echo "Arguments:"
  echo "  cluster_name   - Name of the Kubernetes cluster"
  echo "  aws_region     - AWS region where resources will be created (e.g., us-west-2)"
  echo "  aws_profile    - AWS profile to use for authentication"
  echo ""
  echo "Example:"
  echo "  $0 my-eks-cluster us-west-2 my-aws-profile"
  exit 1
}

# Check if all required parameters are provided
if [ $# -ne 3 ]; then
  echo "Error: Missing required parameters."
  usage
fi

# Validate parameters
if [ -z "$1" ]; then
  echo "Error: Cluster name cannot be empty."
  usage
fi

if [ -z "$2" ]; then
  echo "Error: AWS region cannot be empty."
  usage
fi

if [ -z "$3" ]; then
  echo "Error: AWS profile cannot be empty."
  usage
fi

# Assign parameters to named variables for better readability
CLUSTER_NAME=$1
AWS_REGION=$2
AWS_PROFILE=$3

# Check if AWS profile exists
if ! aws configure list-profiles 2>/dev/null | grep -q "^$AWS_PROFILE$"; then
  echo "Warning: AWS profile '$AWS_PROFILE' not found in your AWS config."
  echo "Please make sure the profile exists or check your AWS configuration."
  read -p "Continue anyway? (y/n): " CONTINUE
  if [[ ! $CONTINUE =~ ^[Yy]$ ]]; then
    echo "Installation aborted."
    exit 1
  fi
fi

# Verify kubectl is installed and configured
if ! command -v kubectl &> /dev/null; then
  echo "Error: kubectl is not installed or not in PATH."
  echo "Please install kubectl and try again."
  exit 1
fi

# Check if the current kubectl context is pointing to the correct cluster
CURRENT_CONTEXT=$(kubectl config current-context 2>/dev/null || echo "none")
echo "Current kubectl context: $CURRENT_CONTEXT"
read -p "Is this the correct Kubernetes cluster? (y/n): " CORRECT_CLUSTER
if [[ ! $CORRECT_CLUSTER =~ ^[Yy]$ ]]; then
  echo "Please set the correct kubectl context and try again."
  exit 1
fi

echo "Installing iam-manager for cluster: $CLUSTER_NAME in region: $AWS_REGION using AWS profile: $AWS_PROFILE"
echo "---"

# Split cluster name by "." delimiter to avoid naming syntax issues
CLUSTER_SHORT_NAME=$(echo $CLUSTER_NAME | cut -d. -f1)

# Get allowed policies from file
POLICY_LIST=$(cat allowed_policies.txt)
if [ -z "$POLICY_LIST" ]; then
  echo "Warning: allowed_policies.txt is empty. The permission boundary will not have any allowed actions."
  read -p "Continue anyway? (y/n): " CONTINUE_EMPTY
  if [[ ! $CONTINUE_EMPTY =~ ^[Yy]$ ]]; then
    echo "Installation aborted."
    exit 1
  fi
fi

echo "Allowed policies: $POLICY_LIST"

# Create CloudFormation stack
echo "Creating CloudFormation stack for IAM resources..."
aws cloudformation create-stack \
  --stack-name iam-manager-$CLUSTER_SHORT_NAME-cfn \
  --template-body file://iam-manager-cfn.yaml \
  --parameters \
    ParameterKey=ParamK8sClusterName,ParameterValue=$CLUSTER_NAME \
    ParameterKey=AllowedPolicyList,ParameterValue="$POLICY_LIST" \
  --capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND \
  --on-failure DELETE \
  --region $AWS_REGION \
  --profile $AWS_PROFILE

echo "Waiting for CloudFormation stack creation to complete..."
aws cloudformation wait stack-create-complete \
  --stack-name iam-manager-$CLUSTER_SHORT_NAME-cfn \
  --region $AWS_REGION \
  --profile $AWS_PROFILE

# Install iam-manager in the cluster
echo "Installing iam-manager Kubernetes resources..."
kubectl apply -f iam-manager/iam-manager.yaml

# Install ConfigMap
echo "Applying iam-manager ConfigMap..."
kubectl apply -f iam-manager/iammanager.keikoproj.io_iamroles-configmap.yaml

echo "---"
echo "Installation complete!"
echo "You can verify the installation by running:"
echo "  kubectl get pods -n iam-manager-system"
