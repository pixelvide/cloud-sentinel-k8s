# Deployment Guide (Kubernetes)

This guide explains how to deploy Cloud Sentinel to a Kubernetes cluster.

## Prerequisites

- A Kubernetes cluster (v1.20+)
- `kubectl` installed and configured
- `kustomize` (built into `kubectl` via `kubectl kustomize` or `kubectl apply -k`)
- Images published to GHCR (automated via CI/CD)

## 1. Manual Resource Creation

Before applying the deployments, you must create the necessary ConfigMap and Secret for environment variables.

### Create ConfigMap
Replace the values below with your actual configuration.

```bash
kubectl create configmap cloud-sentinel-config \
  --from-literal=DB_TYPE=postgres \
  --from-literal=OIDC_ISSUER=https://your-oidc-issuer \
  --from-literal=OIDC_CLIENT_ID=your-client-id \
  --from-literal=FRONTEND_URL=https://cloud-sentinel.yourdomain.com
```

### Create Secrets
Store sensitive information in a Secret.

```bash
kubectl create secret generic cloud-sentinel-secrets \
  --from-literal=DB_DSN="host=your-db-host user=your-db-user password=your-db-password dbname=cloud_sentinel port=5432 sslmode=require TimeZone=UTC" \
  --from-literal=OIDC_CLIENT_SECRET=your-oidc-secret
```

**Note**: The `DB_DSN` should be a complete database connection string. Examples:
- **PostgreSQL**: `host=db.example.com user=postgres password=secret dbname=cloud_sentinel port=5432 sslmode=require TimeZone=UTC`
- **MySQL**: `user:password@tcp(db.example.com:3306)/cloud_sentinel?charset=utf8mb4&parseTime=True&loc=UTC`
- **SQLite**: `file:/data/cloud_sentinel.db`

### GHCR Authentication (Optional)
Since GHCR is public, you do not need an image pull secret. However, if you decide to make it private in the future, you can create one:

```bash
kubectl create secret docker-registry ghcr-pull-secret \
  --docker-server=ghcr.io \
  --docker-username=<GITHUB_USERNAME> \
  --docker-password=<GITHUB_TOKEN>
```
*Note: If you use this, update the deployments in `k8s/*.yaml` to include `imagePullSecrets`.*

## 2. Deploying the Application

Use Kustomize to deploy the service.

### Preview Manifests
```bash
kubectl kustomize k8s/
```

### Apply to Cluster
```bash
kubectl apply -k k8s/
```
## 5. Deploying to a Custom Namespace

If you want to deploy Cloud Sentinel to a specific namespace (e.g., `production`), follow these steps:

### Option A: Using Kustomization (Recommended)
Add the `namespace` field to your `k8s/kustomization.yaml`:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: cloud-sentinel-ns
...
```

Then create the namespace and apply:
```bash
kubectl create namespace cloud-sentinel-ns
kubectl apply -k k8s/
```

### Option B: Using the CLI
You can also specify the namespace directly in the apply command, but ensure your ConfigMap and Secret are created in that same namespace first.

```bash
kubectl apply -k k8s/ --namespace cloud-sentinel-ns
```

## 3. Post-Deployment

- **Backend**: Accessible internally within the cluster at `http://cloud-sentinel-backend:8080`.
- **Frontend**: Exposed via a `LoadBalancer` service. You can find the external IP using:
  ```bash
  kubectl get service cloud-sentinel-ui
  ```

## 4. Updates and Versioning

The image versions are managed in `k8s/kustomization.yaml`. 

The CI/CD pipeline (via `release-please`) automatically updates this file when a new release is created. To manually update the image versions, modify the `newTag` fields in `k8s/kustomization.yaml` and re-apply:

```bash
kubectl apply -k k8s/
```
