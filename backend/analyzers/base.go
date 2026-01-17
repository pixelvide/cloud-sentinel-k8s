package analyzers

import (
	"cloud-sentinel-k8s/pkg/models"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// Analyzer defines the interface for resource analysis rules
type Analyzer interface {
	Name() string
	Analyze(obj *unstructured.Unstructured, client dynamic.Interface, clusterID string) []models.Anomaly
}

// GlobalAnalyzers is the registry of all active analysis rules
var GlobalAnalyzers []Analyzer

// AnalyzeResource performs on-demand analysis of a Kubernetes resource
func AnalyzeResource(obj *unstructured.Unstructured, client dynamic.Interface, clusterID string) models.ResourceAnalysis {
	anomalies := []models.Anomaly{}

	// Run all registered analyzers
	for _, analyzer := range GlobalAnalyzers {
		results := analyzer.Analyze(obj, client, clusterID)
		if len(results) > 0 {
			anomalies = append(anomalies, results...)
		}
	}

	// Calculate base health score (simple for now)
	score := 100
	criticals := 0
	warnings := 0

	for _, a := range anomalies {
		if a.Severity == models.SeverityCritical {
			score -= 20
			criticals++
		} else if a.Severity == models.SeverityWarning {
			score -= 10
			warnings++
		}
	}

	if score < 0 {
		score = 0
	}

	summary := "No issues detected"
	if len(anomalies) > 0 {
		if criticals > 0 {
			summary = "Critical issues found"
		} else {
			summary = "Warnings detected"
		}
	}

	return models.ResourceAnalysis{
		Anomalies: anomalies,
		Summary:   summary,
		Score:     score,
	}
}

// NewAnomaly is a helper to create an anomaly with current timestamp
func NewAnomaly(atype string, severity models.AnomalySeverity, msg, desc, suggestion string) models.Anomaly {
	return models.Anomaly{
		ID:          atype + "-" + fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:        atype,
		Severity:    severity,
		Message:     msg,
		Description: desc,
		Suggestion:  suggestion,
		CreatedAt:   time.Now(),
	}
}
