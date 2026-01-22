#!/bin/bash
set -e

# Add local bin to PATH
export PATH=$(pwd)/bin:$PATH

# Create kind cluster if not exists
if ! sudo $(pwd)/bin/kind get clusters | grep -q "^kind$"; then
    echo "Creating Kind cluster..."
    # Use kind-config.yaml if it exists, otherwise default
    if [ -f "kind-config.yaml" ]; then
        sudo $(pwd)/bin/kind create cluster --config kind-config.yaml --wait 1m
    else
        sudo $(pwd)/bin/kind create cluster --wait 1m
    fi
else
    echo "Kind cluster 'kind' already exists."
fi

# Generate kubeconfig
echo "Exporting kubeconfig..."
sudo $(pwd)/bin/kind get kubeconfig --name kind > kubeconfig.yaml
sudo chown $(id -u):$(id -g) kubeconfig.yaml
export KUBECONFIG=$(pwd)/kubeconfig.yaml

# Set other env vars
export DB_DSN="e2e.db"
export CLOUD_SENTINEL_K8S_USERNAME=admin
export CLOUD_SENTINEL_K8S_PASSWORD=admin

echo "Starting server..."
go run main.go
