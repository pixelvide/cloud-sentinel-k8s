package models

import (
	"time"
)

type UserIdentity struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	Provider   string    `gorm:"not null;index:idx_provider_id,unique" json:"provider"`    // e.g., "oidc", "gitlab"
	ProviderID string    `gorm:"not null;index:idx_provider_id,unique" json:"provider_id"` // The unique ID from the provider (e.g., OIDC 'sub')
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
