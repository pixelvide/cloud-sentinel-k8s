# Development Workflow

This document outlines the standard development workflow for Cloud Sentinel.

## Docker-based Development

As of current development sessions, we use **Docker Compose** as the primary method for running and building the application. Avoid using local `npm` or `go` commands directly unless troubleshooting specific local environment issues.

### Key Commands

- **Build and Start (Full Stack)**:
  ```bash
  docker compose up -d --build
  ```

- **Stop Services**:
  ```bash
  docker compose down
  ```

- **View Logs**:
  ```bash
  docker compose logs -f
  ```

- **Restart Backend**:
  ```bash
  docker compose up -d --build backend
  ```

### Development Environment

- **UI**: Accessible at `http://localhost:3000`
- **Backend**: Served directly from the backend service.
- **Configuration**: Managed via `.env` in the root directory.

### Environment Configuration

The application requires the following environment variables (see `.env.example`):

**Database**:
```env
DB_TYPE=sqlite    # Default: sqlite. Options: postgres, mysql, sqlite
DB_DSN=dev.db     # Default: dev.db
```

**Authentication**:
```env
# Encryption & Security
CLOUD_SENTINAL_ENCRYPT_KEY=your-secure-encryption-key
JWT_SECRET=your-secure-jwt-secret

# OIDC (Optional)
OIDC_ISSUER=https://your-oidc-provider.com
OIDC_CLIENT_ID=your-client-id
OIDC_CLIENT_SECRET=your-client-secret
FRONTEND_URL=http://localhost:3000
```

**Optional - Subpath Hosting**:
```env
CLOUD_SENTINEL_K8S_BASE=/cloud-sentinel
```

### Local Backend Development

For local Go development without Docker:

1. **Install Dependencies**:
   ```bash
   cd backend
   go mod download
   ```

2. **Run the Backend**:
   ```bash
   go run .
   ```

3. **Build the Backend**:
   ```bash
   go build -o server
   ```

### Local Frontend Development

For local React development without Docker:

1. **Install Dependencies**:
   ```bash
   cd ui
   pnpm install
   ```

2. **Run Dev Server**:
   ```bash
   pnpm run dev
   ```

3. **Build for Production**:
   ```bash
   pnpm run build
   ```

### Session Persistence

When performing code changes, always use `docker compose up -d --build` to ensure changes are reflected in the running containers.
