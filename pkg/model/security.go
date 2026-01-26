package model

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VulnerabilityReport is a simplified representation of the Trivy Operator CRD.
// We use this for JSON unmarshalling from Unstructured or for API responses.
type VulnerabilityReport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Report VulnerabilityReportData `json:"report"`
}

type VulnerabilityReportData struct {
	Artifact        Artifact             `json:"artifact"`
	Scanner         Scanner              `json:"scanner"`
	Summary         VulnerabilitySummary `json:"summary"`
	Vulnerabilities []Vulnerability      `json:"vulnerabilities"`
}

type Artifact struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
}

type Scanner struct {
	Name    string `json:"name"`
	Vendor  string `json:"vendor"`
	Version string `json:"version"`
}

type VulnerabilitySummary struct {
	CriticalCount int `json:"criticalCount"`
	HighCount     int `json:"highCount"`
	MediumCount   int `json:"mediumCount"`
	LowCount      int `json:"lowCount"`
	UnknownCount  int `json:"unknownCount"`
}

type Vulnerability struct {
	VulnerabilityID  string   `json:"vulnerabilityID"`
	Resource         string   `json:"resource"`
	InstalledVersion string   `json:"installedVersion"`
	FixedVersion     string   `json:"fixedVersion"`
	Severity         string   `json:"severity"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	PrimaryLink      string   `json:"primaryLink"`
	Links            []string `json:"links"`
	Score            float64  `json:"score"`
}

type VulnerabilityReportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VulnerabilityReport `json:"items"`
}

// SecurityStatusResponse represents the status of the security scanning system
type SecurityStatusResponse struct {
	TrivyInstalled bool `json:"trivyInstalled"`
}

type WorkloadSummary struct {
	Namespace       string               `json:"namespace"`
	Kind            string               `json:"kind"`
	Name            string               `json:"name"`
	Vulnerabilities VulnerabilitySummary `json:"vulnerabilities"`
}

// ClusterSecuritySummary represents aggregated security data
type ClusterSecuritySummary struct {
	TotalVulnerabilities   VulnerabilitySummary `json:"totalVulnerabilities"`
	VulnerableImages       int                  `json:"vulnerableImages"`
	ScannedImages          int                  `json:"scannedImages"`
	TopVulnerableWorkloads []WorkloadSummary    `json:"topVulnerableWorkloads"`
}
