package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ListPodsTool struct{}

func (t *ListPodsTool) Name() string { return "list_pods" }

func (t *ListPodsTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "list_pods",
			Description: "List pods in a namespace, optionally filtered by status or node name. Use this tool when you need to see pods for a specific node or across the cluster.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"namespace": {
						"type": "string",
						"description": "The namespace to list pods from. If empty, lists from all namespaces."
					},
					"status_filter": {
						"type": "string",
						"enum": ["Running", "Pending", "Failed", "Succeeded", "Unknown"],
						"description": "Filter pods by status phase."
					},
					"node": {
						"type": "string",
						"description": "Filter pods by node name."
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
		Node         string `json:"node"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}

	cs, err := GetClientSet(ctx)
	if err != nil {
		return "", err
	}

	// Use cached client
	listUpdates, err := buildListOptions(params.Namespace, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	if params.Node != "" {
		listUpdates = append(listUpdates, client.MatchingFields{"spec.nodeName": params.Node})
	}

	var list corev1.PodList
	if err := cs.K8sClient.List(ctx, &list, listUpdates...); err != nil {
		return "", err
	}

	var results []string
	for _, pod := range list.Items {
		if params.StatusFilter != "" && string(pod.Status.Phase) != params.StatusFilter {
			continue
		}

		restarts := 0
		for _, status := range pod.Status.ContainerStatuses {
			restarts += int(status.RestartCount)
		}

		results = append(results, fmt.Sprintf("%s/%s (Status: %s, Restarts: %d, IP: %s)",
			pod.Namespace, pod.Name, pod.Status.Phase, restarts, pod.Status.PodIP))
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
