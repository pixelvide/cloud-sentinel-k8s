package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetServices lists services for a given namespace and context
func GetServices(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	// ns optional, empty means all namespaces

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	// Parse namespaces
	namespaces := ParseNamespaces(ns)

	type ServiceInfo struct {
		Name        string   `json:"name"`
		Type        string   `json:"type"`
		ClusterIP   string   `json:"cluster_ip"`
		ExternalIPs []string `json:"external_ips"`
		Ports       []string `json:"ports"`
		Age         string   `json:"age"`
		Namespace   string   `json:"namespace"`
	}

	var services []ServiceInfo

	for _, singleNs := range namespaces {
		list, err := clientset.CoreV1().Services(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, s := range list.Items {
			var ports []string
			for _, p := range s.Spec.Ports {
				ports = append(ports, string(p.Protocol)+":"+p.Name)
			}

			services = append(services, ServiceInfo{
				Name:        s.Name,
				Type:        string(s.Spec.Type),
				ClusterIP:   s.Spec.ClusterIP,
				ExternalIPs: s.Spec.ExternalIPs,
				Ports:       ports,
				Age:         s.CreationTimestamp.Time.Format(time.RFC3339),
				Namespace:   s.Namespace,
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"services": services})
}
