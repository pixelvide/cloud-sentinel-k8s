# Frequently Asked Questions (FAQ)



## Permission Issues

If you encounter an error message like the following when accessing resources:

```txt
User admin does not have permission to get configmaps in namespace cloud-sentinel-k8s in cluster in-cluster
```

This means that user `admin` does not have permission to access `configmaps` resources in the `cloud-sentinel-k8s` namespace.

You need to refer to the [RBAC Configuration Guide](./config/rbac-config) to configure user permissions.

## Managed Kubernetes Cluster Connection Issues

If you're using a managed Kubernetes cluster (AKS, EKS, GKE, etc.) and encounter authentication errors when adding the cluster to Cloud Sentinel K8s, this is usually because the default kubeconfig uses `exec` plugins that require CLI tools (like `kubelogin`, `aws`, `gcloud`, or `glab`).

While Cloud Sentinel K8s runs as a server-side application, it now supports native authentication for **AWS EKS** and **GitLab Agent** by configuring your credentials in the **Settings** page.

For other providers (like AKS or GKE), or as an alternative, you should use Service Account token-based authentication.

Please refer to the [Managed Kubernetes Cluster Configuration Guide](./config/managed-k8s-auth) for detailed instructions on both methods.

## SQLite with hostPath Storage

If you're using SQLite as the database and encountering an "out of memory" error when using `hostPath` for persistent storage:

```txt
panic: failed to connect database: unable to open database file: out of memory (14)
```

This issue is related to the pure Go SQLite driver used by Cloud Sentinel K8s (to avoid CGO dependencies). The driver has limitations when accessing database files on certain storage backends.

**Solution**: Add SQLite connection options to improve compatibility with hostPath storage. In your Helm values, set:

```yaml
db:
  sqlite:
    options: "_journal_mode=WAL&_busy_timeout=5000"
```

These options enable Write-Ahead Logging (WAL) mode and increase the busy timeout, which resolves most hostPath compatibility issues.

**Recommended for Production**: For production deployments requiring persistent storage, use MySQL or PostgreSQL instead of SQLite. These databases are better suited for containerized environments and persistent storage scenarios.

For more details, see [Issue #204](https://github.com/pixelvide/cloud-sentinel-k8s/issues/204).

## AI Features

### Which AI providers are supported?
Cloud Sentinel K8s currently supports Google Gemini and OpenAI. More providers can be added through the AI administration interface if they follow a compatible API structure.

### How do I configure my own API key?
If your administrator allows it, you can provide your own API key in **Settings > AI Configuration**. This allows you to use your personal quota and specific models.

### Is my cluster data sent to AI providers?
Data is only sent to the configured AI provider when you explicitly interact with the AI assistant (e.g., asking a question about a resource). Only the necessary context (like resource YAML or logs) is sent to provide accurate answers.

### How can I disable AI chat?
Administrators can disable AI chat for individual users in the **User Management** settings.

## How to Change Font

By default, Cloud Sentinel K8s provides three fonts: system default, `Maple Mono`, and `JetBrains Mono`.

If you want to use a different font, you need to build the project yourself.

Build cloud-sentinel-k8s with make and change the font in `./ui/src/index.css`:

```css
@font-face {
  font-family: "Maple Mono";
  font-style: normal;
  font-display: swap;
  font-weight: 400;
  src: url(https://cdn.jsdelivr.net/fontsource/fonts/maple-mono@latest/latin-400-normal.woff2)
      format("woff2"), url(https://cdn.jsdelivr.net/fontsource/fonts/maple-mono@latest/latin-400-normal.woff)
      format("woff");
}

body {
  font-family: "Maple Mono", var(--font-sans);
}
```

## How Can I Contribute to Cloud Sentinel K8s?

We welcome contributions! You can:

- Report bugs and feature requests on [GitHub Issues](https://github.com/pixelvide/cloud-sentinel-k8s/issues)
- Submit pull requests
- Improve documentation
- Share feedback and use cases

## Where Can I Get Help?

You can get support through:

- [GitHub Issues](https://github.com/pixelvide/cloud-sentinel-k8s/issues) for bug reports and feature requests
- [Slack Community](https://join.slack.com/t/cloud-sentinel-k8s-dashboard/shared_invite/zt-3amy6f23n-~QZYoricIOAYtgLs_JagEw) for questions and community support

---

**Didn't find what you're looking for?** Feel free to [open an issue](https://github.com/pixelvide/cloud-sentinel-k8s/issues/new) on GitHub or start a [discussion](https://github.com/pixelvide/cloud-sentinel-k8s/discussions).
