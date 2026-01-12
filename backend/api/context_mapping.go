package api

import (
	"cloud-sentinel-k8s/db"
	"cloud-sentinel-k8s/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListContextMappings returns all context mappings for the current user
func ListContextMappings(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var mappings []models.K8sClusterContextMapping
	if err := db.DB.Where("user_id = ?", user.ID).Find(&mappings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch mappings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mappings": mappings})
}

// UpsertContextMapping creates or updates a context mapping
func UpsertContextMapping(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input struct {
		ContextName string `json:"context_name" binding:"required"`
		DisplayName string `json:"display_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var mapping models.K8sClusterContextMapping
	result := db.DB.Where("user_id = ? AND context_name = ?", user.ID, input.ContextName).First(&mapping)

	if result.Error != nil {
		// Create new mapping
		mapping = models.K8sClusterContextMapping{
			UserID:      user.ID,
			ContextName: input.ContextName,
			DisplayName: input.DisplayName,
		}
		if err := db.DB.Create(&mapping).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create mapping"})
			return
		}
	} else {
		// Update existing
		mapping.DisplayName = input.DisplayName
		if err := db.DB.Save(&mapping).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update mapping"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"mapping": mapping})
}

// DeleteContextMapping removes a context mapping
func DeleteContextMapping(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Param("id")
	if err := db.DB.Where("id = ? AND user_id = ?", id, user.ID).Delete(&models.K8sClusterContextMapping{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete mapping"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "mapping deleted"})
}
