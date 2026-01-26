package tools

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
	corev1 "k8s.io/api/core/v1"
)

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
		return "", fmt.Errorf("failed to get logs: %w", err)
	}

	return string(logs), nil
}
