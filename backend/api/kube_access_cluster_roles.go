package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetClusterRoles lists cluster-wide cluster roles
func GetClusterRoles(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := GetKubeContext(c)

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	list, err := clientset.RbacV1().ClusterRoles().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type RoleInfo struct {
		Name string `json:"name"`
		Age  string `json:"age"`
	}

	var roles []RoleInfo
	for _, item := range list.Items {
		roles = append(roles, RoleInfo{
			Name: item.Name,
			Age:  item.CreationTimestamp.Time.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"clusterroles": roles})
}
