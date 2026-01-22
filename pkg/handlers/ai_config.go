package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/model"
	"gorm.io/gorm"
)

func getUser(c *gin.Context) *model.User {
	u, exists := c.Get("user")
	if !exists {
		return nil
	}
	user, ok := u.(model.User)
	if !ok {
		return nil
	}
	return &user
}

// --- Config ---

func GetAIConfig(c *gin.Context) {
	user := getUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var settings model.AISettings
	if err := model.DB.First(&settings, "user_id = ?", user.ID).Error; err != nil {
		// Return empty default config
		c.JSON(http.StatusOK, gin.H{
			"provider": "openai",
			"model":    "gpt-3.5-turbo",
		})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func UpdateAIConfig(c *gin.Context) {
	user := getUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input model.AISettings
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.UserID = user.ID

	if err := model.DB.Save(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save settings"})
		return
	}

	c.JSON(http.StatusOK, input)
}

// --- Sessions ---

func ListSessions(c *gin.Context) {
	user := getUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var sessions []model.ChatSession
	if err := model.DB.Where("user_id = ?", user.ID).Order("updated_at desc").Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list sessions"})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

func GetSession(c *gin.Context) {
	user := getUser(c)
	id := c.Param("id")

	var session model.ChatSession
	// Preload messages order by created_at
	if err := model.DB.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at asc")
	}).Where("id = ? AND user_id = ?", id, user.ID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

func DeleteSession(c *gin.Context) {
	user := getUser(c)
	id := c.Param("id")

	if err := model.DB.Where("id = ? AND user_id = ?", id, user.ID).Delete(&model.ChatSession{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
