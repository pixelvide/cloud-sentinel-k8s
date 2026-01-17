package models

import (
	"time"
)

type KubeConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Name      string    `gorm:"not null" json:"name"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	IsDefault bool      `gorm:"default:false" json:"is_default"` // The main config
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
