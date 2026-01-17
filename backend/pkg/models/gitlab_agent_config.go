package models

import (
	"time"
)

type GitlabK8sAgentConfig struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"not null" json:"user_id"`
	GitlabConfigID uint      `gorm:"not null" json:"gitlab_config_id"`
	AgentID        string    `gorm:"not null" json:"agent_id"`
	AgentRepo      string    `gorm:"not null" json:"agent_repo"`
	IsConfigured   bool      `gorm:"default:false" json:"is_configured"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relationships
	GitlabConfig GitlabConfig `gorm:"foreignKey:GitlabConfigID" json:"gitlab_config"`
}
