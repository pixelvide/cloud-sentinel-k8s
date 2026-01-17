package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetClusterRoleBindings lists cluster-wide cluster role bindings
func GetClusterRoleBindings(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := GetKubeContext(c)

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	list, err := clientset.RbacV1().ClusterRoleBindings().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type BindingInfo struct {
		Name string `json:"name"`
		Role string `json:"role"`
		Age  string `json:"age"`
	}

	var bindings []BindingInfo
	for _, item := range list.Items {
		bindings = append(bindings, BindingInfo{
			Name: item.Name,
			Role: item.RoleRef.Name,
			Age:  item.CreationTimestamp.Time.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"clusterrolebindings": bindings})
}
