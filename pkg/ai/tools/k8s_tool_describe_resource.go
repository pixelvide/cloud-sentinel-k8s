package tools

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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

	// 1. Resolve GVK for the given kind
	gvk, err := resolveGVK(cs, params.Kind)
	if err != nil {
		return "", err
	}

	// 2. Prepare unstructured object
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)

	// 3. Get the resource from cache
	key := client.ObjectKey{
		Namespace: params.Namespace,
		Name:      params.Name,
	}
	if err := cs.K8sClient.Get(ctx, key, obj); err != nil {
		return "", fmt.Errorf("failed to get %s %s/%s: %w", params.Kind, params.Namespace, params.Name, err)
	}

	// 4. Serialize to JSON for the LLM
	bytes, err := json.MarshalIndent(obj.Object, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
