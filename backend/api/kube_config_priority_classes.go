package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetPriorityClasses lists priority classes for a given context (cluster-wide)
func GetPriorityClasses(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := GetKubeContext(c)

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	list, err := clientset.SchedulingV1().PriorityClasses().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type PCInfo struct {
		Name  string `json:"name"`
		Value int32  `json:"value"`
		Age   string `json:"age"`
	}

	var pcs []PCInfo
	for _, item := range list.Items {
		pcs = append(pcs, PCInfo{
			Name:  item.Name,
			Value: item.Value,
			Age:   item.CreationTimestamp.Time.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"priorityclasses": pcs})
}
