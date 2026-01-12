package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetStatefulSets lists statefulsets for a given namespace and context
func GetStatefulSets(c *gin.Context) {
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

	type StatefulSetInfo struct {
		Name            string `json:"name"`
		Namespace       string `json:"namespace"`
		Replicas        int32  `json:"replicas"`
		ReadyReplicas   int32  `json:"ready_replicas"`
		CurrentReplicas int32  `json:"current_replicas"`
		UpdatedReplicas int32  `json:"updated_replicas"`
		Age             string `json:"age"`
		Selector        string `json:"selector"`
	}

	var statefulsets []StatefulSetInfo

	for _, singleNs := range namespaces {
		// Note: singleNs can be empty if searchAll is true

		list, err := clientset.AppsV1().StatefulSets(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, ss := range list.Items {
			replicas := int32(0)
			if ss.Spec.Replicas != nil {
				replicas = *ss.Spec.Replicas
			}
			statefulsets = append(statefulsets, StatefulSetInfo{
				Name:            ss.Name,
				Namespace:       ss.Namespace,
				Replicas:        replicas,
				ReadyReplicas:   ss.Status.ReadyReplicas,
				CurrentReplicas: ss.Status.CurrentReplicas,
				UpdatedReplicas: ss.Status.UpdatedReplicas,
				Age:             ss.CreationTimestamp.Time.Format(time.RFC3339),
				Selector:        metav1.FormatLabelSelector(ss.Spec.Selector),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"statefulsets": statefulsets})
}
