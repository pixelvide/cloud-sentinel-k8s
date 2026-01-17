package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetMutatingWebhooks lists mutating webhook configurations for a given context (cluster-wide)
func GetMutatingWebhooks(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := GetKubeContext(c)

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	list, err := clientset.AdmissionregistrationV1().MutatingWebhookConfigurations().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type WebhookInfo struct {
		Name     string `json:"name"`
		Age      string `json:"age"`
		Webhooks int    `json:"webhooks_count"`
	}

	var hooks []WebhookInfo
	for _, item := range list.Items {
		hooks = append(hooks, WebhookInfo{
			Name:     item.Name,
			Age:      item.CreationTimestamp.Time.Format(time.RFC3339),
			Webhooks: len(item.Webhooks),
		})
	}
	c.JSON(http.StatusOK, gin.H{"mutatingwebhooks": hooks})
}
