#!/bin/bash
# Script to create a KIND cluster for testing IAM Manager

set -e

# Configuration
TOOLS_DIR="$(dirname "$0")/tools"
KIND_BIN="${TOOLS_DIR}/bin/kind"
CLUSTER_NAME="iam-manager-test"
KUBECONFIG_PATH="${TOOLS_DIR}/kubeconfig"

# Add tools to path
export PATH="${TOOLS_DIR}/bin:$PATH"

# Ensure KIND is installed
if [ ! -f "${KIND_BIN}" ]; then
  echo "KIND not found at ${KIND_BIN}. Installing..."
  $(dirname "$0")/install-kind.sh
fi

# Check if a cluster with the same name already exists
if ${KIND_BIN} get clusters | grep -q "${CLUSTER_NAME}"; then
  echo "Cluster '${CLUSTER_NAME}' already exists. Deleting..."
  ${KIND_BIN} delete cluster --name "${CLUSTER_NAME}"
fi

# Create a simplified KIND config
cat <<EOF > "${TOOLS_DIR}/kind-config.yaml"
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
EOF

# Create the KIND cluster with minimal resources
echo "Creating KIND cluster '${CLUSTER_NAME}'..."
${KIND_BIN} create cluster --name "${CLUSTER_NAME}" --config="${TOOLS_DIR}/kind-config.yaml" --kubeconfig="${KUBECONFIG_PATH}" --wait 5m

# Configure KUBECONFIG env var for convenience
export KUBECONFIG="${KUBECONFIG_PATH}"
echo "Cluster created successfully!"
echo "To use this cluster, run:"
echo "export KUBECONFIG=${KUBECONFIG_PATH}"

# Test the cluster connectivity
echo "Testing cluster connectivity..."
kubectl get nodes

# Install CRDs for IAM Manager
echo "Installing IAM Manager CRDs..."
kubectl apply -f config/crd/bases/

echo "KIND cluster is ready for testing IAM Manager!"
echo ""
echo "Run tests with:"
echo "export KUBECONFIG=${KUBECONFIG_PATH}"
echo "make test"
