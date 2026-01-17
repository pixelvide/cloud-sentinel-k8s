package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetJobs lists jobs for a given namespace and context
func GetJobs(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := GetKubeContext(c)

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

	selector := c.Query("selector")
	fieldSelector := c.Query("fieldSelector")
	ownerUid := c.Query("ownerUid")

	for _, singleNs := range namespaces {
		// Note: singleNs can be empty if searchAll is true
		listOpts := metav1.ListOptions{}
		if selector != "" {
			listOpts.LabelSelector = selector
		}
		if fieldSelector != "" {
			listOpts.FieldSelector = fieldSelector
		}

		list, err := clientset.BatchV1().Jobs(singleNs).List(c.Request.Context(), listOpts)
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, j := range list.Items {
			if ownerUid != "" {
				isOwned := false
				for _, owner := range j.OwnerReferences {
					if string(owner.UID) == ownerUid {
						isOwned = true
						break
					}
				}
				if !isOwned {
					continue
				}
			}

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
