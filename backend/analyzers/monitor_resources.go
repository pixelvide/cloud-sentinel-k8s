package analyzers

import (
	"cloud-sentinel-k8s/models"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// ResourceLimitsAnalyzer detects missing resource limits/requests in workloads and pods
type ResourceLimitsAnalyzer struct{}

func (r *ResourceLimitsAnalyzer) Name() string { return "ResourceLimits" }

func (r *ResourceLimitsAnalyzer) Analyze(obj *unstructured.Unstructured, client dynamic.Interface) []models.Anomaly {
	kind := obj.GetKind()
	supportedKinds := map[string]bool{
		"Deployment":  true,
		"StatefulSet": true,
		"DaemonSet":   true,
		"Pod":         true,
	}

	if !supportedKinds[kind] {
		return nil
	}

	// Helper to extract containers from different resource types
	var containers []interface{}
	var found bool
	var err error

	if kind == "Pod" {
		containers, found, err = unstructured.NestedSlice(obj.Object, "spec", "containers")
	} else {
		// Workloads (Deployment, StatefulSet, DaemonSet) store containers in spec.template.spec.containers
		containers, found, err = unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "containers")
	}

	if err != nil || !found {
		return nil
	}

	var anomalies []models.Anomaly

	for _, c := range containers {
		container, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		name, _, _ := unstructured.NestedString(container, "name")

		// Check Limits
		limits, foundLimits, _ := unstructured.NestedMap(container, "resources", "limits")
		if !foundLimits || limits == nil {
			anomalies = append(anomalies, NewAnomaly(
				r.Name(),
				models.SeverityWarning,
				"Missing Resource Limits",
				fmt.Sprintf("Container '%s' has no resource limits defined.", name),
				"Set CPU and Memory limits to prevent resource contention and ensure stable scheduling.",
			))
		} else {
			// Check specific limits if "resources.limits" exists but might be partial
			if _, hasCPU := limits["cpu"]; !hasCPU {
				anomalies = append(anomalies, NewAnomaly(
					r.Name(),
					models.SeverityWarning,
					"Missing CPU Limit",
					fmt.Sprintf("Container '%s' has no CPU limit defined.", name),
					"Set a CPU limit to prevent this container from starving others.",
				))
			}
			if _, hasMem := limits["memory"]; !hasMem {
				anomalies = append(anomalies, NewAnomaly(
					r.Name(),
					models.SeverityWarning,
					"Missing Memory Limit",
					fmt.Sprintf("Container '%s' has no Memory limit defined.", name),
					"Set a Memory limit to prevent OutOfMemory (OOM) kills affecting the node.",
				))
			}
		}
	}

	return anomalies
}

func init() {
	GlobalAnalyzers = append(GlobalAnalyzers, &ResourceLimitsAnalyzer{})
}
