package models

import (
	"time"
)

// K8sClusterContextMapping maps kubeconfig context names to user-friendly display names
type K8sClusterContextMapping struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;uniqueIndex:idx_user_context" json:"user_id"`
	ContextName string    `gorm:"not null;uniqueIndex:idx_user_context" json:"context_name"` // Original name from kubeconfig
	DisplayName string    `gorm:"not null" json:"display_name"`                              // Custom user-friendly name
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
