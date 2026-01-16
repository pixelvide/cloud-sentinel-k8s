# Cloud Sentinel - Kubernetes Dashboard

A modern, read-optimized Kubernetes dashboard built with Next.js and Go.

## Project Documentation

- [**Architecture**](./ARCHITECTURE.md): High-level system design and component interaction.
- [**Future Roadmap**](./FUTURE_ROADMAP.md): Planned features and improvements.
- [**Contributing**](./CONTRIBUTING.md): Guidelines for contributing to the project.
- [**Code of Conduct**](./CODE_OF_CONDUCT.md): Community standards and expectations.
- [**Development Guide**](./DEVELOPMENT.md): Instructions for local development and setup.

## Features

### Resource Management
- **Workloads**: Deep insights into Pods, Deployments, ReplicaSets, StatefulSets, DaemonSets, Jobs, and CronJobs.
- **Cluster Resources**: Manage Nodes, Namespaces, StorageClasses, PVs/PVCs, and ClusterRoles.
- **Configuration**: View and manage ConfigMaps, Secrets, RBAC (Roles, ServiceAccounts), and Network Policies.
- **CRDs**: robust support for Custom Resource Definitions with formatted views.

### Actionable & Interactive
- **Resource Actions**: Delete resources, suspend/resume CronJobs, and drain nodes directly from the UI.
- **Helmet Management**: Full lifecycle management for Helm releases including list, filter, history, and rollback capabilities.
- **Terminal & Logs**: Secure, integrated terminal access to pods and real-time log streaming with search.
- **Audit Logging**: Comprehensive audit trails for all user actions (login, delete, update).

### Enhanced Visualization
- **Rich Details**: Right-side details panel with JSON/YAML views, live Events, and deep property inspection (Affinity, Tolerations, Conditions).
- **Resource Relations**: Automatically discover and list related resources (e.g., Pods for a Deployment).
- **Multi-Context**: Seamlessly switch between multiple Kubernetes clusters.

## Prerequisites

- **Docker & Docker Compose**: For containerized deployment.
- **PostgreSQL**: An external database or local instance (required for audit logs and user data).

## Quick Start

1.  **Clone the repository**:
    ```bash
    git clone <repository-url>
    cd cloud-sentinel
    ```

2.  **Configure Environment Variables**:
    Copy the example environment file:
    ```bash
    cp .env.example .env
    ```
    
    Edit `.env` to configure your database and OIDC settings.
    
    **Database Configuration:**
    ```env
    DB_HOST=localhost          # Hostname/IP of your Postgres DB
    DB_PORT=5432               # Port (default: 5432)
    DB_USER=postgres           # Database username
    DB_PASSWORD=secret         # Database password
    DB_NAME=cloud_sentinel     # Database name
    DB_SSLMODE=disable         # Mode: disable, require, verify-ca
    ```

    **OIDC Configuration:**
    ```env
    OIDC_ISSUER=https://accounts.google.com  # OIDC Provider URL
    OIDC_CLIENT_ID=<your-client-id>
    OIDC_CLIENT_SECRET=<your-client-secret>
    FRONTEND_URL=http://localhost:3000       
    ```

3.  **Run with Docker Compose**:
    ```bash
    docker compose up -d --build
    ```

    This will start:
    - **Frontend**: `http://localhost:3000`
    - **Backend**: `http://localhost:8080` (Internal API)

## Deployment Note

- **Frontend-Backend Communication**:
    - The application uses **Next.js rewrites** to proxy requests from `/api/*` to the backend service.
    - **Security**: The backend port (8080) is NOT exposed publicly by default.
    
    **OIDC Redirects**:
    - The `redirect_uri` is constructed as `{FRONTEND_URL}/api/v1/auth/callback`.
    - Whitelist this URL in your OIDC provider settings.

## Development

- **Backend**: Go (Gin), Client-go, Dynamic Client for CRDs
- **Frontend**: TypeScript, Next.js 14, Tailwind CSS, Shadcn UI
- **Tools**: `kubectl`, `helm`, `fd`, `ripgrep` recommended for dev workflow.
