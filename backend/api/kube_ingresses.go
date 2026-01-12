package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetIngresses lists ingresses for a given namespace and context
func GetIngresses(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	if ns == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace required"})
		return
	}

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	// Parse namespaces
	namespaces := ParseNamespaces(ns)

	type IngressInfo struct {
		Name      string   `json:"name"`
		Hosts     []string `json:"hosts"`
		IPs       []string `json:"ips"`
		Namespace string   `json:"namespace"`
		Age       string   `json:"age"`
	}

	var ingresses []IngressInfo

	for _, singleNs := range namespaces {
		// Note: singleNs can be empty if searchAll is true

		// Try both networking.k8s.io/v1 and beta versions if needed, but v1 is standard now
		list, err := clientset.NetworkingV1().Ingresses(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, ing := range list.Items {
			var hosts []string
			for _, rule := range ing.Spec.Rules {
				if rule.Host != "" {
					hosts = append(hosts, rule.Host)
				}
			}

			var ips []string
			for _, lb := range ing.Status.LoadBalancer.Ingress {
				if lb.IP != "" {
					ips = append(ips, lb.IP)
				}
				if lb.Hostname != "" {
					ips = append(ips, lb.Hostname)
				}
			}

			ingresses = append(ingresses, IngressInfo{
				Name:      ing.Name,
				Hosts:     hosts,
				IPs:       ips,
				Namespace: ing.Namespace,
				Age:       ing.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"ingresses": ingresses})
}
