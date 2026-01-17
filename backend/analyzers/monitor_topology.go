package analyzers

import (
	"cloud-sentinel-k8s/models"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// TopologySpreadAnalyzer detects missing topology spread constraints in workloads
type TopologySpreadAnalyzer struct{}

func (t *TopologySpreadAnalyzer) Name() string { return "TopologySpreadConstraints" }

func (t *TopologySpreadAnalyzer) Analyze(obj *unstructured.Unstructured, client dynamic.Interface) []models.Anomaly {
	kind := obj.GetKind()
	if kind != "Deployment" && kind != "StatefulSet" {
		return nil
	}

	// Check spec.template.spec.topologySpreadConstraints
	constraints, found, err := unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "topologySpreadConstraints")
	if err != nil {
		return nil
	}

	if !found || len(constraints) == 0 {
		return []models.Anomaly{
			NewAnomaly(
				t.Name(),
				models.SeverityWarning,
				"Missing Topology Spread Constraints",
				fmt.Sprintf("This %s does not specify any topology spread constraints.", kind),
				"Define 'topologySpreadConstraints' in spec.template.spec to ensure high availability across zones or nodes.",
			),
		}
	}

	return nil
}

func init() {
	// Register analyzers here
	GlobalAnalyzers = append(GlobalAnalyzers, &TopologySpreadAnalyzer{})
	GlobalAnalyzers = append(GlobalAnalyzers, &AffinityAnalyzer{})
}

// AffinityAnalyzer detects conflicting affinity and anti-affinity rules
type AffinityAnalyzer struct{}

func (a *AffinityAnalyzer) Name() string { return "ConflictingAffinity" }

func (a *AffinityAnalyzer) Analyze(obj *unstructured.Unstructured, client dynamic.Interface) []models.Anomaly {
	kind := obj.GetKind()
	supportedKinds := map[string]bool{
		"Deployment":  true,
		"StatefulSet": true,
		"DaemonSet":   true,
		"Job":         true,
		"Pod":         true, // Added Pod support explicitely as user provided a Pod YAML
	}

	if !supportedKinds[kind] {
		return nil
	}

	// Extract affinity
	var affinity map[string]interface{}
	var found bool
	var err error

	if kind == "Pod" {
		affinity, found, err = unstructured.NestedMap(obj.Object, "spec", "affinity")
	} else {
		affinity, found, err = unstructured.NestedMap(obj.Object, "spec", "template", "spec", "affinity")
	}

	if err != nil || !found {
		return nil
	}

	// Helper to extract terms from both required and preferred
	getTerms := func(affinityType string) []map[string]interface{} {
		var terms []map[string]interface{}

		// 1. Required (Hard)
		required, _, _ := unstructured.NestedSlice(affinity, affinityType, "requiredDuringSchedulingIgnoredDuringExecution")
		for _, t := range required {
			if term, ok := t.(map[string]interface{}); ok {
				terms = append(terms, term)
			}
		}

		// 2. Preferred (Soft)
		preferred, _, _ := unstructured.NestedSlice(affinity, affinityType, "preferredDuringSchedulingIgnoredDuringExecution")
		for _, p := range preferred {
			pref, ok := p.(map[string]interface{})
			if !ok {
				continue
			}
			// Preferred has structure: { weight: ..., podAffinityTerm: { ... } }
			if term, found, _ := unstructured.NestedMap(pref, "podAffinityTerm"); found {
				terms = append(terms, term)
			}
		}
		return terms
	}

	affinityTerms := getTerms("podAffinity")
	antiAffinityTerms := getTerms("podAntiAffinity")

	if len(affinityTerms) == 0 || len(antiAffinityTerms) == 0 {
		return nil
	}

	var anomalies []models.Anomaly

	// Naive comparison: if topologyKey and labelSelector match exactly
	for _, affTerm := range affinityTerms {
		affTopology, _, _ := unstructured.NestedString(affTerm, "topologyKey")
		affSelector, _, _ := unstructured.NestedMap(affTerm, "labelSelector")

		for _, antiTerm := range antiAffinityTerms {
			antiTopology, _, _ := unstructured.NestedString(antiTerm, "topologyKey")
			antiSelector, _, _ := unstructured.NestedMap(antiTerm, "labelSelector")

			if affTopology == antiTopology && fmt.Sprint(affSelector) == fmt.Sprint(antiSelector) {
				anomalies = append(anomalies, NewAnomaly(
					a.Name(),
					models.SeverityWarning,
					"Conflicting Affinity Rules",
					fmt.Sprintf("Pod Affinity and Anti-Affinity rules contradict each other on topology key '%s'.", affTopology),
					"Remove or adjust one of the rules. Requiring a pod to be ON a node (Affinity) and NOT ON that node (Anti-Affinity) with the same criteria is self-canceling.",
				))
			}
		}
	}

	return anomalies
}
