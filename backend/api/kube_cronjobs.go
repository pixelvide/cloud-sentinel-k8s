package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetCronJobs lists cronjobs for a given namespace and context
func GetCronJobs(c *gin.Context) {
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

	type CronJobInfo struct {
		Name         string `json:"name"`
		Namespace    string `json:"namespace"`
		Schedule     string `json:"schedule"`
		Suspend      bool   `json:"suspend"`
		Active       int    `json:"active"`
		LastSchedule string `json:"last_schedule"`
		Age          string `json:"age"`
	}

	var cronjobs []CronJobInfo

	for _, singleNs := range namespaces {
		// Note: singleNs can be empty if searchAll is true

		list, err := clientset.BatchV1().CronJobs(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, cj := range list.Items {
			suspend := false
			if cj.Spec.Suspend != nil {
				suspend = *cj.Spec.Suspend
			}
			lastSchedule := ""
			if cj.Status.LastScheduleTime != nil {
				lastSchedule = cj.Status.LastScheduleTime.Time.Format(time.RFC3339)
			}
			cronjobs = append(cronjobs, CronJobInfo{
				Name:         cj.Name,
				Namespace:    cj.Namespace,
				Schedule:     cj.Spec.Schedule,
				Suspend:      suspend,
				Active:       len(cj.Status.Active),
				LastSchedule: lastSchedule,
				Age:          cj.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"cronjobs": cronjobs})
}
