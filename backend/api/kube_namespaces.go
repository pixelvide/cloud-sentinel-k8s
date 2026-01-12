package api

import (
	"net/http"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetNamespaces lists namespaces for a given context
func GetNamespaces(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := c.Query("context")
	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	list, err := clientset.CoreV1().Namespaces().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var names []string
	for _, ns := range list.Items {
		names = append(names, ns.Name)
	}
	c.JSON(http.StatusOK, gin.H{"namespaces": names})
}
