package api

import (
	"context"
	"net/http"

	"cloud-sentinel-k8s/db"
	"cloud-sentinel-k8s/pkg/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/gin-gonic/gin"
)

type ValidationRequest struct {
	AccessKeyID     string `json:"access_key_id" binding:"required"`
	SecretAccessKey string `json:"secret_access_key" binding:"required"`
	SessionToken    string `json:"session_token"`
	Region          string `json:"region" binding:"required"`
}

type CreateAWSConfigRequest struct {
	Name            string `json:"name" binding:"required"`
	AccessKeyID     string `json:"access_key_id" binding:"required"`
	SecretAccessKey string `json:"secret_access_key" binding:"required"`
	SessionToken    string `json:"session_token"`
	Region          string `json:"region" binding:"required"`
}

func getTempAWSConfig(ctx context.Context, req ValidationRequest) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx,
		config.WithRegion(req.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			req.AccessKeyID,
			req.SecretAccessKey,
			req.SessionToken,
		)),
	)
}

func ListAWSConfigs(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var configs []models.AWSConfig
	// Select only necessary fields (excluding secret key)
	if err := db.DB.Where("user_id = ?", user.ID).Select("id, user_id, name, access_key_id, region, created_at, updated_at").Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch aws configs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

func CreateAWSConfig(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateAWSConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// VALIDATE CREDENTIALS via STS
	// Create a temp validation request matching the helper
	valReq := ValidationRequest{
		AccessKeyID:     req.AccessKeyID,
		SecretAccessKey: req.SecretAccessKey,
		SessionToken:    req.SessionToken,
		Region:          req.Region,
	}

	cfg, err := getTempAWSConfig(c.Request.Context(), valReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid AWS configuration"})
		return
	}

	// Reuse validation logic from eks_clusters.go by duplicating/importing or just re-implementing short validation here
	// Since we are in same package 'api', we can verify identity using existing helper if exported, or just call sts here
	identity, err := validateAndGetIdentity(c.Request.Context(), cfg)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to validate AWS credentials: " + err.Error()})
		return
	}

	// Create Config
	newConfig := models.AWSConfig{
		UserID:          user.ID,
		Name:            req.Name, // or fmt.Sprintf("Account %s", identity.AccountID)
		AccessKeyID:     req.AccessKeyID,
		SecretAccessKey: req.SecretAccessKey,
		SessionToken:    req.SessionToken,
		Region:          req.Region,
	}

	if err := db.DB.Create(&newConfig).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save aws config"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "AWS Config saved",
		"id":         newConfig.ID,
		"account_id": identity.AccountID,
		"name":       newConfig.Name,
	})
}

func DeleteAWSConfig(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Param("id")

	// Delete
	if err := db.DB.Where("id = ? AND user_id = ?", id, user.ID).Delete(&models.AWSConfig{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete aws config"})
		return
	}

	// TODO: Cascading delete of EKS clusters attached to this?
	// For now, let's assume manual cleanup or later implementation of cascade logic
	// But strictly speaking, if we delete the creds, the EKS clusters won't work anyway.

	c.JSON(http.StatusOK, gin.H{"message": "AWS Config deleted"})
}
