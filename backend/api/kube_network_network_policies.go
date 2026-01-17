package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetNetworkPolicies lists network policies for a given namespace and context
func GetNetworkPolicies(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := GetKubeContext(c)

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type NetworkPolicyInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Age       string `json:"age"`
	}

	var policies []NetworkPolicyInfo

	for _, singleNs := range namespaces {
		list, err := clientset.NetworkingV1().NetworkPolicies(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			policies = append(policies, NetworkPolicyInfo{
				Name:      item.Name,
				Namespace: item.Namespace,
				Age:       item.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"networkpolicies": policies})
}
