package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetEndpoints lists endpoints for a given namespace and context
func GetEndpoints(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := GetKubeContext(c)

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type EndpointInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Endpoints string `json:"endpoints"`
		Age       string `json:"age"`
	}

	var endpoints []EndpointInfo

	for _, singleNs := range namespaces {
		list, err := clientset.CoreV1().Endpoints(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			// Format endpoints briefly (e.g., "10.1.2.3:80, 10.1.2.4:80")
			var epStr string
			for _, subset := range item.Subsets {
				for _, addr := range subset.Addresses {
					for _, port := range subset.Ports {
						if epStr != "" {
							epStr += ", "
						}
						epStr += addr.IP + ":" + string(port.Port)
					}
				}
			}

			endpoints = append(endpoints, EndpointInfo{
				Name:      item.Name,
				Namespace: item.Namespace,
				Endpoints: epStr,
				Age:       item.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"endpoints": endpoints})
}
