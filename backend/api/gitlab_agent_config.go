package api

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"cloud-sentinel-k8s/db"
	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
)

func ListGitlabAgentConfigs(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var configs []models.GitlabK8sAgentConfig
	if err := db.DB.Preload("GitlabConfig").Where("user_id = ?", user.ID).Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch agent configs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

func CreateGitlabAgentConfig(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input struct {
		GitlabConfigID uint   `json:"gitlab_config_id" binding:"required"`
		AgentID        string `json:"agent_id" binding:"required"`
		AgentRepo      string `json:"agent_repo" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify the GitLab config exists and belongs to the user
	var gitlabConfig models.GitlabConfig
	if err := db.DB.Where("id = ? AND user_id = ?", input.GitlabConfigID, user.ID).First(&gitlabConfig).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "GitLab config not found"})
		return
	}

	config := models.GitlabK8sAgentConfig{
		UserID:         user.ID,
		GitlabConfigID: input.GitlabConfigID,
		AgentID:        input.AgentID,
		AgentRepo:      input.AgentRepo,
	}

	if err := db.DB.Create(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create agent config"})
		return
	}

	c.JSON(http.StatusCreated, config)
}

func DeleteGitlabAgentConfig(c *gin.Context) {
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

	result := db.DB.Where("id = ? AND user_id = ?", id, user.ID).Delete(&models.GitlabK8sAgentConfig{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete agent config"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func ConfigureGitlabAgent(c *gin.Context) {
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

	var agentConfig models.GitlabK8sAgentConfig
	if err := db.DB.Preload("GitlabConfig").Where("id = ? AND user_id = ?", id, user.ID).First(&agentConfig).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent config not found"})
		return
	}

	glabConfigDir := GetUserGlabConfigDir(user.StorageNamespace)
	kubeConfigPath := GetUserKubeConfigPath(user.StorageNamespace)

	// Ensure directories exist with wide permissions for Docker on Mac compatibility
	if err := os.MkdirAll(glabConfigDir, 0777); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create glab config directory"})
		return
	}
	os.Chmod(glabConfigDir, 0777)

	kubeConfigDir := filepath.Dir(kubeConfigPath)
	if err := os.MkdirAll(kubeConfigDir, 0777); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create kube config directory"})
		return
	}
	os.Chmod(kubeConfigDir, 0777)

	// 1. Auth Login (ensure we use the latest token)
	loginCmd := exec.Command("glab", "auth", "login", "--hostname", agentConfig.GitlabConfig.Host, "--token", agentConfig.GitlabConfig.Token)
	loginCmd.Env = append(os.Environ(), "GLAB_CONFIG_DIR="+glabConfigDir)
	if output, err := loginCmd.CombinedOutput(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Login failed: " + string(output)})
		return
	}

	// 2. Update Kubeconfig
	protocol := "https://"
	if !agentConfig.GitlabConfig.IsHTTPS {
		protocol = "http://"
	}
	gitlabHost := protocol + agentConfig.GitlabConfig.Host

	// Note: We use --use-context to make it the active context in the kubeconfig
	// We use --cache-mode=no to avoid keyring issues in a containerized environment
	agentCmd := exec.Command("glab", "cluster", "agent", "update-kubeconfig", "--agent", agentConfig.AgentID, "--repo", agentConfig.AgentRepo, "--use-context", "--cache-mode=no")
	agentCmd.Env = append(os.Environ(),
		"GLAB_CONFIG_DIR="+glabConfigDir,
		"KUBECONFIG="+kubeConfigPath,
		"GITLAB_HOST="+gitlabHost,
	)

	// We want to suppress any prompt and just overwrite/update
	if output, err := agentCmd.CombinedOutput(); err != nil {
		log.Printf("Glab cluster agent update-kubeconfig failed: %s", string(output))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update kubeconfig via glab: " + string(output)})
		return
	}

	// Ensure the final kubeconfig has permissive permissions for internal container robustness
	os.Chmod(kubeConfigPath, 0666)

	// Update configuration status in database
	if err := db.DB.Model(&agentConfig).Update("is_configured", true).Error; err != nil {
		log.Printf("Failed to update agent config status: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Agent configured successfully",
		"host":            agentConfig.GitlabConfig.Host,
		"kubeconfig_path": kubeConfigPath,
	})
}
