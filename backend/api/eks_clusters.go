package api

import (
	"cloud-sentinel-k8s/pkg/models"
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/gin-gonic/gin"
)

type ValidatedContext struct {
	AccountID string
}

type EKSClusterInfo struct {
	Name      string `json:"name"`
	Region    string `json:"region"`
	AccountID string `json:"account_id"`
}

type ListEKSClustersRequest struct {
	AWSConfigID uint   `json:"aws_config_id" binding:"required"`
	Region      string `json:"region"` // Optional override? Or strictly force config region? Let's allow optional override if needed, but default to config.
}

type ImportEKSClustersRequest struct {
	AWSConfigID  uint     `json:"aws_config_id" binding:"required"`
	Region       string   `json:"region"` // Optional override
	ClusterNames []string `json:"cluster_names" binding:"required"`
}

// Helper to get config from DB and load AWS config
func getAWSConfigFromDB(ctx context.Context, awsConfigID uint, userID uint, regionOverride string) (aws.Config, *models.AWSConfig, error) {
	var awsCfg models.AWSConfig
	if err := models.DB.Where("id = ? AND user_id = ?", awsConfigID, userID).First(&awsCfg).Error; err != nil {
		return aws.Config{}, nil, err
	}

	targetRegion := awsCfg.Region
	if regionOverride != "" {
		targetRegion = regionOverride
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(targetRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			awsCfg.AccessKeyID,
			awsCfg.SecretAccessKey,
			awsCfg.SessionToken,
		)),
	)
	if err != nil {
		return aws.Config{}, nil, err
	}

	return cfg, &awsCfg, nil
}

func validateAndGetIdentity(ctx context.Context, cfg aws.Config) (*ValidatedContext, error) {
	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}
	return &ValidatedContext{
		AccountID: *identity.Account,
	}, nil
}

func ListEKSClusters(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req ListEKSClustersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Load AWS Config from DB
	cfg, _, err := getAWSConfigFromDB(c.Request.Context(), req.AWSConfigID, user.ID, req.Region)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load aws config: " + err.Error()})
		return
	}

	identity, err := validateAndGetIdentity(c.Request.Context(), cfg)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials in stored config: " + err.Error()})
		return
	}

	eksClient := eks.NewFromConfig(cfg)
	input := &eks.ListClustersInput{}
	var clusterNames []string

	paginator := eks.NewListClustersPaginator(eksClient, input)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list clusters: " + err.Error()})
			return
		}
		clusterNames = append(clusterNames, output.Clusters...)
	}

	var results []EKSClusterInfo
	for _, name := range clusterNames {
		results = append(results, EKSClusterInfo{
			Name:      name,
			Region:    cfg.Region, // Use the region from the loaded config
			AccountID: identity.AccountID,
		})
	}

	c.JSON(http.StatusOK, gin.H{"clusters": results})
}

func ImportEKSClusters(c *gin.Context) {
	user, exists := c.MustGet("user").(*models.User)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req ImportEKSClustersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg, awsConfig, err := getAWSConfigFromDB(c.Request.Context(), req.AWSConfigID, user.ID, req.Region)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load aws config: " + err.Error()})
		return
	}

	identity, err := validateAndGetIdentity(c.Request.Context(), cfg)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials: " + err.Error()})
		return
	}

	eksClient := eks.NewFromConfig(cfg)

	// Process each cluster
	var importedCount int
	for _, clusterName := range req.ClusterNames {
		// Get Cluster Details
		desc, err := eksClient.DescribeCluster(c.Request.Context(), &eks.DescribeClusterInput{
			Name: aws.String(clusterName),
		})
		if err != nil {
			continue
		}

		cluster := desc.Cluster
		if cluster.Endpoint == nil || cluster.CertificateAuthority == nil || cluster.CertificateAuthority.Data == nil {
			continue
		}

		// Save to DB (EKSCluster model)
		newCluster := models.EKSCluster{
			UserID:                   user.ID,
			AWSConfigID:              awsConfig.ID,
			Name:                     clusterName,
			Region:                   cfg.Region,
			AccountID:                identity.AccountID,
			Endpoint:                 *cluster.Endpoint,
			CertificateAuthorityData: *cluster.CertificateAuthority.Data,
		}

		// Check if exists
		var existing models.EKSCluster
		if err := models.DB.Where("user_id = ? AND aws_config_id = ? AND name = ?", user.ID, awsConfig.ID, clusterName).First(&existing).Error; err == nil {
			// Update
			existing.Endpoint = newCluster.Endpoint
			existing.CertificateAuthorityData = newCluster.CertificateAuthorityData
			existing.Region = newCluster.Region
			models.DB.Save(&existing)
		} else {
			// Create
			models.DB.Create(&newCluster)
		}
		importedCount++
	}

	if importedCount > 0 {
		if err := SyncKubeConfigs(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "saved to db but failed to sync to disk: " + err.Error()})
			return
		}
	}

	RecordAuditLog(c, "IMPORT_EKS_CLUSTERS", gin.H{"count": importedCount, "account": identity.AccountID})
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Imported %d clusters", importedCount), "imported_count": importedCount})
}
