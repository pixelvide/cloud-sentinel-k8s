package api

import (
	"cloud-sentinel-k8s/pkg/models"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// SyncKubeConfigs merges all DB records for a user and writes to the disk.
// This includes manually uploaded KubeConfigs AND EKS Clusters.
// This is the "source of truth" -> "artifact" generation step.
func SyncKubeConfigs(user *models.User) error {
	var configs []models.KubeConfig
	if err := models.DB.Where("user_id = ?", user.ID).Find(&configs).Error; err != nil {
		return err
	}

	var eksClusters []models.EKSCluster
	if err := models.DB.Preload("AWSConfig").Where("user_id = ?", user.ID).Find(&eksClusters).Error; err != nil {
		return err
	}

	if len(configs) == 0 && len(eksClusters) == 0 {
		return nil
	}

	// Start with an empty config
	finalConfig := clientcmdapi.NewConfig()

	// 1. Process Manual Configs
	for _, cfg := range configs {
		loaded, err := clientcmd.Load([]byte(cfg.Content))
		if err != nil {
			// Log error but continue with other configs?
			fmt.Printf("Error loading config %s: %v\n", cfg.Name, err)
			continue
		}

		for key, cluster := range loaded.Clusters {
			// Prefix key to avoid collisions if NOT default
			newKey := key
			if !cfg.IsDefault {
				newKey = fmt.Sprintf("%s-%s", cfg.Name, key)
			}
			finalConfig.Clusters[newKey] = cluster
		}

		for key, authInfo := range loaded.AuthInfos {
			newKey := key
			if !cfg.IsDefault {
				newKey = fmt.Sprintf("%s-%s", cfg.Name, key)
			}
			finalConfig.AuthInfos[newKey] = authInfo
		}

		for key, context := range loaded.Contexts {
			newKey := key
			if !cfg.IsDefault {
				newKey = fmt.Sprintf("%s-%s", cfg.Name, key)
			}
			// Update context references
			if !cfg.IsDefault {
				context.Cluster = fmt.Sprintf("%s-%s", cfg.Name, context.Cluster)
				context.AuthInfo = fmt.Sprintf("%s-%s", cfg.Name, context.AuthInfo)
			}
			finalConfig.Contexts[newKey] = context
		}

		// Preserve current context if it's the default config
		if cfg.IsDefault {
			finalConfig.CurrentContext = loaded.CurrentContext
		}
	}

	// 2. Process EKS Clusters
	for _, cluster := range eksClusters {
		configName := fmt.Sprintf("eks-%s-%s", cluster.AccountID, cluster.Name)
		contextName := configName

		// Cluster
		finalConfig.Clusters[configName] = &clientcmdapi.Cluster{
			Server:                   cluster.Endpoint,
			CertificateAuthorityData: []byte(cluster.CertificateAuthorityData), // encoded string in DB? Wait, stored as string but is it base64? Usually yes.
			// However in Import logic we just cast []byte to string. So here we cast string back to []byte. Correct.
		}

		// Auth Info
		execEnv := []clientcmdapi.ExecEnvVar{
			{Name: "AWS_ACCESS_KEY_ID", Value: cluster.AWSConfig.AccessKeyID},
			{Name: "AWS_SECRET_ACCESS_KEY", Value: cluster.AWSConfig.SecretAccessKey},
		}
		if cluster.AWSConfig.SessionToken != "" {
			execEnv = append(execEnv, clientcmdapi.ExecEnvVar{
				Name:  "AWS_SESSION_TOKEN",
				Value: cluster.AWSConfig.SessionToken,
			})
		}

		// Use cluster region if set, otherwise config region
		region := cluster.Region
		if region == "" {
			region = cluster.AWSConfig.Region
		}

		finalConfig.AuthInfos[configName] = &clientcmdapi.AuthInfo{
			Exec: &clientcmdapi.ExecConfig{
				APIVersion: "client.authentication.k8s.io/v1beta1",
				Command:    "aws",
				Args: []string{
					"eks", "get-token",
					"--region", region,
					"--cluster-name", cluster.Name,
					"--output", "json",
				},
				Env: execEnv,
			},
		}

		// Context
		finalConfig.Contexts[contextName] = &clientcmdapi.Context{
			Cluster:  configName,
			AuthInfo: configName,
		}
	}

	// Write to disk
	userConfigPath := GetUserKubeConfigPath(user.StorageNamespace)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(userConfigPath), 0777); err != nil {
		return err
	}

	return clientcmd.WriteToFile(*finalConfig, userConfigPath)
}

func UploadKubeConfig(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Limit upload size (e.g., 2MB)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request: " + err.Error()})
		return
	}
	defer file.Close()

	// Read content
	content := make([]byte, header.Size)
	_, err = file.Read(content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	// Validate it's a valid kubeconfig
	_, err = clientcmd.Load(content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid kubeconfig file"})
		return
	}

	// Save to DB
	configName := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))

	newConfig := models.KubeConfig{
		UserID:    user.ID,
		Name:      configName,
		Content:   string(content),
		IsDefault: false,
	}

	if err := models.DB.Create(&newConfig).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save to database"})
		return
	}

	// Sync
	if err := SyncKubeConfigs(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "saved to db but failed to sync to disk: " + err.Error()})
		return
	}

	// Record Audit
	RecordAuditLog(c, "UPLOAD_KUBE_CONFIG", gin.H{"name": configName})

	c.JSON(http.StatusOK, gin.H{"message": "config uploaded and synced"})
}

func ListKubeConfigs(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var configs []models.KubeConfig
	if err := models.DB.Where("user_id = ?", user.ID).Select("id, name, is_default, created_at").Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch configs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

func DeleteKubeConfig(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Param("id")

	var config models.KubeConfig
	if err := models.DB.Where("id = ? AND user_id = ?", id, user.ID).First(&config).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}

	if config.IsDefault {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete default config"})
		return
	}

	if err := models.DB.Delete(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete config"})
		return
	}

	// Sync
	if err := SyncKubeConfigs(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "deleted from db but failed to sync to disk: " + err.Error()})
		return
	}

	RecordAuditLog(c, "DELETE_KUBE_CONFIG", gin.H{"name": config.Name})

	c.JSON(http.StatusOK, gin.H{"message": "config deleted and synced"})
}
