package analyzers

import (
	"cloud-sentinel-k8s/pkg/models"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// IngressClassAnalyzer detects deprecated ingress class annotations
type IngressClassAnalyzer struct{}

func (i *IngressClassAnalyzer) Name() string { return "DeprecatedIngressClass" }

func (i *IngressClassAnalyzer) Analyze(obj *unstructured.Unstructured, client dynamic.Interface, clusterID string) []models.Anomaly {
	if obj.GetKind() != "Ingress" {
		return nil
	}

	annotations := obj.GetAnnotations()
	if annotations == nil {
		return nil
	}

	if _, exists := annotations["kubernetes.io/ingress.class"]; exists {
		return []models.Anomaly{
			NewAnomaly(
				i.Name(),
				models.SeverityWarning,
				"Deprecated Ingress Class Annotation",
				"This Ingress resource uses the deprecated 'kubernetes.io/ingress.class' annotation.",
				"Move the ingress class name to 'spec.ingressClassName' for better compliance with the modern Networking API.",
			),
		}
	}

	return nil
}

func init() {
	// Register analyzers here
	GlobalAnalyzers = append(GlobalAnalyzers, &IngressClassAnalyzer{})
}
