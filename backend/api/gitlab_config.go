package api

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"cloud-sentinel-k8s/db"
	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
)

func ListGitlabConfigs(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var configs []models.GitlabConfig
	if err := db.DB.Where("user_id = ?", user.ID).Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch configs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

func CreateGitlabConfig(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input struct {
		Host    string `json:"gitlab_host" binding:"required"`
		Token   string `json:"token" binding:"required"`
		IsHTTPS bool   `json:"is_https"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config := models.GitlabConfig{
		UserID:  user.ID,
		Host:    input.Host,
		Token:   input.Token,
		IsHTTPS: input.IsHTTPS,
	}

	if err := db.DB.Create(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create config"})
		return
	}

	// Record Audit Log
	RecordAuditLog(c, "CREATE_GITLAB_CONFIG", gin.H{"host": config.Host})

	c.JSON(http.StatusCreated, config)
}

func UpdateGitlabConfig(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var input struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var config models.GitlabConfig
	if err := db.DB.Where("id = ? AND user_id = ?", id, user.ID).First(&config).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}

	config.Token = input.Token
	config.IsValidated = false // Reset validation status on token change
	if err := db.DB.Save(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update config"})
		return
	}

	// Record Audit Log
	RecordAuditLog(c, "UPDATE_GITLAB_CONFIG", gin.H{"host": config.Host})

	c.JSON(http.StatusOK, config)
}

func DeleteGitlabConfig(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	result := db.DB.Where("id = ? AND user_id = ?", id, user.ID).Delete(&models.GitlabConfig{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete config"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}

	// Record Audit Log
	RecordAuditLog(c, "DELETE_GITLAB_CONFIG", gin.H{"id": id})

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func ValidateGitlabConfig(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var config models.GitlabConfig
	if err := db.DB.Where("id = ? AND user_id = ?", id, user.ID).First(&config).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}

	// Set GLAB_CONFIG_DIR to user's storage namespace for perfect isolation
	glabConfigDir := GetUserGlabConfigDir(user.StorageNamespace)
	if err := os.MkdirAll(glabConfigDir, 0777); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create config directory"})
		return
	}
	os.Chmod(glabConfigDir, 0777)

	// Command 1: Login
	loginCmd := exec.Command("glab", "auth", "login", "--hostname", config.Host, "--token", config.Token)
	loginCmd.Env = append(os.Environ(), "GLAB_CONFIG_DIR="+glabConfigDir)
	if output, err := loginCmd.CombinedOutput(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"valid": false, "error": "Login failed: " + string(output)})
		return
	}

	// Command 2: Status
	statusCmd := exec.Command("glab", "auth", "status", "--hostname", config.Host)
	statusCmd.Env = append(os.Environ(), "GLAB_CONFIG_DIR="+glabConfigDir)
	if output, err := statusCmd.CombinedOutput(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"valid": false, "error": "Status check failed: " + string(output)})
		return
	}

	// Validation success: persist to DB
	config.IsValidated = true
	if err := db.DB.Save(&config).Error; err != nil {
		log.Printf("Failed to update validation status in DB: %v", err)
		// We still return success to the user as the CLI check passed
	}

	// Record Audit Log
	RecordAuditLog(c, "VALIDATE_GITLAB_CONFIG", gin.H{"host": config.Host})

	c.JSON(http.StatusOK, gin.H{"valid": true, "message": "Credentials verified successfully via glab CLI for " + config.Host})
}
