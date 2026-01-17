package models

import (
	"time"
)

type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	AppName   string    `gorm:"index" json:"app_name"` // e.g., "cloud-sentinel"
	Action    string    `gorm:"index" json:"action"`   // e.g., "USER_LOGIN", "UPDATE_KUBE_CONFIG"
	Actor     string    `gorm:"index" json:"actor"`    // User email or ID
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Payload   string    `gorm:"type:text" json:"payload"` // JSON encoded payload
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
