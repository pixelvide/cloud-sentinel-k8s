package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/model"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/utils"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

type UpdateUserAWSConfigReq struct {
	CredentialsContent string `json:"credentials_content" binding:"required"`
}

func GetUserAWSConfig(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	u := user.(model.User)

	config, err := model.GetUserAWSConfig(u.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"credentials_content": ""})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

func UpdateUserAWSConfig(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	u := user.(model.User)

	var req UpdateUserAWSConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get or Create UserConfig for StorageNamespace
	userConfig, err := model.GetUserConfig(u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user configuration"})
		return
	}

	var config model.UserAWSConfig
	err = model.DB.Where("user_id = ?", u.ID).First(&config).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create
			config = model.UserAWSConfig{
				UserID:             u.ID,
				CredentialsContent: model.SecretString(req.CredentialsContent),
			}
			if err := model.DB.Create(&config).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		// Update
		config.CredentialsContent = model.SecretString(req.CredentialsContent)
		if err := model.DB.Save(&config).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Write to filesystem
	if err := utils.WriteUserAWSCredentials(userConfig.StorageNamespace, req.CredentialsContent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write credentials file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

func RestoreAWSConfigs() {
	var configs []model.UserAWSConfig
	if err := model.DB.Find(&configs).Error; err != nil {
		klog.Errorf("Failed to fetch user AWS configs for restoration: %v", err)
		return
	}

	userIDs := make([]uint, 0, len(configs))
	for _, config := range configs {
		userIDs = append(userIDs, config.UserID)
	}

	var userConfigs []model.UserConfig
	if err := model.DB.Where("user_id IN ?", userIDs).Find(&userConfigs).Error; err != nil {
		klog.Errorf("Failed to fetch user configs for restoration: %v", err)
	}

	userConfigMap := make(map[uint]model.UserConfig)
	for _, uc := range userConfigs {
		userConfigMap[uc.UserID] = uc
	}

	for _, config := range configs {
		var userConfig *model.UserConfig

		if uc, ok := userConfigMap[config.UserID]; ok {
			userConfig = &uc
		} else {
			var err error
			userConfig, err = model.GetUserConfig(config.UserID)
			if err != nil {
				klog.Errorf("Failed to get user config for user %d during AWS config restoration: %v", config.UserID, err)
				continue
			}
		}

		if err := utils.WriteUserAWSCredentials(userConfig.StorageNamespace, string(config.CredentialsContent)); err != nil {
			klog.Errorf("Failed to restore AWS credentials for user %d: %v", config.UserID, err)
			continue
		}
		klog.Infof("Restored AWS credentials for user %d", config.UserID)
	}
}
