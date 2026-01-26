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

// WorkloadSummaryList represents a list of WorkloadSummaries
type WorkloadSummaryList struct {
	Items []WorkloadSummary `json:"items"`
}

// ClusterSecuritySummary represents aggregated security data
type ClusterSecuritySummary struct {
	TotalVulnerabilities   VulnerabilitySummary `json:"totalVulnerabilities"`
	TotalConfigAuditIssues CheckSummary         `json:"totalConfigAuditIssues"`
	TotalExposedSecrets    CheckSummary         `json:"totalExposedSecrets"`
	VulnerableImages       int                  `json:"vulnerableImages"`
	ScannedImages          int                  `json:"scannedImages"`
	TopVulnerableWorkloads []WorkloadSummary    `json:"topVulnerableWorkloads"`
	TopMisconfigured       []WorkloadSummary    `json:"topMisconfigured"`
}

// CheckSummary is a generic summary for pass/fail check reports (ConfigAudit, RBAC, Secrets)
type CheckSummary struct {
	CriticalCount int `json:"criticalCount"`
	HighCount     int `json:"highCount"`
	MediumCount   int `json:"mediumCount"`
	LowCount      int `json:"lowCount"`
}

// =====================
// ConfigAuditReport
// =====================

// ConfigAuditCheck represents a single configuration check result
type ConfigAuditCheck struct {
	ID          string   `json:"checkID"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"` // CRITICAL, HIGH, MEDIUM, LOW
	Category    string   `json:"category"`
	Success     bool     `json:"success"`
	Messages    []string `json:"messages"`
}

// ConfigAuditReportData contains the audit findings
type ConfigAuditReportData struct {
	Scanner Scanner            `json:"scanner"`
	Summary CheckSummary       `json:"summary"`
	Checks  []ConfigAuditCheck `json:"checks"`
}

// ConfigAuditReport represents the Trivy Operator ConfigAuditReport CRD
type ConfigAuditReport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Report            ConfigAuditReportData `json:"report"`
}

// ConfigAuditReportList is a list of ConfigAuditReports
type ConfigAuditReportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConfigAuditReport `json:"items"`
}

// =====================
// ExposedSecretReport
// =====================

// ExposedSecret represents a single exposed secret finding
type ExposedSecret struct {
	Target   string `json:"target"`
	RuleID   string `json:"ruleID"`
	Title    string `json:"title"`
	Category string `json:"category"`
	Severity string `json:"severity"`
	Match    string `json:"match"`
}

// ExposedSecretReportData contains the secret findings
type ExposedSecretReportData struct {
	Scanner  Scanner         `json:"scanner"`
	Artifact Artifact        `json:"artifact"`
	Summary  CheckSummary    `json:"summary"`
	Secrets  []ExposedSecret `json:"secrets"`
}

// ExposedSecretReport represents the Trivy Operator ExposedSecretReport CRD
type ExposedSecretReport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Report            ExposedSecretReportData `json:"report"`
}

// ExposedSecretReportList is a list of ExposedSecretReports
type ExposedSecretReportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExposedSecretReport `json:"items"`
}

// =====================
// RbacAssessmentReport
// =====================

// RbacAssessmentCheck represents a single RBAC check result
type RbacAssessmentCheck struct {
	ID          string   `json:"checkID"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	Category    string   `json:"category"`
	Success     bool     `json:"success"`
	Messages    []string `json:"messages"`
}

// RbacAssessmentReportData contains the RBAC assessment findings
type RbacAssessmentReportData struct {
	Scanner Scanner               `json:"scanner"`
	Summary CheckSummary          `json:"summary"`
	Checks  []RbacAssessmentCheck `json:"checks"`
}

// RbacAssessmentReport represents the Trivy Operator RbacAssessmentReport CRD
type RbacAssessmentReport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Report            RbacAssessmentReportData `json:"report"`
}

// RbacAssessmentReportList is a list of RbacAssessmentReports
type RbacAssessmentReportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RbacAssessmentReport `json:"items"`
}

// =====================
// ClusterComplianceReport
// =====================

// ComplianceControl represents a single compliance control check result
type ComplianceControl struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Severity    string `json:"severity"` // CRITICAL, HIGH, MEDIUM, LOW
	Status      string `json:"status"`   // PASS, FAIL, WARN
	TotalFail   int    `json:"totalFail,omitempty"`
}

// ComplianceSummary represents the summary of a compliance report
type ComplianceSummary struct {
	FailCount int `json:"failCount"`
	PassCount int `json:"passCount"`
}

// ClusterComplianceReportSpec defines the compliance spec (benchmark info)
type ClusterComplianceReportSpec struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Cron        string `json:"cron,omitempty"`
}

// ClusterComplianceReportStatus contains the compliance results
type ClusterComplianceReportStatus struct {
	Summary       ComplianceSummary `json:"summary"`
	SummaryReport SummaryReport     `json:"summaryReport,omitempty"`
	DetailReport  DetailReport      `json:"detailReport,omitempty"`
	UpdatedAt     metav1.Time       `json:"updateTimestamp,omitempty"`
}

// SummaryReport contains control summaries
type SummaryReport struct {
	ControlCheck []ControlCheckSummary `json:"controlCheck,omitempty"`
}

// ControlCheckSummary is a summary for each control
type ControlCheckSummary struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Severity  string `json:"severity"`
	TotalFail int    `json:"totalFail"`
}

// DetailReport contains detailed control results
type DetailReport struct {
	// Add more fields if needed
}

// ClusterComplianceReport represents the Trivy Operator ClusterComplianceReport CRD
type ClusterComplianceReport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ClusterComplianceReportSpec   `json:"spec"`
	Status            ClusterComplianceReportStatus `json:"status"`
}

// ClusterComplianceReportList is a list of ClusterComplianceReports
type ClusterComplianceReportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterComplianceReport `json:"items"`
}
