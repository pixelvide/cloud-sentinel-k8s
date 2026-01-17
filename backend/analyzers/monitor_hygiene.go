package analyzers

import (
	"cloud-sentinel-k8s/models"
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// DanglingServiceAnalyzer detects Services that do not select any Pods
type DanglingServiceAnalyzer struct{}

func (d *DanglingServiceAnalyzer) Name() string { return "DanglingService" }

func (d *DanglingServiceAnalyzer) Analyze(obj *unstructured.Unstructured, client dynamic.Interface) []models.Anomaly {
	if obj.GetKind() != "Service" {
		return nil
	}

	namespace := obj.GetNamespace()
	selector, found, err := unstructured.NestedMap(obj.Object, "spec", "selector")
	if err != nil || !found || len(selector) == 0 {
		return nil
	}

	// List Pods matching the selector
	// We need to construct a label selector string
	labelSelector := ""
	for k, v := range selector {
		if labelSelector != "" {
			labelSelector += ","
		}
		labelSelector += fmt.Sprintf("%s=%v", k, v)
	}

	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	pods, err := client.Resource(gvr).Namespace(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})

	if err != nil {
		return nil
	}

	if len(pods.Items) == 0 {
		return []models.Anomaly{
			NewAnomaly(
				d.Name(),
				models.SeverityWarning,
				"Dangling Service Detected",
				"This Service's selector matches 0 active Pods.",
				"Check if the selector labels match the labels on your target Pods/Deployments.",
			),
		}
	}

	return nil
}

// EmptyNamespaceAnalyzer detects Namespaces that have no significant resources
type EmptyNamespaceAnalyzer struct{}

func (e *EmptyNamespaceAnalyzer) Name() string { return "EmptyNamespace" }

func (e *EmptyNamespaceAnalyzer) Analyze(obj *unstructured.Unstructured, client dynamic.Interface) []models.Anomaly {
	if obj.GetKind() != "Namespace" {
		return nil
	}

	name := obj.GetName()
	// Skip system namespaces
	if name == "kube-system" || name == "kube-public" || name == "kube-node-lease" || name == "default" {
		return nil
	}

	// Check for presence of key resources: Deployments, Services, Pods, ConfigMaps, Secrets
	// If ALL are empty, flag it.
	// Note: This makes multiple API calls. It might be expensive if many namespaces are analyzed at once.
	// But analysis is on-demand per resource (when viewing Namespace details).

	ctx := context.TODO()

	// Helper to check resource count
	hasResource := func(gvr schema.GroupVersionResource) bool {
		list, err := client.Resource(gvr).Namespace(name).List(ctx, metav1.ListOptions{Limit: 1})
		if err != nil {
			return false
		}
		return len(list.Items) > 0
	}

	// Check Deployments
	if hasResource(schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}) {
		return nil
	}
	// Check StatefulSets
	if hasResource(schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}) {
		return nil
	}
	// Check DaemonSets
	if hasResource(schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}) {
		return nil
	}
	// Check Pods
	if hasResource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}) {
		return nil
	}
	// Check Services
	if hasResource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}) {
		return nil
	}
	// Check ConfigMaps
	if hasResource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}) {
		// Ignore default kube-root-ca.crt
		list, err := client.Resource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}).Namespace(name).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, cm := range list.Items {
				if cm.GetName() != "kube-root-ca.crt" {
					return nil
				}
			}
		}
	}
	// Check Secrets
	if hasResource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}) {
		// Secrets might be auto-created service account tokens.
		// If only secrets exist and likely default ones, we might still consider it empty-ish,
		// but let's be conservative and say if there are secrets (maybe user secrets), it's not empty.
		// Actually, default service account secret is no longer auto-created in newer k8s,
		// but there is "default-token-xxx".
		// Let's rely on workloads/services/pods mostly.
		return nil
	}

	return []models.Anomaly{
		NewAnomaly(
			e.Name(),
			models.SeverityWarning,
			"Empty Namespace Detected",
			fmt.Sprintf("Namespace '%s' appears to have no active workloads or services.", name),
			"Delete this namespace if it is no longer in use to reduce cluster clutter.",
		),
	}
}

func init() {
	GlobalAnalyzers = append(GlobalAnalyzers, &DanglingServiceAnalyzer{})
	GlobalAnalyzers = append(GlobalAnalyzers, &EmptyNamespaceAnalyzer{})
}
