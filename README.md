# Cloud Sentinel - Kubernetes Dashboard

A modern, read-optimized Kubernetes dashboard built with Next.js and Go.

## Features

- **Read-Only Dashboard**: Secure viewing of Kubernetes resources.
- **Multi-Context Support**: Switch between multiple clusters easily.
- **Resource Views**: Pods, Deployments, Services, Ingresses, Events, and more.
- **Interactive Terminal**: Exec into pods directly from the browser.
- **OIDC Authentication**: Secure login with generic OIDC providers.

## Prerequisites

- **Docker & Docker Compose**: For containerized deployment.
- **PostgreSQL**: An external database (or local instance) is required.

## Quick Start

1.  **Clone the repository**:
    ```bash
    git clone <repository-url>
    cd cloud-sentinel
    ```

2.  **Configure Environment Variables**:
    Create a `.env` file in the root directory or set environment variables in your shell. The backend relies on these to connect to the database and OIDC provider.

    **Database Configuration:**
    ```env
    DB_HOST=localhost          # Hostname/IP of your Postgres DB
    DB_PORT=5432               # Port (default: 5432)
    DB_USER=postgres           # Database username
    DB_PASSWORD=secret         # Database password
    DB_NAME=cloud_sentinel     # Database name
    DB_SSLMODE=disable         # SSL Mode: disable, require, verify-ca, etc. (default: require)
    DB_TIMEZONE=UTC            # Timezone for sessions (default: UTC). Note: DB server must support it.
    ```

    **OIDC Configuration:**
    ```env
    OIDC_ISSUER=https://accounts.google.com  # OIDC Provider URL
    OIDC_CLIENT_ID=<your-client-id>
    OIDC_CLIENT_SECRET=<your-client-secret>
    # OIDC_REDIRECT_URL is automatically derived from FRONTEND_URL + /api/v1/auth/callback
    FRONTEND_URL=http://localhost:3000       # URL where frontend is accessible (default: http://localhost:3000)
    ```

3.  **Run with Docker Compose**:
    ```bash
    docker compose up -d --build
    ```

    This will start:
    - **Frontend**: Available at `http://localhost:3000`
    - **Backend**: Available at `http://localhost:8080` (API)

## Deployment Note

- **Frontend-Backend Communication**:
    - The application is configured with **Next.js rewrites** to proxy requests from `/api/*` to the backend service internally.
    - This means API calls from the browser go to `http://localhost:3000/api/...`, and Next.js forwards them to the backend container.
    - **The backend port (8080) is NOT exposed publicly by default** to improve security.
    
    **Important: OIDC Redirects**
    - The `redirect_uri` sent to your OIDC provider is automatically constructed as:
      `{FRONTEND_URL}/api/v1/auth/callback`
    - Ensure your OIDC provider (e.g. Google Console) has this URL whitelisted.

## Development

- **Backend**: Go (Gin framework)
- **Frontend**: TypeScript, Next.js, Tailwind CSS, Shadcn UI
