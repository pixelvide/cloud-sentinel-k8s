package api

import (
	"net/http"
	"os"
	"path/filepath"

	"cloud-sentinel-k8s/db"
	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/tools/clientcmd"
)

// GetContexts lists valid contexts from kubeconfig with display name mappings
func GetContexts(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	kubeconfig := GetUserKubeConfigPath(user.StorageNamespace)
	if _, err := os.Stat(kubeconfig); err != nil {
		// Fallback to global if user has no config
		kubeconfig = os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		}
	}

	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch user's context mappings
	var mappings []models.K8sClusterContextMapping
	db.DB.Where("user_id = ?", user.ID).Find(&mappings)
	mappingMap := make(map[string]string)
	for _, m := range mappings {
		mappingMap[m.ContextName] = m.DisplayName
	}

	type ContextInfo struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
	}

	var contexts []ContextInfo
	for name := range config.Contexts {
		displayName := name
		if dn, ok := mappingMap[name]; ok {
			displayName = dn
		}
		contexts = append(contexts, ContextInfo{Name: name, DisplayName: displayName})
	}
	c.JSON(http.StatusOK, gin.H{"contexts": contexts, "current": config.CurrentContext})
}
