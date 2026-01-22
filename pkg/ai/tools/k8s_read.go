package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pixelvide/cloud-sentinel-k8s/pkg/cluster"
	openai "github.com/sashabaranov/go-openai"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClientSetKey struct{}

func GetClientSet(ctx context.Context) (*cluster.ClientSet, error) {
	cs, ok := ctx.Value(ClientSetKey{}).(*cluster.ClientSet)
	if !ok || cs == nil {
		return nil, fmt.Errorf("kubernetes client not found in context")
	}
	return cs, nil
}

// --- List Pods Tool ---

type ListPodsTool struct{}

func (t *ListPodsTool) Name() string { return "list_pods" }

func (t *ListPodsTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "list_pods",
			Description: "List pods in a namespace, optionally filtered by status",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"namespace": {
						"type": "string",
						"description": "The namespace to list pods from. Defaults to 'default'."
					},
					"status_filter": {
						"type": "string",
						"enum": ["Running", "Pending", "Failed", "Succeeded", "Unknown"],
						"description": "Filter pods by status phase."
					}
				}
			}`),
		},
	}
}

func (t *ListPodsTool) Execute(ctx context.Context, args string) (string, error) {
	var params struct {
		Namespace    string `json:"namespace"`
		StatusFilter string `json:"status_filter"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}
	if params.Namespace == "" {
		params.Namespace = "default"
	}

	cs, err := GetClientSet(ctx)
	if err != nil {
		return "", err
	}

	pods, err := cs.K8sClient.ClientSet.CoreV1().Pods(params.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	var results []string
	for _, pod := range pods.Items {
		if params.StatusFilter != "" && string(pod.Status.Phase) != params.StatusFilter {
			continue
		}

		restarts := 0
		for _, status := range pod.Status.ContainerStatuses {
			restarts += int(status.RestartCount)
		}

		results = append(results, fmt.Sprintf("%s (Status: %s, Restarts: %d, IP: %s)",
			pod.Name, pod.Status.Phase, restarts, pod.Status.PodIP))
	}

	if len(results) == 0 {
		return "No pods found.", nil
	}

	// Limit output to prevent token overflow
	if len(results) > 50 {
		return strings.Join(results[:50], "\n") + fmt.Sprintf("\n... and %d more", len(results)-50), nil
	}

	return strings.Join(results, "\n"), nil
}

// --- Get Pod Logs Tool ---

type GetPodLogsTool struct{}

func (t *GetPodLogsTool) Name() string { return "get_pod_logs" }

func (t *GetPodLogsTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "get_pod_logs",
			Description: "Get logs from a specific pod",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"namespace": {
						"type": "string",
						"description": "The namespace of the pod."
					},
					"pod_name": {
						"type": "string",
						"description": "The name of the pod."
					},
					"container": {
						"type": "string",
						"description": "Optional container name."
					},
					"lines": {
						"type": "integer",
						"description": "Number of lines to retrieve (max 100). Defaults to 50."
					}
				},
				"required": ["namespace", "pod_name"]
			}`),
		},
	}
}

func (t *GetPodLogsTool) Execute(ctx context.Context, args string) (string, error) {
	var params struct {
		Namespace string `json:"namespace"`
		PodName   string `json:"pod_name"`
		Container string `json:"container"`
		Lines     int64  `json:"lines"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}

	if params.Lines <= 0 {
		params.Lines = 50
	}
	if params.Lines > 100 {
		params.Lines = 100
	}

	cs, err := GetClientSet(ctx)
	if err != nil {
		return "", err
	}

	opts := &corev1.PodLogOptions{
		TailLines: &params.Lines,
	}
	if params.Container != "" {
		opts.Container = params.Container
	}

	req := cs.K8sClient.ClientSet.CoreV1().Pods(params.Namespace).GetLogs(params.PodName, opts)
	logs, err := req.DoRaw(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %v", err)
	}

	return string(logs), nil
}

// --- Describe Resource Tool ---

type DescribeResourceTool struct{}

func (t *DescribeResourceTool) Name() string { return "describe_resource" }

func (t *DescribeResourceTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "describe_resource",
			Description: "Get details (JSON) of a specific resource",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"namespace": {
						"type": "string",
						"description": "The namespace of the resource."
					},
					"kind": {
						"type": "string",
						"description": "The kind of resource (Pod, Deployment, Service, etc)."
					},
					"name": {
						"type": "string",
						"description": "The name of the resource."
					}
				},
				"required": ["namespace", "kind", "name"]
			}`),
		},
	}
}

func (t *DescribeResourceTool) Execute(ctx context.Context, args string) (string, error) {
	var params struct {
		Namespace string `json:"namespace"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}

	cs, err := GetClientSet(ctx)
	if err != nil {
		return "", err
	}

	var obj interface{}
	var getErr error

	switch strings.ToLower(params.Kind) {
	case "pod":
		obj, getErr = cs.K8sClient.ClientSet.CoreV1().Pods(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
	case "deployment":
		obj, getErr = cs.K8sClient.ClientSet.AppsV1().Deployments(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
	case "service":
		obj, getErr = cs.K8sClient.ClientSet.CoreV1().Services(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
	case "node":
		obj, getErr = cs.K8sClient.ClientSet.CoreV1().Nodes().Get(ctx, params.Name, metav1.GetOptions{})
	default:
		return "", fmt.Errorf("unsupported resource kind: %s", params.Kind)
	}

	if getErr != nil {
		return "", getErr
	}

	// Serialize to JSON for the LLM
	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
