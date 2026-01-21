package cluster

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/common"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/model"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/rbac"
	"gorm.io/gorm"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func (cm *ClusterManager) GetClusters(c *gin.Context) {
	user := c.MustGet("user").(model.User)
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make([]common.ClusterInfo, 0)

	// We should show clusters that are either in shared map OR in user map for this user
	// But actually, we should iterate over all ENABLED clusters from DB that the user can access
	clusters, err := model.ListClusters()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, cluster := range clusters {
		if !cluster.Enable || !rbac.CanAccessCluster(user, cluster.Name) {
			continue
		}

		info := common.ClusterInfo{
			Name:           cluster.Name,
			IsDefault:      cluster.Name == cm.defaultContext,
			SkipSystemSync: cluster.SkipSystemSync,
		}

		// Check shared client
		if cs, ok := cm.clusters[cluster.Name]; ok {
			info.Version = cs.Version
		} else if userMap, ok := cm.userClients[cluster.Name]; ok {
			// Check user client
			if uc, ok := userMap[user.ID]; ok {
				if uc.ClientSet != nil {
					info.Version = uc.ClientSet.Version
					klog.Infof("GetClusters: Using user version %s for cluster %s (user %d)", info.Version, cluster.Name, user.ID)
				}
				if uc.Error != "" {
					info.Error = uc.Error
				}
			}
		}

		if info.Version == "" && info.Error == "" {
			// Check for shared errors
			if errMsg, ok := cm.errors[cluster.Name]; ok {
				info.Error = errMsg
			}
		}

		result = append(result, info)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	c.JSON(200, result)
}

func (cm *ClusterManager) GetClusterList(c *gin.Context) {
	clusters, err := model.ListClusters()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if user is admin
	user, exists := c.Get("user")
	isAdmin := exists && rbac.UserHasRole(user.(model.User), model.DefaultAdminRole.Name)

	result := make([]gin.H, 0, len(clusters))
	for _, cluster := range clusters {
		config := ""
		if isAdmin {
			config = string(cluster.Config)
		}

		clusterInfo := gin.H{
			"id":             cluster.ID,
			"name":           cluster.Name,
			"description":    cluster.Description,
			"enabled":        cluster.Enable,
			"inCluster":      cluster.InCluster,
			"isDefault":      cluster.IsDefault,
			"prometheusURL":  cluster.PrometheusURL,
			"config":         config,
			"skipSystemSync": cluster.SkipSystemSync,
		}

		if clientSet, exists := cm.clusters[cluster.Name]; exists {
			clusterInfo["version"] = clientSet.Version
		} else if userMap, ok := cm.userClients[cluster.Name]; ok {
			if uc, ok := userMap[user.(model.User).ID]; ok {
				if uc.ClientSet != nil {
					clusterInfo["version"] = uc.ClientSet.Version
				}
				if uc.Error != "" {
					clusterInfo["error"] = uc.Error
				}
			}
		}

		if clusterInfo["version"] == nil && clusterInfo["error"] == nil {
			if errMsg, exists := cm.errors[cluster.Name]; exists {
				clusterInfo["error"] = errMsg
			}
		}

		result = append(result, clusterInfo)
	}

	c.JSON(http.StatusOK, result)
}

func (cm *ClusterManager) CreateCluster(c *gin.Context) {
	var req struct {
		Name           string `json:"name" binding:"required"`
		Description    string `json:"description"`
		Config         string `json:"config"`
		PrometheusURL  string `json:"prometheusURL"`
		InCluster      bool   `json:"inCluster"`
		IsDefault      bool   `json:"isDefault"`
		SkipSystemSync bool   `json:"skipSystemSync"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := model.GetClusterByName(req.Name); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "cluster already exists"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.IsDefault {
		if err := model.ClearDefaultCluster(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	cluster := &model.Cluster{
		Name:           req.Name,
		Description:    req.Description,
		Config:         model.SecretString(req.Config),
		PrometheusURL:  req.PrometheusURL,
		InCluster:      req.InCluster,
		IsDefault:      req.IsDefault,
		SkipSystemSync: req.SkipSystemSync,
		Enable:         true,
	}

	if err := model.AddCluster(cluster); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	syncNow <- struct{}{}

	c.JSON(http.StatusCreated, gin.H{
		"id":      cluster.ID,
		"message": "cluster created successfully",
	})
}

func (cm *ClusterManager) UpdateCluster(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cluster id"})
		return
	}

	var req struct {
		Name           string `json:"name"`
		Description    string `json:"description"`
		Config         string `json:"config"`
		PrometheusURL  string `json:"prometheusURL"`
		InCluster      bool   `json:"inCluster"`
		IsDefault      bool   `json:"isDefault"`
		Enabled        bool   `json:"enabled"`
		SkipSystemSync bool   `json:"skipSystemSync"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cluster, err := model.GetClusterByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if req.IsDefault && !cluster.IsDefault {
		if err := model.ClearDefaultCluster(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	updates := map[string]interface{}{
		"description":      req.Description,
		"prometheus_url":   req.PrometheusURL,
		"in_cluster":       req.InCluster,
		"is_default":       req.IsDefault,
		"enable":           req.Enabled,
		"skip_system_sync": req.SkipSystemSync,
	}

	if req.Name != "" && req.Name != cluster.Name {
		updates["name"] = req.Name
	}

	if req.Config != "" {
		updates["config"] = model.SecretString(req.Config)
	}

	if err := model.UpdateCluster(cluster, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	syncNow <- struct{}{}

	c.JSON(http.StatusOK, gin.H{"message": "cluster updated successfully"})
}

func (cm *ClusterManager) DeleteCluster(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cluster id"})
		return
	}

	cluster, err := model.GetClusterByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if cluster.IsDefault {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete default cluster"})
		return
	}

	if err := model.DeleteCluster(cluster); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	syncNow <- struct{}{}

	c.JSON(http.StatusOK, gin.H{"message": "cluster deleted successfully"})
}

func (cm *ClusterManager) ImportClustersFromKubeconfig(c *gin.Context) {
	var clusterReq common.ImportClustersRequest
	if err := c.ShouldBindJSON(&clusterReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !clusterReq.InCluster && clusterReq.Config == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "config is required when inCluster is false"})
		return
	}

	if clusterReq.InCluster {
		// In-cluster config
		cluster := &model.Cluster{
			Name:        "in-cluster",
			InCluster:   true,
			Description: "Kubernetes in-cluster config",
			IsDefault:   true,
			Enable:      true,
		}
		if err := model.AddCluster(cluster); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		syncNow <- struct{}{}
		// wait for sync to complete
		time.Sleep(1 * time.Second)
		c.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("imported %d clusters successfully", 1)})
		return
	}

	kubeconfig, err := clientcmd.Load([]byte(clusterReq.Config))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	importedCount := ImportClustersFromKubeconfig(kubeconfig)
	syncNow <- struct{}{}
	// wait for sync to complete
	time.Sleep(1 * time.Second)
	c.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("imported %d clusters successfully", importedCount)})
}
