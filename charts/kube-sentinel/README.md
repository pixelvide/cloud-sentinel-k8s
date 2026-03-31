# kube-sentinel

![Version: v1.2.0](https://img.shields.io/badge/Version-v1.2.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v1.2.0](https://img.shields.io/badge/AppVersion-v1.2.0-informational?style=flat-square)

A Helm chart for Kubernetes Dashboard - Kube Sentinel

## Installation

### Add Helm Repository

```bash
helm repo add kube-sentinel https://pixelvide.github.io/kube-sentinel
helm repo update
```

### Install Chart

```bash
# Install in kube-system namespace (recommended)
helm install kube-sentinel kube-sentinel/kube-sentinel -n kube-system

# Or install in custom namespace
helm install kube-sentinel kube-sentinel/kube-sentinel -n my-namespace --create-namespace
```

### Upgrade Chart

```bash
helm upgrade kube-sentinel kube-sentinel/kube-sentinel -n kube-system
```

### Uninstall Chart

```bash
helm uninstall kube-sentinel -n kube-system
```

### Chart Values

[Chart Values](https://kube-sentinel.pixelvide.cloud/config/chart-values)
