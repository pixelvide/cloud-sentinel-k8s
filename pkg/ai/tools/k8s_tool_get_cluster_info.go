package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	corev1 "k8s.io/api/core/v1"
)

type GetClusterInfoTool struct{}

func (t *GetClusterInfoTool) Name() string { return "get_cluster_info" }

func (t *GetClusterInfoTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "get_cluster_info",
			Description: "Get general information about the Kubernetes cluster, including server version and capacity (nodes, CPU, memory).",
			Parameters:  json.RawMessage(`{"type": "object", "properties": {}}`),
		},
	}
}

func (t *GetClusterInfoTool) Execute(ctx context.Context, args string) (string, error) {
	cs, err := GetClientSet(ctx)
	if err != nil {
		return "", err
	}

	// 1. Get Version
	version, err := cs.K8sClient.ClientSet.Discovery().ServerVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get cluster version: %w", err)
	}

	// 2. Get Nodes for detailed info
	var nodeList corev1.NodeList
	if err := cs.K8sClient.List(ctx, &nodeList); err != nil {
		return "", fmt.Errorf("failed to list nodes: %w", err)
	}

	var totalCPU int64
	var totalMem int64
	var controlPlaneCount int
	var workerCount int
	var readyCount int
	var notReadyCount int
	var platform string

	for _, node := range nodeList.Items {
		totalCPU += node.Status.Capacity.Cpu().MilliValue()
		totalMem += node.Status.Capacity.Memory().Value()

		// Check if control plane node (either label indicates control plane)
		_, isControlPlane := node.Labels["node-role.kubernetes.io/control-plane"]
		_, isMaster := node.Labels["node-role.kubernetes.io/master"]
		if isControlPlane || isMaster {
			controlPlaneCount++
		} else {
			workerCount++
		}

		// Check node readiness
		isReady := false
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
				isReady = true
				break
			}
		}
		if isReady {
			readyCount++
		} else {
			notReadyCount++
		}

		// Capture platform info from first node
		if platform == "" {
			platform = fmt.Sprintf("%s/%s", node.Status.NodeInfo.OperatingSystem, node.Status.NodeInfo.Architecture)
		}
	}

	// 3. Get Namespace count
	var nsList corev1.NamespaceList
	namespaceCount := 0
	if err := cs.K8sClient.List(ctx, &nsList); err == nil {
		namespaceCount = len(nsList.Items)
	}

	// 4. Get Pod summary
	var podList corev1.PodList
	podStats := make(map[string]int)
	totalPods := 0
	if err := cs.K8sClient.List(ctx, &podList); err == nil {
		totalPods = len(podList.Items)
		for _, pod := range podList.Items {
			podStats[string(pod.Status.Phase)]++
		}
	}

	// Build the info string
	var sb strings.Builder
	sb.WriteString("Cluster Information:\n")
	sb.WriteString(fmt.Sprintf("- Kubernetes Version: %s\n", version.GitVersion))
	sb.WriteString(fmt.Sprintf("- Platform: %s\n", platform))
	sb.WriteString("\nNode Summary:\n")
	sb.WriteString(fmt.Sprintf("- Total Nodes: %d\n", len(nodeList.Items)))
	sb.WriteString(fmt.Sprintf("- Control Plane Nodes: %d\n", controlPlaneCount))
	sb.WriteString(fmt.Sprintf("- Worker Nodes: %d\n", workerCount))
	sb.WriteString(fmt.Sprintf("- Ready Nodes: %d\n", readyCount))
	sb.WriteString(fmt.Sprintf("- Not Ready Nodes: %d\n", notReadyCount))
	sb.WriteString("\nCapacity:\n")
	sb.WriteString(fmt.Sprintf("- Total CPU: %dm\n", totalCPU))
	sb.WriteString(fmt.Sprintf("- Total Memory: %d MiB\n", totalMem/(1024*1024)))
	sb.WriteString(fmt.Sprintf("\nNamespaces: %d\n", namespaceCount))
	sb.WriteString(fmt.Sprintf("\nPod Summary (Total: %d):\n", totalPods))
	for phase, count := range podStats {
		sb.WriteString(fmt.Sprintf("- %s: %d\n", phase, count))
	}

	return sb.String(), nil
}
