# Contributing to Cloud Sentinel

Thank you for your interest in contributing to Cloud Sentinel! We welcome contributions from everyone. This document outlines the process for setting up your environment and submitting changes.

## 🛠️ Tech Stack
- **UI**: React + Vite, TypeScript, Tailwind CSS, Shadcn UI.
- **Backend**: Go (Golang) 1.25+, Gin Gonic framework.
- **Database**: PostgreSQL.
- **Infrastructure**: Kubernetes, GitLab CI/CD.

## 🚀 Development Setup

### Prerequisites
- Node.js 18+ and `npm`.
- Go 1.22+.
- Docker & Docker Compose (for DB).
- A running Kubernetes cluster (Minikube, Kind, or remote) with a valid `~/.kube/config`.

### 1. Database Setup
Start the local PostgreSQL instance:
```bash
docker-compose up -d db
```

### 2. Backend Setup
Navigate to the backend directory and run the server:
```bash
cd backend
go mod download
# Ensure you have a valid KUBECONFIG
export KUBECONFIG=~/.kube/config
go run main.go
```
The server will start on `http://localhost:8080`.

### 3. UI Setup
Navigate to the ui directory and start the dev server:
```bash
cd ui
pnpm install
pnpm dev
```
The application will be available at `http://localhost:3000`.

## 📐 Architecture
Please refer to [ARCHITECTURE.md](./ARCHITECTURE.md) for a high-level overview of the system components and data flow.

## 🤝 Code Standards

- **Go**: We adhere to standard Go formatting. Please run `go fmt ./...` before committing.
- **TypeScript**: We use ESLint. Run `pnpm run lint` to check for issues.
- **Commits**: Use descriptive commit messages.

## 🔀 Submission Guidelines (GitLab)

1.  **Fork** the project in GitLab.
2.  Create a **feature branch** (`git checkout -b feature/amazing-feature`).
3.  Commit your changes.
4.  Push to the branch (`git push origin feature/amazing-feature`).
5.  Open a **Merge Request** (MR) targeting the `main` branch.
6.  Ensure your MR has a clear description of the changes and screenshots if UI is affected.

## 📦 Release Management

Cloud Sentinel follows [Semantic Versioning](https://semver.org/).
- **Automated Releases**: We use **Release-Please** to automate changelog generation and versioning.
- **Conventional Commits**: Please use [Conventional Commits](https://www.conventionalcommits.org/) for your commit messages (e.g., `feat:`, `fix:`, `chore:`, `docs:`) to ensure they are correctly picked up by the release automation.
- **Deployment**: Official images are built and pushed to GHCR on every tagged release.

## 📜 Code of Conduct

Please review our [Code of Conduct](./CODE_OF_CONDUCT.md) before contributing.
