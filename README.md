# Kube Sentinel - Modern Kubernetes Dashboard

<div align="center">

<img src="./docs/assets/logo.svg" alt="Kube Sentinel Logo" width="128" height="128">

_A modern, intuitive Kubernetes dashboard_

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![React](https://img.shields.io/badge/React-19+-61DAFB?style=flat&logo=react)](https://reactjs.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-5+-3178C6?style=flat&logo=typescript)](https://www.typescriptlang.org)
[![License](https://img.shields.io/badge/License-Apache-green.svg)](LICENSE)
[![HelloGitHub](https://api.hellogithub.com/v1/widgets/recommend.svg?rid=a8bd165df55f41a295b62c716228b007&claim_uid=w5uk718RFhDzdCX&theme=small)](https://hellogithub.com/repository/pixelvide/kube-sentinel)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/pixelvide/kube-sentinel)

[**Documentation**](https://pixelvide.github.io/kube-sentinel)
<br>
**English**

</div>

Kube Sentinel is a lightweight, modern Kubernetes dashboard that provides an intuitive interface for managing and monitoring your Kubernetes clusters. It offers real-time metrics, comprehensive resource management, multi-cluster support, and a beautiful user experience.

> [!WARNING]
> This project is currently in rapid development and testing, and the usage and API may change.

![Dashboard Overview](docs/screenshots/overview.png)
_Comprehensive cluster overview with real-time metrics and resource statistics_

## ✨ Features

### 🎯 **Modern User Experience**

- 🌓 **Multi-Theme Support** - Dark/light/color themes with system preference detection
- 🔍 **Advanced Search** - Global search across all resources
- 🌐 **Internationalization** - Support for English language
- 📱 **Responsive Design** - Optimized for desktop, tablet, and mobile devices

### 🏘️ **Multi-Cluster Management**

- 🔄 **Seamless Cluster Switching** - Switch between multiple Kubernetes clusters
- 📊 **Per-Cluster Monitoring** - Independent Prometheus configuration for each cluster
- ⚙️ **Kubeconfig Integration** - Automatic discovery of clusters from your kubeconfig file
- 🔐 **Cluster Access Control** - Fine-grained permissions for cluster access management

### 🔍 **Comprehensive Resource Management**

- 📋 **Full Resource Coverage** - Pods, Deployments, Services, ConfigMaps, Secrets, PVs, PVCs, Nodes, and more
- 📄 **Live YAML Editing** - Built-in Monaco editor with syntax highlighting and validation
- 📊 **Detailed Resource Views** - In-depth information with containers, volumes, events, and conditions
- 🔗 **Resource Relationships** - Visualize connections between related resources (e.g., Deployment → Pods)
- ⚙️ **Resource Operations** - Create, update, delete, scale, and restart resources directly from the UI
- 🔄 **Custom Resources** - Full support for CRDs (Custom Resource Definitions)
- 🏷️ **Quick Image Tag Selector** - Easily select and change container image tags based on Docker and container registry APIs
- 🎨 **Customizable Sidebar** - Customize sidebar visibility and order, and add CRDs for quick access
- 🔌 **Kube Proxy** - Access pods or services directly through Kube Sentinel, no more `kubectl port-forward`

### 📈 **Monitoring & Observability**

- 📊 **Real-time Metrics** - CPU, memory, and network usage charts powered by Prometheus
- 📋 **Cluster Overview** - Comprehensive cluster health and resource statistics
- 📝 **Live Logs** - Stream pod logs in real-time with filtering and search capabilities
- 💻 **Web/Node Terminal** - Execute commands directly in pods/nodes through the browser
- 📈 **Node Monitoring** - Detailed node-level performance metrics and utilization
- 📊 **Pod Monitoring** - Individual pod resource usage and performance tracking

### 🔐 **Security**

- 🛡️ **OAuth Integration** - Supports OAuth management in the UI
- 🔒 **Role-Based Access Control** - Supports user permission management in the UI
- 👥 **User Management** - Comprehensive user management and role allocation in the UI
- 🔍 **Security Scanning** - Integration with Trivy Operator for vulnerability, config audit, and secret scanning
- 📊 **Security Dashboard** - Cluster-wide security posture overview with severity breakdowns

---

## 🚀 Quick Start

For detailed instructions, please refer to the [documentation](https://pixelvide.github.io/kube-sentinel/guide/installation.html).

### Docker

To run Kube Sentinel using Docker, you can use the pre-built image:

```bash
docker run --rm -p 8080:8080 ghcr.io/pixelvide/kube-sentinel:1.2.0
```

### Deploy in Kubernetes

#### Using Helm (Recommended)

1. **Add Helm repository**

   ```bash
   helm repo add kube-sentinel https://pixelvide.github.io/kube-sentinel
   helm repo update
   ```

2. **Install with default values**

   ```bash
   helm install kube-sentinel kube-sentinel/kube-sentinel -n kube-system
   ```

#### Using kubectl

1. **Apply deployment manifests**

   ```bash
   kubectl apply -f deploy/install.yaml
   # or install it online
   kubectl apply -f https://raw.githubusercontent.com/pixelvide/kube-sentinel/refs/tags/v1.2.0/deploy/install.yaml
   ```

2. **Access via port-forward**

   ```bash
   kubectl port-forward -n kube-system svc/kube-sentinel 8080:8080
   ```

### Build from Source

#### 📋 Prerequisites

1. **Clone the repository**

   ```bash
   git clone https://github.com/pixelvide/kube-sentinel.git
   cd kube-sentinel
   ```

2. **Build the project**

   ```bash
   make deps
   make build
   ```

3. **Run the server**

   ```bash
   make run
   ```

---

## 🔍 Troubleshooting

For troubleshooting, please refer to the [documentation](https://pixelvide.github.io/kube-sentinel).


## 🤝 Contributing

We welcome contributions! Please see our [contributing guidelines](https://pixelvide.github.io/kube-sentinel/faq.html#how-can-i-contribute-to-kube-sentinel) for details on how to get involved.

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
