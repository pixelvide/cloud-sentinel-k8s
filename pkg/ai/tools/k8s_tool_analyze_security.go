package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// --- Analyze Security Tool ---

type AnalyzeSecurityTool struct{}

func (t *AnalyzeSecurityTool) Name() string { return "analyze_security" }

func (t *AnalyzeSecurityTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "analyze_security",
			Description: "Analyze the security context of a specific resource (Pod or Deployment) and report potential issues. If you don't know the specific pod or namespace, find them first using 'list_pods' or 'list_resources'.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"namespace": {
						"type": "string",
						"description": "The namespace of the resource."
					},
					"kind": {
						"type": "string",
						"enum": ["Pod", "Deployment"],
						"description": "The kind of resource."
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

func (t *AnalyzeSecurityTool) Execute(ctx context.Context, args string) (string, error) {
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

	var findings []string
	var podSpec *corev1.PodSpec

	switch params.Kind {
	case "Pod":
		pod, err := cs.K8sClient.ClientSet.CoreV1().Pods(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		podSpec = &pod.Spec
	case "Deployment":
		deploy, err := cs.K8sClient.ClientSet.AppsV1().Deployments(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		podSpec = &deploy.Spec.Template.Spec
	default:
		return "", fmt.Errorf("unsupported kind: %s", params.Kind)
	}

	// Simple analysis logic
	for _, container := range podSpec.Containers {
		if container.SecurityContext != nil {
			if container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
				findings = append(findings, fmt.Sprintf("Container '%s' is running as Privileged.", container.Name))
			}
			if container.SecurityContext.RunAsNonRoot != nil && !*container.SecurityContext.RunAsNonRoot {
				findings = append(findings, fmt.Sprintf("Container '%s' allows running as root.", container.Name))
			}
		} else {
			findings = append(findings, fmt.Sprintf("Container '%s' has no SecurityContext defined.", container.Name))
		}
	}

	if podSpec.HostNetwork {
		findings = append(findings, "Pod is using HostNetwork.")
	}
	if podSpec.HostPID {
		findings = append(findings, "Pod is using HostPID.")
	}

	if len(findings) == 0 {
		return "No obvious security issues found in basic scan.", nil
	}
	return "Security Analysis Findings:\n- " + strings.Join(findings, "\n- "), nil
}
