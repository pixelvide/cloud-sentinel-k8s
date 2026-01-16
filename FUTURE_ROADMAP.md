# Future Roadmap & Project Improvements

This document outlines potential feature enhancements and project improvements planned for `Cloud Sentinel`.

## 1. Testing & Quality
- **`TESTING.md`**: Strategies for unit, integration, and E2E testing.
- **`e2e/`**: Directory for End-to-End tests (e.g., Playwright/Cypress).

## 2. Deployment & IaC
- **`k8s/`**: Manifests or Helm charts for deploying Cloud Sentinel.
- **`terraform/`**: IaC for provisioning required cloud resources.

## 3. Anomaly Detection (Advanced Roadmap)

Implementation of an intelligent monitoring system to detect and alert on unusual cluster behavior.

### Phase 1: Foundation (Data Collection)
- **Metrics Integration**: Prometheus/Grafana integration to ingest CPU, Memory, and Network usage.
- **Event Streaming**: Real-time ingestion of Kubernetes events for correlation.
- **Baseline Establishment**: Historical data collection to define "normal" resource behavior.

### Phase 2: Detection Engine
- **Statistically-based Detection**: Identify outliers in resource consumption using standard deviation and Z-score methods.
- **CrashLoop Correlation**: Detect patterns of restarts and failures across related services.
- **Threshold Learning**: Dynamic adjustment of warning thresholds based on seasonal patterns.

### Phase 3: Intelligence & Automated Response
- **ML-powered Forecasting**: Predict potential resource exhaustion or OOM kills before they occur.
- **Root Cause Analysis (RCA)**: Automated grouping of related alerts to pinpoint the source of failure.
- **Proactive Remediation**: Optional automated actions (e.g., scaling up, restarting pods) based on detected anomalies.

## 4. Proposed Feature Enhancements

### Compute & Orchestration
- **Resource Editor**: Advanced YAML editor with schema validation for safely editing live resources.

### Security & Governance
- **Vulnerability Scanning**: Integration with scanners (e.g., Trivy, Clair) to display vulnerability reports for running images.
- **Policy Enforcement**: Integration with OPA/Kyverno to visualize policy violations.

### Observability & Troubleshooting
- **Advanced Log Search**: Enhancing the existing Log Viewer with regex search, time-range filtering, and download capabilities.
- **Event Timeline**: A graphical timeline of cluster events to correlate cascading failures.
- **Network Policy Viewer**: Visualizer for allowed/blocked traffic flows between namespaces/pods.

### Cost & Operations
- **Cost Estimates**: Integration with OpenCost/Kubecost to display estimated run-rate for namespaces/workloads.
- **GitOps Status**: Native integration to view reconciliation status of Flux or ArgoCD resources.
- **Alerting Integration**: Webhooks for routing critical cluster alerts to Slack, Teams, or PagerDuty.

