# Resource Management

Cloud Sentinel K8s offers comprehensive tools for managing your Kubernetes resources directly from the dashboard. You can view detailed information, perform operations, and monitor the status of your workloads.

## Resource Operations

For resources like Deployments, StatefulSets, and DaemonSets, you can perform common operations directly from the resource detail view.

### Scaling

To scale a resource (e.g., a Deployment):
1. Navigate to the resource detail page.
2. Click the **Scale** button in the top right corner.
3. Adjust the number of replicas using the input field or the `+`/`-` buttons.
4. Click **Scale** to apply the changes.

### Restarting

You can trigger a rolling restart of your workloads:
1. Click the **Restart** button in the top right corner.
2. Confirm the action in the dialog.

This will update the resource's pod template with an annotation (e.g., `cloud-sentinel-k8s.kubernetes.io/restartedAt`) to trigger a rollout.

### Deleting

To delete a resource:
1. Click the **Delete** button (red trash icon) in the top right corner.
2. Confirm the deletion in the warning dialog.

::: danger
This action is irreversible. Please proceed with caution.
:::

## Live YAML Editing

Cloud Sentinel K8s includes a built-in YAML editor with syntax highlighting and validation.

1. Navigate to the **YAML** tab in the resource detail view.
2. Edit the configuration directly in the editor.
3. Click **Save** to apply the changes to the cluster.

The editor will validate your YAML before saving, helping to prevent configuration errors.

## Detailed Views

The resource detail page provides several tabs to help you analyze and troubleshoot your resources:

### Overview
Displays the status, replicas, selector, labels, annotations, containers, and conditions of the resource. You can also view init containers and ephemeral containers here.

### Pods
Lists the pods managed by the resource (e.g., pods belonging to a Deployment). You can view their status, IP, node, and age.

### Logs
Stream real-time logs from the pods associated with the resource. You can:
- Select specific pods and containers.
- Filter logs by text.
- Enable auto-scrolling.

### Terminal
Open a web-based terminal directly into the containers of the associated pods. This is useful for debugging and running quick commands inside the container environment.

### Monitor
View real-time resource usage metrics (CPU and Memory) for the pods associated with the resource. This requires Prometheus to be configured.

### History
For resources that support it (like Deployments), the **History** tab shows the revision history, allowing you to track changes over time.

### Events
Lists Kubernetes events related to the resource, helping you identify issues like scheduling failures, image pull errors, or health check failures.

### Related Resources
Visualize and navigate to resources related to the current one. For example, from a Service, you can see the Pods it selects.
