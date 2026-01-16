package analyzers

import (
	"cloud-sentinel-k8s/models"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// SingleReplicaAnalyzer detects deployments/statefulsets with only 1 replica
type SingleReplicaAnalyzer struct{}

func (s *SingleReplicaAnalyzer) Name() string { return "SingleReplica" }

func (s *SingleReplicaAnalyzer) Analyze(obj *unstructured.Unstructured) []models.Anomaly {
	kind := obj.GetKind()
	if kind != "Deployment" && kind != "StatefulSet" {
		return nil
	}

	// Helpers to get replicas
	// replicas is a pointer in Deployment Spec, thus can be nil (defaults to 1) or 0 or >1.
	replicas, found, err := unstructured.NestedInt64(obj.Object, "spec", "replicas")

	// If not found or err, it usually means default of 1 for these resources.
	// But strictly speaking, if it's missing, it is 1.
	currentReplicas := int64(1)
	if err == nil && found {
		currentReplicas = replicas
	}

	if currentReplicas == 1 {
		return []models.Anomaly{
			NewAnomaly(
				s.Name(),
				models.SeverityWarning,
				"Single Replica Detected",
				fmt.Sprintf("This %s has only 1 replica running.", kind),
				"Increase replica count to at least 2 to ensure high availability and zero-downtime rolling updates.",
			),
		}
	}

	return nil
}

func init() {
	GlobalAnalyzers = append(GlobalAnalyzers, &SingleReplicaAnalyzer{})
	GlobalAnalyzers = append(GlobalAnalyzers, &ProbeAnalyzer{})
}

// ProbeAnalyzer detects missing liveness and readiness probes
type ProbeAnalyzer struct{}

func (p *ProbeAnalyzer) Name() string { return "MissingProbes" }

func (p *ProbeAnalyzer) Analyze(obj *unstructured.Unstructured) []models.Anomaly {
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

	// Helper to extract containers
	var containers []interface{}
	var found bool
	var err error

	if kind == "Pod" {
		containers, found, err = unstructured.NestedSlice(obj.Object, "spec", "containers")
	} else {
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

		_, hasLiveness, _ := unstructured.NestedMap(container, "livenessProbe")
		_, hasReadiness, _ := unstructured.NestedMap(container, "readinessProbe")

		if !hasLiveness {
			anomalies = append(anomalies, NewAnomaly(
				p.Name(),
				models.SeverityWarning,
				"Missing Liveness Probe",
				fmt.Sprintf("Container '%s' is missing a liveness probe.", name),
				"Define a liveness probe to allow Kubernetes to restart the container if it deadlocks or crashes locally.",
			))
		}

		if !hasReadiness {
			anomalies = append(anomalies, NewAnomaly(
				p.Name(),
				models.SeverityWarning,
				"Missing Readiness Probe",
				fmt.Sprintf("Container '%s' is missing a readiness probe.", name),
				"Define a readiness probe to prevent traffic from being sent to the container before it is ready to handle requests.",
			))
		}
	}

	return anomalies
}
