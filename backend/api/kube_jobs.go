package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetJobs lists jobs for a given namespace and context
func GetJobs(c *gin.Context) {
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

	type JobInfo struct {
		Name        string `json:"name"`
		Namespace   string `json:"namespace"`
		Completions int32  `json:"completions"`
		Succeeded   int32  `json:"succeeded"`
		Failed      int32  `json:"failed"`
		Active      int32  `json:"active"`
		Age         string `json:"age"`
		Selector    string `json:"selector"`
	}

	var jobs []JobInfo

	for _, singleNs := range namespaces {
		// Note: singleNs can be empty if searchAll is true

		list, err := clientset.BatchV1().Jobs(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, j := range list.Items {
			completions := int32(1)
			if j.Spec.Completions != nil {
				completions = *j.Spec.Completions
			}
			jobs = append(jobs, JobInfo{
				Name:        j.Name,
				Namespace:   j.Namespace,
				Completions: completions,
				Succeeded:   j.Status.Succeeded,
				Failed:      j.Status.Failed,
				Active:      j.Status.Active,
				Age:         j.CreationTimestamp.Time.Format(time.RFC3339),
				Selector:    metav1.FormatLabelSelector(j.Spec.Selector),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"jobs": jobs})
}
