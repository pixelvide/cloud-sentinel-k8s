# Backend API Documentation

This document provides a comprehensive list of all backend API endpoints for the Cloud Sentinel application.

## Base URL
All API requests (except health check) are prefixed with `/api/v1`.
- Production/Development: `http://<backend-host>:8080/api/v1`

## Authentication
All endpoints under `/api/v1` require authentication via a JWT token sent in the `Authorization` header:
`Authorization: Bearer <token>`

---

## Public Endpoints

### Health Check
- **URL**: `/health`
- **Method**: `GET`
- **Description**: Verify the backend is running.
- **Response**: `{"status": "ok"}`

### Login (OIDC)
- **URL**: `/api/v1/auth/login`
- **Method**: `GET`
- **Description**: Initiates OIDC login flow. Redirects to the OIDC provider (e.g., Google, Okta).

### OIDC Callback
- **URL**: `/api/v1/auth/callback`
- **Method**: `GET`
- **Description**: Handles the callback from the OIDC provider.

### Logout
- **URL**: `/api/v1/auth/logout`
- **Method**: `GET`
- **Description**: Logs out the user.

---

## User & Profile

### Get Current User
- **URL**: `/me`
- **Method**: `GET`
- **Description**: Retrieves the profile and metadata for the currently authenticated user.

---

## Settings

### GitLab Configurations

#### List GitLab Configs
- **URL**: `/settings/gitlab`
- **Method**: `GET`
- **Description**: Returns all GitLab configurations for the user.

#### Create GitLab Config
- **URL**: `/settings/gitlab`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "gitlab_host": "gitlab.com",
    "token": "your-personal-access-token",
    "is_https": true
  }
  ```
- **Description**: Adds a new GitLab instance configuration.

#### Update GitLab Config
- **URL**: `/settings/gitlab/:id`
- **Method**: `PUT`
- **Body**:
  ```json
  {
    "token": "new-token"
  }
  ```
- **Description**: Updates the access token for a GitLab configuration.

#### Delete GitLab Config
- **URL**: `/settings/gitlab/:id`
- **Method**: `DELETE`
- **Description**: Removes a GitLab configuration.

#### Validate GitLab Config
- **URL**: `/settings/gitlab/:id/validate`
- **Method**: `POST`
- **Description**: Uses the `glab` CLI to verify credentials.

---

### GitLab Agent Configurations

#### List Agent Configs
- **URL**: `/settings/gitlab/agents`
- **Method**: `GET`
- **Description**: Lists all GitLab Agent configurations.

#### Create Agent Config
- **URL**: `/settings/gitlab/agents`
- **Method**: `POST`
- **Description**: Creates a new agent configuration.

#### Delete Agent Config
- **URL**: `/settings/gitlab/agents/:id`
- **Method**: `DELETE`
- **Description**: Removes an agent configuration.

#### Configure Agent
- **URL**: `/settings/gitlab/agents/:id/configure`
- **Method**: `POST`
- **Description**: Triggers the configuration/installation of the GitLab Agent.

---

### Kubernetes Configuration

#### Get KubeConfig
- **URL**: `/settings/kube`
- **Method**: `GET`
- **Description**: Retrieves the current kubeconfig (masked).

#### Update KubeConfig
- **URL**: `/settings/kube`
- **Method**: `POST`
- **Description**: Updates the backend's kubeconfig content.

#### Validate KubeConfig
- **URL**: `/settings/kube/validate`
- **Method**: `POST`
- **Description**: Validates the uploaded kubeconfig content.

#### Set Current Context
- **URL**: `/settings/kube/context`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "context": "my-context-name"
  }
  ```
- **Description**: Updates the `current-context` in the backend's kubeconfig.

---

### Context Mappings

#### List Mappings
- **URL**: `/settings/context-mappings`
- **Method**: `GET`

#### Upsert Mapping
- **URL**: `/settings/context-mappings`
- **Method**: `POST`
- **Description**: Creates or updates a mapping between a context and a friendly display name.

#### Delete Mapping
- **URL**: `/settings/context-mappings/:id`
- **Method**: `DELETE`

---

### Audit Logs
- **URL**: `/settings/audit-logs`
- **Method**: `GET`
- **Description**: Retrieves the audit logs for the current user.

---

## Kubernetes Resources

### Resource Listing
The following endpoints return a list of Kubernetes resources from the currently selected context.
- **Base Path**: `/kube/`
- **Method**: `GET`

| Resource | Endpoint |
| :--- | :--- |
| **Contexts** | `/kube/contexts` |
| **Namespaces** | `/kube/namespaces` |
| **Nodes** | `/kube/nodes` |
| **Pods** | `/kube/pods` |
| **Services** | `/kube/services` |
| **Ingresses** | `/kube/ingresses` |
| **Deployments** | `/kube/deployments` |
| **ReplicaSets** | `/kube/replica-sets` |
| **ReplicationControllers** | `/kube/replication-controllers` |
| **DaemonSets** | `/kube/daemon-sets` |
| **StatefulSets** | `/kube/stateful-sets` |
| **Jobs** | `/kube/jobs` |
| **CronJobs** | `/kube/cron-jobs` |
| **ConfigMaps** | `/kube/config-maps` |
| **Secrets** | `/kube/secrets` |
| **ResourceQuotas** | `/kube/resource-quotas` |
| **LimitRanges** | `/kube/limit-ranges` |
| **HPAs** | `/kube/hpas` |
| **PDBs** | `/kube/pdbs` |
| **PriorityClasses** | `/kube/priority-classes` |
| **RuntimeClasses** | `/kube/runtime-classes` |
| **Leases** | `/kube/leases` |
| **MutatingWebhooks** | `/kube/mutating-webhooks` |
| **ValidatingWebhooks** | `/kube/validating-webhooks` |
| **Endpoints** | `/kube/endpoints` |
| **IngressClasses** | `/kube/ingress-classes` |
| **NetworkPolicies** | `/kube/network-policies` |
| **PortForwards** | `/kube/port-forwards` |
| **PVCs** | `/kube/pvcs` |
| **PVs** | `/kube/pvs` |
| **StorageClasses** | `/kube/storage-classes` |
| **ServiceAccounts** | `/kube/service-accounts` |
| **ClusterRoles** | `/kube/cluster-roles` |
| **Roles** | `/kube/roles` |
| **ClusterRoleBindings** | `/kube/cluster-role-bindings` |
| **RoleBindings** | `/kube/role-bindings` |
| **Events** | `/kube/events` |
| **Scopes** | `/kube/scopes` |

### Resource Details
- **URL**: `/kube/resource`
- **Method**: `GET`
- **Query Params**:
  - `namespace`: Resource namespace
  - `kind`: Resource kind (e.g., Pod, Service)
  - `name`: Resource name
- **Description**: Returns full JSON details and events for a specific resource.

### Dashboard
- **URL**: `/kube/dashboard`
- **Method**: `GET`
- **Description**: Returns a summary of cluster health and resource counts.

---

## Interactive / Real-time (WebSockets)

### Container Exec
- **URL**: `/kube/exec`
- **Method**: `WS`
- **Query Params**:
  - `namespace`: Pod namespace
  - `pod`: Pod name
  - `container`: Container name
- **Description**: Establishes a WebSocket connection for interactive terminal access.

### Container Logs
- **URL**: `/kube/logs`
- **Method**: `WS`
- **Query Params**:
  - `namespace`: Pod namespace
  - `pod`: Pod name (comma-separated for multi-pod)
  - `container`: Container name (optional)
- **Description**: Streams logs via WebSocket. Supports multiple pods and heartbeat (Ping/Pong).
