#!/bin/bash
set -e

mkdir -p bin

# Install kubectl
echo "Downloading kubectl..."
curl -LO "https://dl.k8s.io/release/v1.30.0/bin/linux/amd64/kubectl"
chmod +x kubectl
mv kubectl bin/

# Install kind
echo "Downloading kind..."
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.23.0/kind-linux-amd64
chmod +x ./kind
mv ./kind bin/

echo "Installation complete."
