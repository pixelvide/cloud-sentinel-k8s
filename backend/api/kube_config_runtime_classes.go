package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetRuntimeClasses lists runtime classes for a given context (cluster-wide)
func GetRuntimeClasses(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := GetKubeContext(c)

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	list, err := clientset.NodeV1().RuntimeClasses().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type RCInfo struct {
		Name    string `json:"name"`
		Handler string `json:"handler"`
		Age     string `json:"age"`
	}

	var rcs []RCInfo
	for _, item := range list.Items {
		rcs = append(rcs, RCInfo{
			Name:    item.Name,
			Handler: item.Handler,
			Age:     item.CreationTimestamp.Time.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"runtimeclasses": rcs})
}
