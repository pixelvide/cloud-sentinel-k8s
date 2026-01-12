package api

import (
	"cloud-sentinel-k8s/models"
	"net/http"
	"os"
	"path/filepath"
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

	// Target path via central helper
	userConfigPath := GetUserKubeConfigPath(user.StorageNamespace)

	// Ensure user directory exists
	if err := os.MkdirAll(filepath.Dir(userConfigPath), 0777); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create config directory"})
		return
	}
	os.Chmod(filepath.Dir(userConfigPath), 0777)

	// Write kubeconfig
	if err := os.WriteFile(userConfigPath, []byte(input.Config), 0666); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save kubeconfig"})
		return
	}
	os.Chmod(userConfigPath, 0666)

	c.JSON(http.StatusOK, gin.H{"message": "kubeconfig saved successfully"})
}

func GetKubeConfig(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userConfigPath := GetUserKubeConfigPath(user.StorageNamespace)

	content, err := os.ReadFile(userConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusOK, gin.H{"config": ""})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read kubeconfig"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"config": string(content)})
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

	c.JSON(http.StatusOK, gin.H{"message": "current context updated", "context": input.Context})
}
