package tools

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

// --- Scale Deployment Tool ---

type ScaleDeploymentTool struct{}

func (t *ScaleDeploymentTool) Name() string { return "scale_deployment" }

func (t *ScaleDeploymentTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "scale_deployment",
			Description: "Scale a deployment to a specified number of replicas. Requires confirmation.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"namespace": {
						"type": "string",
						"description": "The namespace of the deployment."
					},
					"name": {
						"type": "string",
						"description": "The name of the deployment."
					},
					"replicas": {
						"type": "integer",
						"description": "The target number of replicas."
					},
					"confirm": {
						"type": "boolean",
						"description": "Set to true to actually execute the scaling. Defaults to false (dry-run)."
					}
				},
				"required": ["namespace", "name", "replicas"]
			}`),
		},
	}
}

func (t *ScaleDeploymentTool) Execute(ctx context.Context, args string) (string, error) {
	var params struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
		Replicas  int32  `json:"replicas"`
		Confirm   bool   `json:"confirm"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}
	if params.Replicas < 0 {
		return "", fmt.Errorf("replicas cannot be negative")
	}

	cs, err := GetClientSet(ctx)
	if err != nil {
		return "", err
	}

	deployClient := cs.K8sClient.ClientSet.AppsV1().Deployments(params.Namespace)

	// Always get first to check existence
	deploy, err := deployClient.Get(ctx, params.Name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get deployment: %w", err)
	}

	currentReplicas := int32(0)
	if deploy.Spec.Replicas != nil {
		currentReplicas = *deploy.Spec.Replicas
	}

	if !params.Confirm {
		return fmt.Sprintf("Dry run: Deployment '%s/%s' currently has %d replicas. Would scale to %d. To execute, call this tool again with 'confirm' set to true.",
			params.Namespace, params.Name, currentReplicas, params.Replicas), nil
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		d, err := deployClient.Get(ctx, params.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		d.Spec.Replicas = &params.Replicas
		_, err = deployClient.Update(ctx, d, metav1.UpdateOptions{})
		return err
	})

	if retryErr != nil {
		return "", fmt.Errorf("failed to update deployment: %w", retryErr)
	}

	return fmt.Sprintf("Successfully scaled deployment '%s/%s' from %d to %d replicas.",
		params.Namespace, params.Name, currentReplicas, params.Replicas), nil
}
