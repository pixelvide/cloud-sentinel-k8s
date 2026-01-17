package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetReplicationControllers lists replication controllers for a given namespace and context
func GetReplicationControllers(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := GetKubeContext(c)

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type ReplicationControllerInfo struct {
		Name              string `json:"name"`
		Namespace         string `json:"namespace"`
		Replicas          int32  `json:"replicas"`
		ReadyReplicas     int32  `json:"ready_replicas"`
		AvailableReplicas int32  `json:"available_replicas"`
		Age               string `json:"age"`
		Selector          string `json:"selector"`
	}

	var rcs []ReplicationControllerInfo

	for _, singleNs := range namespaces {
		list, err := clientset.CoreV1().ReplicationControllers(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, rc := range list.Items {
			replicas := int32(0)
			if rc.Spec.Replicas != nil {
				replicas = *rc.Spec.Replicas
			}
			rcs = append(rcs, ReplicationControllerInfo{
				Name:              rc.Name,
				Namespace:         rc.Namespace,
				Replicas:          replicas,
				ReadyReplicas:     rc.Status.ReadyReplicas,
				AvailableReplicas: rc.Status.AvailableReplicas,
				Age:               rc.CreationTimestamp.Time.Format(time.RFC3339),
				Selector:          metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: rc.Spec.Selector}),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"replicationcontrollers": rcs})
}
