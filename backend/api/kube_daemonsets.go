package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetDaemonSets lists daemonsets for a given namespace and context
func GetDaemonSets(c *gin.Context) {
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

	type DaemonSetInfo struct {
		Name             string `json:"name"`
		Namespace        string `json:"namespace"`
		DesiredScheduled int32  `json:"desired_scheduled"`
		CurrentScheduled int32  `json:"current_scheduled"`
		Ready            int32  `json:"ready"`
		UpToDate         int32  `json:"up_to_date"`
		Available        int32  `json:"available"`
		Age              string `json:"age"`
		Selector         string `json:"selector"`
	}

	var daemonsets []DaemonSetInfo

	for _, singleNs := range namespaces {
		// Note: singleNs can be empty if searchAll is true

		list, err := clientset.AppsV1().DaemonSets(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, ds := range list.Items {
			daemonsets = append(daemonsets, DaemonSetInfo{
				Name:             ds.Name,
				Namespace:        ds.Namespace,
				DesiredScheduled: ds.Status.DesiredNumberScheduled,
				CurrentScheduled: ds.Status.CurrentNumberScheduled,
				Ready:            ds.Status.NumberReady,
				UpToDate:         ds.Status.UpdatedNumberScheduled,
				Available:        ds.Status.NumberAvailable,
				Age:              ds.CreationTimestamp.Time.Format(time.RFC3339),
				Selector:         metav1.FormatLabelSelector(ds.Spec.Selector),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"daemonsets": daemonsets})
}
