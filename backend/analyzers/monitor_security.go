package analyzers

import (
	"cloud-sentinel-k8s/models"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ImmutableTagAnalyzer detects usage of 'latest' tag or missing tags
type ImmutableTagAnalyzer struct{}

func (i *ImmutableTagAnalyzer) Name() string { return "ImmutableTags" }

func (i *ImmutableTagAnalyzer) Analyze(obj *unstructured.Unstructured) []models.Anomaly {
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
	var initContainers []interface{}
	var foundC, foundIC bool
	var errC, errIC error

	if kind == "Pod" {
		containers, foundC, errC = unstructured.NestedSlice(obj.Object, "spec", "containers")
		initContainers, foundIC, errIC = unstructured.NestedSlice(obj.Object, "spec", "initContainers")
	} else {
		// Workloads (Deployment, StatefulSet, DaemonSet) store containers in spec.template.spec.containers
		containers, foundC, errC = unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "containers")
		initContainers, foundIC, errIC = unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "initContainers")
	}

	if (errC != nil || !foundC) && (errIC != nil || !foundIC) {
		return nil
	}

	var anomalies []models.Anomaly
	allContainers := append(containers, initContainers...)

	for _, c := range allContainers {
		container, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		name, _, _ := unstructured.NestedString(container, "name")
		image, found, _ := unstructured.NestedString(container, "image")

		if found {
			// Check if tag is 'latest' or missing (which implies latest)
			// Examples: "nginx", "nginx:latest", "my-registry.io/img:latest"
			isLatest := false
			if strings.HasSuffix(image, ":latest") {
				isLatest = true
			} else if !strings.Contains(image, ":") {
				// No tag specified usually means latest, unless it has a digest (sha256:...)
				// "ubuntu@sha256:..." is valid and immutable.
				if !strings.Contains(image, "@") {
					isLatest = true
				}
			}

			if isLatest {
				anomalies = append(anomalies, NewAnomaly(
					i.Name(),
					models.SeverityWarning,
					"Mutable Image Tag Detected",
					fmt.Sprintf("Container '%s' is using a mutable image tag: '%s'.", name, image),
					"Use a specific version tag (e.g., :v1.0.0) or digest (@sha256:...) to ensure immutability and reproducible deployments.",
				))
			}
		}
	}

	return anomalies
}

// PrivilegedContainerAnalyzer detects containers running with privileged: true
type PrivilegedContainerAnalyzer struct{}

func (p *PrivilegedContainerAnalyzer) Name() string { return "PrivilegedContainer" }

func (p *PrivilegedContainerAnalyzer) Analyze(obj *unstructured.Unstructured) []models.Anomaly {
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
	var initContainers []interface{}
	var foundC, foundIC bool
	var errC, errIC error

	if kind == "Pod" {
		containers, foundC, errC = unstructured.NestedSlice(obj.Object, "spec", "containers")
		initContainers, foundIC, errIC = unstructured.NestedSlice(obj.Object, "spec", "initContainers")
	} else {
		// Workloads (Deployment, StatefulSet, DaemonSet) store containers in spec.template.spec.containers
		containers, foundC, errC = unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "containers")
		initContainers, foundIC, errIC = unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "initContainers")
	}

	if (errC != nil || !foundC) && (errIC != nil || !foundIC) {
		return nil
	}

	var anomalies []models.Anomaly
	allContainers := append(containers, initContainers...)

	for _, c := range allContainers {
		container, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		name, _, _ := unstructured.NestedString(container, "name")
		privileged, found, _ := unstructured.NestedBool(container, "securityContext", "privileged")

		if found && privileged {
			anomalies = append(anomalies, NewAnomaly(
				p.Name(),
				models.SeverityWarning,
				"Privileged Container Detected",
				fmt.Sprintf("Container '%s' is running in privileged mode.", name),
				"Avoid running containers as privileged unless absolutely necessary. Grant specific capabilities instead.",
			))
		}
	}

	return anomalies
}

func init() {
	GlobalAnalyzers = append(GlobalAnalyzers, &ImmutableTagAnalyzer{})
	GlobalAnalyzers = append(GlobalAnalyzers, &PrivilegedContainerAnalyzer{})
}
