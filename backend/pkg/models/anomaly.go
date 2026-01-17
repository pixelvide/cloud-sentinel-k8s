package models

import "time"

// AnomalySeverity defines the severity level of an anomaly
type AnomalySeverity string

const (
	SeverityCritical AnomalySeverity = "Critical"
	SeverityWarning  AnomalySeverity = "Warning"
	SeverityInfo     AnomalySeverity = "Info"
)

// Anomaly represents a single finding by the analyzer
type Anomaly struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Severity    AnomalySeverity `json:"severity"`
	Message     string          `json:"message"`
	Description string          `json:"description"`
	Suggestion  string          `json:"suggestion"`
	CreatedAt   time.Time       `json:"created_at"`
}

// ResourceAnalysis contains all findings for a specific resource
type ResourceAnalysis struct {
	Anomalies []Anomaly `json:"anomalies"`
	Summary   string    `json:"summary"`
	Score     int       `json:"score"` // 0-100 health score
}
