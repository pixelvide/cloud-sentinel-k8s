package api

import (
	"cloud-sentinel-k8s/pkg/models"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func UpdateKubeConfig(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input struct {
		Config string `json:"config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update or Create Default Config in DB
	var config models.KubeConfig
	err := models.DB.Where("user_id = ? AND is_default = ?", user.ID, true).First(&config).Error
	if err != nil {
		// Create new default
		config = models.KubeConfig{
			UserID:    user.ID,
			Name:      "Default",
			Content:   input.Config,
			IsDefault: true,
		}
		if err := models.DB.Create(&config).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save kubeconfig to db"})
			return
		}
	} else {
		// Update existing
		config.Content = input.Config
		if err := models.DB.Save(&config).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update kubeconfig in db"})
			return
		}
	}

	// Sync to disk
	if err := SyncKubeConfigs(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sync kubeconfig to disk"})
		return
	}

	// Record Audit Log
	RecordAuditLog(c, "UPDATE_KUBE_CONFIG", nil)

	c.JSON(http.StatusOK, gin.H{"message": "kubeconfig saved successfully"})
}

func GetKubeConfig(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Try to get from DB first
	var config models.KubeConfig
	err := models.DB.Where("user_id = ? AND is_default = ?", user.ID, true).First(&config).Error
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"config": config.Content})
		return
	}

	// If not in DB, check file (Migration path)
	userConfigPath := GetUserKubeConfigPath(user.StorageNamespace)
	content, err := os.ReadFile(userConfigPath)
	if err == nil {
		// Migrate to DB
		newConfig := models.KubeConfig{
			UserID:    user.ID,
			Name:      "Default",
			Content:   string(content),
			IsDefault: true,
		}
		models.DB.Create(&newConfig)
		c.JSON(http.StatusOK, gin.H{"config": string(content)})
		return
	}

	// Not found anywhere
	c.JSON(http.StatusOK, gin.H{"config": ""})
}
func ValidateKubeConfig(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Use GetClientInfo with storage namespace to load the stored config
	clientset, _, err := GetClientInfo(user.StorageNamespace, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"valid": false, "error": "Failed to load config: " + err.Error()})
		return
	}

	// Try a simple operation: List namespaces
	list, err := clientset.CoreV1().Namespaces().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"valid": false, "error": "Connectivity check failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"message": "Successfully connected to cluster. Found " + strconv.Itoa(len(list.Items)) + " namespaces.",
	})
}

// SetCurrentContext updates the current-context in the user's kubeconfig
func SetCurrentContext(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input struct {
		Context string `json:"context" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userConfigPath := GetUserKubeConfigPath(user.StorageNamespace)

	// Read existing config
	content, err := os.ReadFile(userConfigPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read kubeconfig"})
		return
	}

	// Use clientcmd to properly update
	config, err := clientcmd.Load(content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse kubeconfig"})
		return
	}

	// Verify context exists
	if _, ok := config.Contexts[input.Context]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "context not found in kubeconfig"})
		return
	}

	config.CurrentContext = input.Context

	// Write back
	if err := clientcmd.WriteToFile(*config, userConfigPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save kubeconfig"})
		return
	}

	// Record Audit Log
	RecordAuditLog(c, "SET_CURRENT_CONTEXT", gin.H{"context": input.Context})

	c.JSON(http.StatusOK, gin.H{"message": "current context updated", "context": input.Context})
}
