package api

import (
	"encoding/json"
	"log"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
)

// RecordAuditLog records an action in the audit log table, auto-detecting actor
func RecordAuditLog(c *gin.Context, action string, payload interface{}) {
	RecordAuditLogForUser(c, action, "", payload)
}

// RecordAuditLogForUser records an action with a specific actor email
func RecordAuditLogForUser(c *gin.Context, action string, actor string, payload interface{}) {
	if actor == "" {
		actor = "system"
		user, exists := c.Get("user")
		if exists {
			if u, ok := user.(*models.User); ok {
				actor = u.Email
			}
		}
	}

	payloadStr := ""
	if payload != nil {
		if b, err := json.Marshal(payload); err == nil {
			payloadStr = string(b)
		}
	}

	auditLog := models.AuditLog{
		AppName:   "cloud-sentinel-k8s",
		Action:    action,
		Actor:     actor,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Payload:   payloadStr,
	}

	if err := models.DB.Create(&auditLog).Error; err != nil {
		log.Printf("CRITICAL: Failed to record audit log: %v", err)
	}
}
