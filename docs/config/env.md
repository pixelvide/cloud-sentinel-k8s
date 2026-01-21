# Environment Variables

Cloud Sentinel K8s supports several environment variables to customize its behavior.

## Core Configuration

- **PORT**: Port on which Cloud Sentinel K8s runs, default value is `8080`.
- **HOST**: Used for generating OAuth 2.0 authorization callback addresses. Usually detected from request headers, but can be set manually (e.g., `https://cloud-sentinel-k8s.example.com`).
- **CLOUD_SENTINEL_K8S_BASE**: Base path for the application. If set to `/cloud-sentinel-k8s`, the application will be accessible at `domain.com/cloud-sentinel-k8s`.
- **JWT_SECRET**: Secret key used for signing and verifying JWT tokens. **Must be changed in production!**
- **CLOUD_SENTINEL_K8S_ENCRYPT_KEY**: Secret key used for encrypting sensitive data (user passwords, tokens, kubeconfigs). **Must be changed in production!**

## Database Configuration

- **DB_TYPE**: The type of database to use. Supported: `sqlite`, `mysql`, `postgres`. Default is `sqlite`.
- **DB_DSN**: Database connection DSN.
    - For `sqlite`: Path to the database file (e.g., `data/cloud-sentinel-k8s.db`). Default is `dev.db`.
    - For `mysql`/`postgres`: Standard connection string (e.g., `user:pass@tcp(host:3306)/dbname`).

## Authentication & Authorization

- **CLOUD_SENTINEL_K8S_USERNAME**: Set the initial administrator username during bootstrap.
- **CLOUD_SENTINEL_K8S_PASSWORD**: Set the initial administrator password during bootstrap.
- **KUBECONFIG**: Path to the initial Kubernetes configuration file. Default is `~/.kube/config`. Clusters from this config will be discovered and imported on the first run.

## Third-party Integrations

- **GITLAB_HOSTS**: Comma-separated list of GitLab hosts to pre-configure (e.g., `https://gitlab.com,http://my-gitlab.local`). These will be seeded into the database on startup.

## Specialized Settings

- **NODE_TERMINAL_IMAGE**: Docker image used for the Node Terminal Agent. Default is `busybox:latest`.
- **DISABLE_GZIP**: Disable GZIP compression for API responses. Default is `true`.
- **DISABLE_VERSION_CHECK**: Disable the automatic check for new application versions. Default is `false`.
- **DISABLE_CACHE**: Disable the Kubernetes client-side cache. Default is `false`.
