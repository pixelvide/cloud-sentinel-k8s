# Frontend Route Documentation

This document lists all available frontend routes in the application, organized by their functional groups as seen in the navigation sidebar.

## Core Resources

| Path | Header | Description | Visibility |
|------|--------|-------------|------------|
| `/` | Dashboard | Cluster overview and health status | Always |
| `/kube-nodes` | Nodes | Cluster nodes and capacity | Cluster-wide |
| `/kube-namespaces` | Namespaces | Manage cluster namespaces | Cluster-wide |

## Workloads

| Path | Header | Description | Scoping |
|------|--------|-------------|---------|
| `/kube-workload/pods` | Pods | Manage workload instances | Namespace |
| `/kube-workload/deployments` | Deployments | Manage application deployments | Namespace |
| `/kube-workload/daemon-sets` | DaemonSets | Manage daemon set workloads | Namespace |
| `/kube-workload/stateful-sets` | StatefulSets | Manage stateful applications | Namespace |
| `/kube-workload/replica-sets` | ReplicaSets | Manage replica set workloads | Namespace |
| `/kube-workload/replication-controllers` | Replication Controllers | Legacy workload management | Namespace |
| `/kube-workload/jobs` | Jobs | Manage batch jobs | Namespace |
| `/kube-workload/cron-jobs` | CronJobs | Manage scheduled jobs | Namespace |

## Configuration

| Path | Header | Description | Scoping |
|------|--------|-------------|---------|
| `/kube-config/config-maps` | Config Maps | Manage configuration data | Namespace |
| `/kube-config/secrets` | Secrets | Manage sensitive information | Namespace |
| `/kube-config/resource-quotas` | Resource Quotas | Manage resource limits | Namespace |
| `/kube-config/limit-ranges` | Limit Ranges | Manage container resource limits | Namespace |
| `/kube-config/hpa` | HPA | Horizontal Pod Autoscalers | Namespace |
| `/kube-config/pod-disruption-budgets` | PDBs | Pod Disruption Budgets | Namespace |
| `/kube-config/priority-classes` | Priority Classes | Cluster-wide priority scheduling | Cluster-wide |
| `/kube-config/runtime-classes` | Runtime Classes | Cluster-wide container runtime configurations | Cluster-wide |
| `/kube-config/leases` | Leases | Distributed coordination and locking | Namespace |
| `/kube-config/mutating-webhooks` | Mutating Webhooks | Cluster-wide mutation configurations | Cluster-wide |
| `/kube-config/validating-webhooks` | Validating Webhooks | Cluster-wide validation configurations | Cluster-wide |

## Network

| Path | Header | Description | Scoping |
|------|--------|-------------|---------|
| `/kube-network/services` | Services | Manage networking endpoints | Namespace |
| `/kube-network/endpoints` | Endpoints | Manage service endpoints | Namespace |
| `/kube-network/ingresses` | Ingresses | Manage external access | Namespace |
| `/kube-network/ingress-classes` | Ingress Classes | Manage ingress controllers | Cluster-wide |
| `/kube-network/network-policies` | Network Policies | Manage network security policies | Namespace |
| `/kube-network/port-forwarding` | Port Forwarding | Manage active port forwards | Namespace |

## Storage

| Path | Header | Description | Scoping |
|------|--------|-------------|---------|
| `/kube-storage/persistent-volume-claims` | Persistent Volume Claims | Manage storage requests | Namespace |
| `/kube-storage/persistent-volumes` | Persistent Volumes | Manage cluster storage volumes | Cluster-wide |
| `/kube-storage/storage-classes` | Storage Classes | Manage storage provisioning | Cluster-wide |

## Access Control

| Path | Header | Description | Scoping |
|------|--------|-------------|---------|
| `/kube-access/service-accounts` | Service Accounts | Manage identity for processes | Namespace |
| `/kube-access/cluster-roles` | Cluster Roles | Manage cluster-wide permissions | Cluster-wide |
| `/kube-access/roles` | Roles | Manage namespace permissions | Namespace |
| `/kube-access/cluster-role-bindings` | Cluster Role Bindings | Manage cluster-wide role assignments | Cluster-wide |
| `/kube-access/role-bindings` | Role Bindings | Manage namespace role assignments | Namespace |

## System & Settings

| Path | Header | Description |
|------|--------|-------------|
| `/kube-events` | Events | Cluster events and alerts |
| `/settings/gitlab` | GitLab Settings | Configure GitLab integration |
| `/settings/clusters` | Cluster Settings | Manage connected clusters |
| `/settings/audit-logs` | Audit Logs | View system audit logs |

## Utilities

| Path | Description |
|------|-------------|
| `/exec` | Terminal session for container interaction |
| `/login` | User authentication page |
