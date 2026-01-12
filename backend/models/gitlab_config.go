package models

import (
	"time"
)

type GitlabConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;uniqueIndex:idx_user_host" json:"user_id"`
	Host        string    `gorm:"not null;uniqueIndex:idx_user_host" json:"gitlab_host"`
	IsHTTPS     bool      `gorm:"default:true" json:"is_https"`
	Token       string    `gorm:"not null" json:"token"` // TODO: Encrypt this field in production
	IsValidated bool      `gorm:"default:false" json:"is_validated"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
