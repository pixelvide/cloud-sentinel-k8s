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

func ListAIConfigs(c *gin.Context) {
	user := getUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var settings []model.AISettings
	if err := model.DB.Where("user_id = ?", user.ID).Find(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func GetAIConfig(c *gin.Context) {
	user := getUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var settings model.AISettings
	// Priority: 1. Default, 2. Active, 3. Any
	if err := model.DB.Where("user_id = ? AND is_default = ?", user.ID, true).First(&settings).Error; err != nil {
		if err := model.DB.Where("user_id = ? AND is_active = ?", user.ID, true).First(&settings).Error; err != nil {
			if err := model.DB.Where("user_id = ?", user.ID).First(&settings).Error; err != nil {
				c.JSON(http.StatusOK, gin.H{})
				return
			}
		}
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
	input.IsActive = true

	// Clear ID from input - we'll set it only if we find an existing record for this user+profile
	// This fixes the bug where the frontend was sending the old ID when switching profiles
	input.ID = 0

	// Ensure we don't violate unique constraint (user_id, profile_id)
	var existing model.AISettings
	if err := model.DB.Where("user_id = ? AND profile_id = ?", user.ID, input.ProfileID).First(&existing).Error; err == nil {
		input.ID = existing.ID
		input.CreatedAt = existing.CreatedAt
	}

	// Ensure only one is active for user
	model.DB.Model(&model.AISettings{}).Where("user_id = ?", user.ID).Update("is_active", false)

	// If IsDefault is true, clear all other defaults for this user
	if input.IsDefault {
		model.DB.Model(&model.AISettings{}).Where("user_id = ?", user.ID).Update("is_default", false)
	}

	if err := model.DB.Save(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save settings"})
		return
	}

	c.JSON(http.StatusOK, input)
}

func DeleteAIConfig(c *gin.Context) {
	user := getUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")
	if err := model.DB.Where("id = ? AND user_id = ?", id, user.ID).Delete(&model.AISettings{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func UpdateUserConfig(c *gin.Context) {
	user := getUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input struct {
		IsAIChatEnabled *bool `json:"is_ai_chat_enabled"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := model.GetUserConfig(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user config"})
		return
	}

	if input.IsAIChatEnabled != nil {
		config.IsAIChatEnabled = *input.IsAIChatEnabled
	}

	if err := model.DB.Save(config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user config"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// --- Sessions ---

func ListAIChatSessions(c *gin.Context) {
	user := getUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var sessions []model.AIChatSession
	if err := model.DB.Where("user_id = ?", user.ID).Order("updated_at desc").Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list sessions"})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

func GetAIChatSession(c *gin.Context) {
	user := getUser(c)
	id := c.Param("id")

	var session model.AIChatSession
	// Preload messages order by created_at
	if err := model.DB.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at asc")
	}).Where("id = ? AND user_id = ?", id, user.ID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

func DeleteAIChatSession(c *gin.Context) {
	user := getUser(c)
	id := c.Param("id")

	if err := model.DB.Where("id = ? AND user_id = ?", id, user.ID).Delete(&model.AIChatSession{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
func GetAvailableModels(c *gin.Context) {
	user := getUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Logic:
	// 1. Check if the user has a configured AI profile (priority: default, active, any).
	// 2. If found, return its AllowedModels, DefaultModel, and Provider.
	// 3. If not found:
	//    - Check AI_FORCE_USER_KEYS app config.
	//    - If true, return message asking user to configure profiles.
	//    - If false, use system default models list.

	var userSettings model.AISettings
	err := model.DB.Where("user_id = ? AND is_default = ?", user.ID, true).First(&userSettings).Error
	if err != nil {
		err = model.DB.Where("user_id = ? AND is_active = ?", user.ID, true).First(&userSettings).Error
		if err != nil {
			err = model.DB.Where("user_id = ?", user.ID).First(&userSettings).Error
		}
	}

	if err == nil {
		// Use user settings
		var profile model.AIProviderProfile
		if err := model.DB.Where("is_enabled = ?", true).First(&profile, userSettings.ProfileID).Error; err == nil {
			c.JSON(http.StatusOK, gin.H{
				"models":   profile.AllowedModels,
				"default":  profile.DefaultModel,
				"provider": profile.Provider,
			})
			return
		}
	}

	// Check governance
	aiForceUserKeysCfg, _ := model.GetAppConfig(model.CurrentApp.ID, model.AIForceUserKeys)
	if aiForceUserKeysCfg != nil && aiForceUserKeysCfg.Value == "true" {
		c.JSON(http.StatusOK, gin.H{
			"models":  []string{},
			"message": "Please configure your AI profile in settings to use the chat.",
		})
		return
	}

	// Use system default
	var profile model.AIProviderProfile
	if err := model.DB.Where("is_system = ? AND is_enabled = ?", true, true).First(&profile).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{
			"models":   profile.AllowedModels,
			"default":  profile.DefaultModel,
			"provider": profile.Provider,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"models":  []string{},
		"message": "AI is not configured by the administrator.",
	})
}
