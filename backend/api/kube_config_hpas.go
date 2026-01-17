package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetHPAs lists horizontal pod autoscalers for a given namespace and context
func GetHPAs(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := GetKubeContext(c)

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type HPAInfo struct {
		Name         string `json:"name"`
		Namespace    string `json:"namespace"`
		Age          string `json:"age"`
		Reference    string `json:"reference"`
		MinReplicas  int32  `json:"min_replicas"`
		MaxReplicas  int32  `json:"max_replicas"`
		CurrReplicas int32  `json:"curr_replicas"`
	}

	var hpas []HPAInfo

	for _, singleNs := range namespaces {
		list, err := clientset.AutoscalingV2().HorizontalPodAutoscalers(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			minReplicas := int32(0)
			if item.Spec.MinReplicas != nil {
				minReplicas = *item.Spec.MinReplicas
			}
			hpas = append(hpas, HPAInfo{
				Name:         item.Name,
				Namespace:    item.Namespace,
				Age:          item.CreationTimestamp.Time.Format(time.RFC3339),
				Reference:    item.Spec.ScaleTargetRef.Kind + "/" + item.Spec.ScaleTargetRef.Name,
				MinReplicas:  minReplicas,
				MaxReplicas:  item.Spec.MaxReplicas,
				CurrReplicas: item.Status.CurrentReplicas,
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"hpas": hpas})
}
