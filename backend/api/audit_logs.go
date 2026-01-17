package api

import (
	"log"
	"net/http"
	"strconv"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
)

// GetMyAuditLogs returns paginated audit logs for the current user
func GetMyAuditLogs(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	var logs []models.AuditLog
	var total int64

	query := models.DB.Model(&models.AuditLog{}).Where("actor = ?", user.Email)

	if err := query.Count(&total).Error; err != nil {
		log.Printf("Failed to count audit logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch logs count"})
		return
	}

	if err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		log.Printf("Failed to fetch audit logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":     logs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}
