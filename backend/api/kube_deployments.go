package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetDeployments lists deployments for a given namespace and context
func GetDeployments(c *gin.Context) {
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

	type DeploymentInfo struct {
		Name              string `json:"name"`
		Namespace         string `json:"namespace"`
		Replicas          int32  `json:"replicas"`
		ReadyReplicas     int32  `json:"ready_replicas"`
		AvailableReplicas int32  `json:"available_replicas"`
		Age               string `json:"age"`
		Selector          string `json:"selector"`
	}

	var deployments []DeploymentInfo

	for _, singleNs := range namespaces {
		list, err := clientset.AppsV1().Deployments(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, d := range list.Items {
			replicas := int32(0)
			if d.Spec.Replicas != nil {
				replicas = *d.Spec.Replicas
			}
			deployments = append(deployments, DeploymentInfo{
				Name:              d.Name,
				Namespace:         d.Namespace,
				Replicas:          replicas,
				ReadyReplicas:     d.Status.ReadyReplicas,
				AvailableReplicas: d.Status.AvailableReplicas,
				Age:               d.CreationTimestamp.Time.Format(time.RFC3339),
				Selector:          metav1.FormatLabelSelector(d.Spec.Selector),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"deployments": deployments})
}
