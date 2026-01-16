# Architecture Overview

**Cloud Sentinel** is a modern, unified Kubernetes dashboard designed for DevOps engineers to manage multiple clusters, visualize workloads, and integrate closely with GitLab for GitOps workflows.

## System Components

The system consists of three main components:

1.  **Frontend**: A Next.js (React) application.
2.  **Backend**: A Go (Golang) REST API.
3.  **Database**: PostgreSQL for persistent storage of user configs and audit logs.

```mermaid
graph TD
    User["User / Browser"] -->|HTTPS| Frontend["Frontend (Next.js)"]
    Frontend -->|API Calls / WS| Backend["Backend API (Go)"]
    Backend -->|SQL| DB[(PostgreSQL)]
    Backend -->|K8s API| K8s["Kubernetes Cluster(s)"]
    Backend -->|API| GitLab[GitLab API]
```

## 1. Frontend Layer
- **Framework**: Next.js 14+ (App Router).
- **Styling**: Tailwind CSS with Shadcn UI components.
- **State Management**: React Hooks & Context.
- **Communication**:
    - **REST**: Standard API calls for fetching resource lists.
    - **WebSockets**: Real-time streaming for Logs (`/api/v1/kube/logs`) and Terminal Exec (`/api/v1/kube/exec`).

## 2. Backend Layer
- **Framework**: Gin Gonic (High-performance HTTP web framework).
- **Language**: Go 1.22+.
- **Key Modules**:
    - `api/`: REST handlers for Kubernetes resources (Pods, Deployments, CRDs, etc.).
    - `auth/`: OIDC authentication flow and JWT token generation.
    - `db/`: Database connection and ORM models using GORM.
    - `k8s/`: Wrapper around `client-go` for interacting with K8s clusters.

## 3. Data Storage (PostgreSQL)
The application handles minimal state, delegating most source-of-truth to Kubernetes itself. However, it persists:
- **Users**: User profiles and authentication context.
- **Audit Logs**: Records of critical actions (e.g., Delete Resource, Cordon Node).
- **Cluster Mappings**: Context configuration for multi-cluster management.
- **GitLab Config**: Access tokens and project mappings for GitOps integration.

## 4. Kubernetes Integration
- **Client**: Uses official `k8s.io/client-go`.
- **Dynamic Client**: Utilized for CRDs and resources where the schema is not known at compile time.
- **Multi-Context**: Supports switching between different kubeconfig contexts dynamically per request.

## 5. Security Architecture
- **Authentication**: OIDC (OpenID Connect) integration with providers like GitLab, Google, etc.
- **Authorization**:
    - Backend validates valid JWT in `Authorization` header.
    - Kubernetes RBAC is respected by using the underlying kubeconfig credentials (impersonation or direct use depending on deployment).
- **Sensitive Data**: Kubeconfigs and Secrets are never exposed to the frontend; the backend acts as a secure proxy.

## 6. Resource Action Flow
When a user triggers an action (e.g., Restart Deployment or Suspend CronJob):
1.  **Frontend**: Sends a POST/PATCH request to the specific action endpoint (e.g., `/api/v1/kube/workloads/deployments/restart`).
2.  **Backend**:
    - Validates the request and user permissions.
    - Uses the `dynamic client` or `client-go` to apply the transformation (e.g., updating a `rollout.kubernetes.io/restartedAt` annotation).
    - Records the action in the **Audit Log** (PostgreSQL).
3.  **Kubernetes**: Reconciles the resource based on the updated specification.

## 7. CI/CD & Delivery
The project follows an automated delivery pipeline:
- **Versioning**: Managed by `release-please` based on commit types (feat, fix, etc.).
- **Build**: GitHub Actions triggers on new tags to build multi-arch Docker images (`linux/amd64`, `linux/arm64`).
- **Registry**: Images are stored in **GitHub Container Registry (GHCR)**.
- **Environment**: Configuration is managed via environment variables and `.env` files, ensuring secrets are handled securely outside the code.

