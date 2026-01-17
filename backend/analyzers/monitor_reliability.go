package analyzers

import (
	"cloud-sentinel-k8s/pkg/models"
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// SingleReplicaAnalyzer detects deployments/statefulsets with only 1 replica
type SingleReplicaAnalyzer struct{}

func (s *SingleReplicaAnalyzer) Name() string { return "SingleReplica" }

func (s *SingleReplicaAnalyzer) Analyze(obj *unstructured.Unstructured, client dynamic.Interface, clusterID string) []models.Anomaly {
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

// ProbeAnalyzer detects missing liveness and readiness probes
type ProbeAnalyzer struct{}

func (p *ProbeAnalyzer) Name() string { return "MissingProbes" }

func (p *ProbeAnalyzer) Analyze(obj *unstructured.Unstructured, client dynamic.Interface, clusterID string) []models.Anomaly {
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

// MissingPDBAnalyzer detects Deployments/StatefulSets without a PodDisruptionBudget
type MissingPDBAnalyzer struct{}

func (m *MissingPDBAnalyzer) Name() string { return "MissingPDB" }

func (m *MissingPDBAnalyzer) Analyze(obj *unstructured.Unstructured, client dynamic.Interface, clusterID string) []models.Anomaly {
	kind := obj.GetKind()
	if kind != "Deployment" && kind != "StatefulSet" {
		return nil
	}

	namespace := obj.GetNamespace()
	matchLabels, found, err := unstructured.NestedMap(obj.Object, "spec", "selector", "matchLabels")
	if err != nil || !found || len(matchLabels) == 0 {
		return nil
	}

	// Dynamic client lets us list PDBs
	// GVR for PDB is policy/v1/poddisruptionbudgets
	gvr := schema.GroupVersionResource{Group: "policy", Version: "v1", Resource: "poddisruptionbudgets"}

	// We need context to call List. Since Analyzer interface doesn't pass context,
	// we will use context.Background() here. Ideally, context should be passed down.
	// For this task, we will verify if `client` calls work without explicit context in signature or Use TODO.
	// client.Resource(gvr).Namespace(namespace).List(...) requires context.
	// Since we are refactoring, we should arguably pass context in AnalyzeResource, but
	// to minimize changes we can use context.TODO() as this is a background-like analysis triggered by user request.

	pdbs, err := client.Resource(gvr).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		// If we fail to list (e.g. RBAC), we silent fail analysis rather than error out the request
		return nil
	}

	hasPDB := false
	for _, pdb := range pdbs.Items {
		// Check if PDB selector matches Workload selector
		// Simplified check: if PDB selector is a subset of Workload labels?
		// Actually PDB selects Pods. Workload creates Pods with labels.
		// So we check if PDB selector matches the Workload's Pod Template Labels.

		pdbSelector, found, _ := unstructured.NestedMap(pdb.Object, "spec", "selector", "matchLabels")
		if !found {
			continue
		}

		// Check if matchLabels are identical or subset
		// For simplicity, we check if all labels in PDB selector exist in Workload selector with same value.
		match := true
		for k, v := range pdbSelector {
			if val, ok := matchLabels[k]; !ok || val != v {
				match = false
				break
			}
		}

		if match {
			hasPDB = true
			break
		}
	}

	if !hasPDB {
		return []models.Anomaly{
			NewAnomaly(
				m.Name(),
				models.SeverityWarning,
				"Missing PodDisruptionBudget",
				fmt.Sprintf("This %s is not covered by any PodDisruptionBudget (PDB).", kind),
				"Create a PDB to ensure service availability during voluntary disruptions (e.g. node drains, upgrades).",
			),
		}
	}

	return nil
}

func init() {
	GlobalAnalyzers = append(GlobalAnalyzers, &SingleReplicaAnalyzer{})
	GlobalAnalyzers = append(GlobalAnalyzers, &ProbeAnalyzer{})
	GlobalAnalyzers = append(GlobalAnalyzers, &MissingPDBAnalyzer{})
}
