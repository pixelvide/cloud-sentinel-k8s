# cloud-sentinel-k8s

![Version: v0.9.0](https://img.shields.io/badge/Version-v0.9.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.9.0](https://img.shields.io/badge/AppVersion-v0.9.0-informational?style=flat-square)

A Helm chart for Kubernetes Dashboard - Cloud Sentinel K8s

## Installation

### Add Helm Repository

```bash
helm repo add cloud-sentinel-k8s https://pixelvide.github.io/cloud-sentinel-k8s
helm repo update
```

### Install Chart

```bash
# Install in kube-system namespace (recommended)
helm install cloud-sentinel-k8s cloud-sentinel-k8s/cloud-sentinel-k8s -n kube-system

# Or install in custom namespace
helm install cloud-sentinel-k8s cloud-sentinel-k8s/cloud-sentinel-k8s -n my-namespace --create-namespace
```

### Upgrade Chart

```bash
helm upgrade cloud-sentinel-k8s cloud-sentinel-k8s/cloud-sentinel-k8s -n kube-system
```

### Uninstall Chart

```bash
helm uninstall cloud-sentinel-k8s -n kube-system
```

### Chart Values

[Chart Values](https://cloud-sentinel-k8s.pixelvide.cloud/config/chart-values)
