package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/analyzer"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func (m *MCPServer) registerTools() {
	// 1. List Clusters
	m.server.AddTool(mcp.NewTool("list_clusters",
		mcp.WithDescription("List all managed Kubernetes clusters and their status"),
	), m.handleListClusters)

	// 2. List Resources
	m.server.AddTool(mcp.NewTool("list_resources",
		mcp.WithDescription("List Kubernetes resources (pods, nodes, deployments, etc.) in a specific cluster and namespace"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Name of the cluster")),
		mcp.WithString("resource", mcp.Required(), mcp.Description("Type of resource (e.g., pods, services, deployments)")),
		mcp.WithString("namespace", mcp.Description("Namespace (optional, defaults to all namespaces)")),
	), m.handleListResources)

	// 3. Get Resource YAML
	m.server.AddTool(mcp.NewTool("get_resource_yaml",
		mcp.WithDescription("Fetch the full YAML manifest of a specific Kubernetes resource"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Name of the cluster")),
		mcp.WithString("resource", mcp.Required(), mcp.Description("Type of resource")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name of the resource")),
	), m.handleGetResourceYAML)

	// 4. Get Pod Logs
	m.server.AddTool(mcp.NewTool("get_pod_logs",
		mcp.WithDescription("Fetch logs from a specific pod"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Name of the cluster")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Pod name")),
		mcp.WithNumber("tailLines", mcp.Description("Number of lines to tail (default 100)")),
	), m.handleGetPodLogs)

	// 5. Run Security Scan
	m.server.AddTool(mcp.NewTool("run_security_scan",
		mcp.WithDescription("Run a security analysis on a specific Kubernetes resource"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("Name of the cluster")),
		mcp.WithString("resource", mcp.Required(), mcp.Description("Type of resource")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name of the resource")),
	), m.handleRunSecurityScan)
}

func (m *MCPServer) handleListClusters(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	clusters := m.cm.GetActiveClusters()
	data, err := json.Marshal(clusters)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal clusters: %w", err)
	}
	return mcp.NewToolResultText(string(data)), nil
}

func (m *MCPServer) handleListResources(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	clusterName, _ := request.RequireString("cluster")
	resourceType, _ := request.RequireString("resource")
	namespace := request.GetString("namespace", "")

	cs, err := m.cm.GetCluster(clusterName)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("Error: cluster %s not found", clusterName)), err
	}

	gvr := getGVR(resourceType)
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gvr.Group,
		Version: gvr.Version,
		Kind:    guessKind(resourceType),
	})

	var listOpts []client.ListOption
	if namespace != "" {
		listOpts = append(listOpts, client.InNamespace(namespace))
	}

	if err := cs.K8sClient.List(ctx, list, listOpts...); err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("Error listing resources: %v", err)), nil
	}

	summary := []string{}
	for _, item := range list.Items {
		summary = append(summary, fmt.Sprintf("%s/%s", item.GetNamespace(), item.GetName()))
	}

	if len(summary) == 0 {
		return mcp.NewToolResultText("No resources found"), nil
	}

	return mcp.NewToolResultText(strings.Join(summary, "\n")), nil
}

func (m *MCPServer) handleGetResourceYAML(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	clusterName, _ := request.RequireString("cluster")
	resourceType, _ := request.RequireString("resource")
	namespace, _ := request.RequireString("namespace")
	name, _ := request.RequireString("name")

	cs, err := m.cm.GetCluster(clusterName)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("Error: cluster %s not found", clusterName)), err
	}

	gvr := getGVR(resourceType)
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gvr.Group,
		Version: gvr.Version,
		Kind:    guessKind(resourceType),
	})

	err = cs.K8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, obj)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("Error fetching resource: %v", err)), nil
	}

	y, err := yaml.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource to YAML: %w", err)
	}
	return mcp.NewToolResultText(string(y)), nil
}

func (m *MCPServer) handleGetPodLogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	clusterName, _ := request.RequireString("cluster")
	namespace, _ := request.RequireString("namespace")
	name, _ := request.RequireString("name")
	tailLines := int64(request.GetInt("tailLines", 100))

	cs, err := m.cm.GetCluster(clusterName)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("Error: cluster %s not found", clusterName)), err
	}

	opts := &corev1.PodLogOptions{
		TailLines: &tailLines,
	}

	req := cs.K8sClient.ClientSet.CoreV1().Pods(namespace).GetLogs(name, opts)
	podLogs, err := req.DoRaw(ctx)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("Error getting logs: %v", err)), nil
	}

	return mcp.NewToolResultText(string(podLogs)), nil
}

func (m *MCPServer) handleRunSecurityScan(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	clusterName, _ := request.RequireString("cluster")
	resourceType, _ := request.RequireString("resource")
	namespace, _ := request.RequireString("namespace")
	name, _ := request.RequireString("name")

	cs, err := m.cm.GetCluster(clusterName)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("Error: cluster %s not found", clusterName)), err
	}

	gvr := getGVR(resourceType)
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gvr.Group,
		Version: gvr.Version,
		Kind:    guessKind(resourceType),
	})

	err = cs.K8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, obj)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("Error fetching resource: %v", err)), nil
	}

	results := analyzer.Analyze(ctx, cs.K8sClient, obj)
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal security results: %w", err)
	}
	return mcp.NewToolResultText(string(data)), nil
}

// Helpers
func getGVR(resource string) schema.GroupVersionResource {
	switch strings.ToLower(resource) {
	case "pods", "pod":
		return schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	case "services", "service":
		return schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	case "deployments", "deployment":
		return schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	case "statefulsets", "statefulset":
		return schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}
	case "daemonsets", "daemonset":
		return schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}
	case "nodes", "node":
		return schema.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"}
	case "namespaces", "namespace":
		return schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
	default:
		return schema.GroupVersionResource{Group: "", Version: "v1", Resource: resource}
	}
}

func guessKind(resource string) string {
	switch strings.ToLower(resource) {
	case "pods", "pod":
		return "Pod"
	case "services", "service":
		return "Service"
	case "deployments", "deployment":
		return "Deployment"
	case "statefulsets", "statefulset":
		return "StatefulSet"
	case "daemonsets", "daemonset":
		return "DaemonSet"
	case "nodes", "node":
		return "Node"
	case "namespaces", "namespace":
		return "Namespace"
	default:
		res := strings.TrimSuffix(resource, "s")
		if res == "" {
			return strings.Title(resource)
		}
		return strings.Title(res)
	}
}
