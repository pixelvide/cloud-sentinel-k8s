package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetServiceAccounts lists service accounts for a given namespace and context
func GetServiceAccounts(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := GetKubeContext(c)

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type SAInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Secrets   int    `json:"secrets"`
		Age       string `json:"age"`
	}

	var sas []SAInfo

	for _, singleNs := range namespaces {
		list, err := clientset.CoreV1().ServiceAccounts(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			sas = append(sas, SAInfo{
				Name:      item.Name,
				Namespace: item.Namespace,
				Secrets:   len(item.Secrets),
				Age:       item.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"serviceaccounts": sas})
}
