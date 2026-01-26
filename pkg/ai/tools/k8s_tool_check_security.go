package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pixelvide/cloud-sentinel-k8s/pkg/model"
	openai "github.com/sashabaranov/go-openai"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pixelvide/cloud-sentinel-k8s/pkg/cluster"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// --- Check Security Tool ---

// --- Check Image Security Tool ---

type CheckImageSecurityTool struct {
	ClientSet *cluster.ClientSet
}

func (t *CheckImageSecurityTool) Name() string { return "check_image_security" }

func (t *CheckImageSecurityTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "check_image_security",
			Description: "Check for security vulnerabilities in a specific resource (Pod, Deployment, etc.) by reading Trivy reports. Make sure Trivy Operator is installed in the cluster.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"namespace": {
						"type": "string",
						"description": "The namespace of the resource."
					},
					"kind": {
						"type": "string",
						"description": "The kind of resource (e.g., Pod, Deployment, StatefulSet)."
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

func (t *CheckImageSecurityTool) Execute(ctx context.Context, args string) (string, error) {
	var params struct {
		Namespace string `json:"namespace"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}

	// Initialize ClientSet if not already set (e.g., when tool is created dynamically)
	if t.ClientSet == nil {
		cs, err := GetClientSet(ctx)
		if err != nil {
			return "", err
		}
		t.ClientSet = cs
	}

	// 1. Check if CRD exists
	var crd apiextensionsv1.CustomResourceDefinition
	if err := t.ClientSet.K8sClient.Get(ctx, client.ObjectKey{Name: "vulnerabilityreports.aquasecurity.github.io"}, &crd); err != nil {
		// If CRD is missing, return friendly message without error (as this is a valid state for the tool)
		return "Trivy Operator is not installed or VulnerabilityReport CRD is missing in this cluster. Please install trivy-operator to enable security scanning.", nil //nolint:nilerr
	}

	// 2. List Reports
	var list unstructured.UnstructuredList
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "aquasecurity.github.io",
		Version: "v1alpha1",
		Kind:    "VulnerabilityReport",
	})

	opts := []client.ListOption{
		client.InNamespace(params.Namespace),
		client.MatchingLabels{
			"trivy-operator.resource.kind": params.Kind,
			"trivy-operator.resource.name": params.Name,
		},
	}

	if err := t.ClientSet.K8sClient.List(ctx, &list, opts...); err != nil {
		return "", fmt.Errorf("failed to list vulnerability reports: %w", err)
	}

	if len(list.Items) == 0 {
		return fmt.Sprintf("No vulnerability reports found for %s/%s in namespace %s. The scan might be pending or the resource has not been scanned yet.", params.Kind, params.Name, params.Namespace), nil
	}

	// 3. Summarize
	var summaryBuilder strings.Builder
	summaryBuilder.WriteString(fmt.Sprintf("Security Report for %s/%s:\n", params.Kind, params.Name))

	for _, u := range list.Items {
		var report model.VulnerabilityReport
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
			continue
		}

		repo := report.Report.Artifact.Repository
		tag := report.Report.Artifact.Tag
		s := report.Report.Summary

		summaryBuilder.WriteString(fmt.Sprintf("- Image: %s:%s\n", repo, tag))
		summaryBuilder.WriteString(fmt.Sprintf("  Summary: Critical: %d, High: %d, Medium: %d, Low: %d\n", s.CriticalCount, s.HighCount, s.MediumCount, s.LowCount))

		if s.CriticalCount > 0 || s.HighCount > 0 {
			summaryBuilder.WriteString("  Top Vulnerabilities:\n")
			count := 0
			for _, v := range report.Report.Vulnerabilities {
				if (v.Severity == "CRITICAL" || v.Severity == "HIGH") && count < 5 {
					summaryBuilder.WriteString(fmt.Sprintf("    - [%s] %s (%s): %s\n", v.Severity, v.VulnerabilityID, v.Resource, v.Title))
					count++
				}
			}
			if count == 0 {
				summaryBuilder.WriteString("    (No Critical/High vulnerabilities listed in detail)\n")
			}
		}
		summaryBuilder.WriteString("\n")
	}

	return summaryBuilder.String(), nil
}
