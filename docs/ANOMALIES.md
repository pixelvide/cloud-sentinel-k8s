# Supported Anomaly Detections

Cloud Sentinel automatically scans your Kubernetes resources for common misconfigurations, security risks, and reliability issues. Below is a list of the currently implemented analyzers.

## Security
| Analyzer | Description | Severity |
| :--- | :--- | :--- |
| **Privileged Container** | Detects containers running in privileged mode. | Warning |
| **Root User** | Identifies containers running as root (UID 0) or without `runAsNonRoot: true`. | Warning |
| **Immutable Tags** | Flags usage of `latest` image tags or missing tags, which can lead to non-reproducible deployments. | Warning |

## Reliability & Resilience
| Analyzer | Description | Severity |
| :--- | :--- | :--- |
| **Missing Probes** | Detects containers missing Liveness or Readiness probes. | Warning |
| **Missing PDB** | specific workloads (`Deployment`, `StatefulSet`) missing a PodDisruptionBudget. | Warning |
| **Single Replica** | Flags Deployments/StatefulSets with only 1 replica (single point of failure). | Warning |
| **Resource Limits** | Identifies containers without CPU/Memory limits or requests defined. | Warning |

## Architecture & Topology
| Analyzer | Description | Severity |
| :--- | :--- | :--- |
| **Topology Spread** | Detects workloads missing `topologySpreadConstraints` for high availability. | Warning |
| **Conflicting Affinity** | Identifies Pods with conflicting Affinity and Anti-Affinity rules that might make scheduling impossible. | Warning |

## Hygiene & Best Practices
| Analyzer | Description | Severity |
| :--- | :--- | :--- |
| **Dangling Service** | Detects Services whose label selectors do not match any active Pods. | Warning |
| **Empty Namespace** | Identifies namespaces that contain no significant workloads or services. | Warning |
| **Deprecated Ingress** | Flags Ingress resources using the deprecated `kubernetes.io/ingress.class` annotation. | Warning |
