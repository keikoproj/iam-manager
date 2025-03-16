#!/bin/bash
# Script to download and install KIND locally for the repository

set -e

# Configuration
TOOLS_DIR="$(dirname "$0")/tools"
KIND_VERSION="v0.20.0"
ARCH=$(uname -m)

# Convert architecture to KIND's format
case "${ARCH}" in
  x86_64)
    KIND_ARCH="amd64"
    ;;
  arm64)
    KIND_ARCH="arm64"
    ;;
  *)
    echo "Unsupported architecture: ${ARCH}"
    exit 1
    ;;
esac

# Create tools directory if it doesn't exist
mkdir -p "${TOOLS_DIR}/bin"

KIND_URL="https://kind.sigs.k8s.io/dl/${KIND_VERSION}/kind-darwin-${KIND_ARCH}"
KIND_BIN="${TOOLS_DIR}/bin/kind"

echo "Downloading KIND ${KIND_VERSION} for ${ARCH}..."
curl -Lo "${KIND_BIN}" "${KIND_URL}"
chmod +x "${KIND_BIN}"

echo "KIND installed successfully at ${KIND_BIN}"
echo "Run the following command to add it to your PATH:"
echo "  export PATH=\"${TOOLS_DIR}/bin:\$PATH\""

# Add bin directory to PATH for this script
export PATH="${TOOLS_DIR}/bin:$PATH"

# Verify installation
"${KIND_BIN}" version
