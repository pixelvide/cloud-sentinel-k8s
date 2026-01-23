package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/model"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/rbac"
)

// --- Profiles ---

func ListAIProfiles(c *gin.Context) {
	var profiles []model.AIProviderProfile
	if err := model.DB.Find(&profiles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list profiles"})
		return
	}

	// Mask API keys for non-admin users
	user := getUser(c)
	isAdmin := user != nil && rbac.UserHasRole(*user, model.DefaultAdminRole.Name)

	if !isAdmin {
		for i := range profiles {
			if profiles[i].APIKey != "" {
				profiles[i].APIKey = "***"
			}
		}
	}

	c.JSON(http.StatusOK, profiles)
}

func CreateAIProfile(c *gin.Context) {
	var profile model.AIProviderProfile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := model.DB.Create(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile"})
		return
	}
	c.JSON(http.StatusOK, profile)
}

func UpdateAIProfile(c *gin.Context) {
	id := c.Param("id")
	var profile model.AIProviderProfile
	if err := model.DB.First(&profile, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If this profile is being set as system, unset all others
	if profile.IsSystem {
		if err := model.DB.Model(&model.AIProviderProfile{}).Where("id <> ?", profile.ID).Update("is_system", false).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update system profiles"})
			return
		}
	}

	if err := model.DB.Save(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}
	c.JSON(http.StatusOK, profile)
}

func DeleteAIProfile(c *gin.Context) {
	id := c.Param("id")
	if err := model.DB.Delete(&model.AIProviderProfile{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete profile"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func ToggleAIProfile(c *gin.Context) {
	id := c.Param("id")
	var profile model.AIProviderProfile
	if err := model.DB.First(&profile, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	profile.IsEnabled = !profile.IsEnabled
	if err := model.DB.Save(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle profile"})
		return
	}
	c.JSON(http.StatusOK, profile)
}

// --- Admin Global Config ---

func GetAdminAIConfig(c *gin.Context) {
	// 1. App Level Configs
	allowKeys, _ := model.GetAppConfig(model.CurrentApp.ID, model.AIAllowUserKeys)
	forceKeys, _ := model.GetAppConfig(model.CurrentApp.ID, model.AIForceUserKeys)

	// 2. Global System Settings (Profile with IsSystem = true)
	var profile model.AIProviderProfile
	model.DB.Where("is_system = ?", true).First(&profile)

	// For backward compatibility or if we still want to return a "settings" like object
	// we can construct it if needed, or just return the profile.
	// But let's return it separately for clarity.

	// Mask system profile key if not admin
	user := getUser(c)
	isAdmin := user != nil && rbac.UserHasRole(*user, model.DefaultAdminRole.Name)

	if !isAdmin && profile.APIKey != "" {
		profile.APIKey = "***"
	}

	c.JSON(http.StatusOK, gin.H{
		"allow_user_keys": allowKeys.Value,
		"force_user_keys": forceKeys.Value,
		"system_profile":  profile,
	})
}

func UpdateAIGovernance(c *gin.Context) {
	var input struct {
		AllowUserKeys string `json:"allow_user_keys"`
		ForceUserKeys string `json:"force_user_keys"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update AppConfigs
	model.SetAppConfig(model.CurrentApp.ID, model.AIAllowUserKeys, input.AllowUserKeys)
	model.SetAppConfig(model.CurrentApp.ID, model.AIForceUserKeys, input.ForceUserKeys)

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}
