package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetPDBs lists pod disruption budgets for a given namespace and context
func GetPDBs(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := GetKubeContext(c)

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type PDBInfo struct {
		Name       string `json:"name"`
		Namespace  string `json:"namespace"`
		Age        string `json:"age"`
		MinAvail   string `json:"min_available"`
		MaxUnavail string `json:"max_unavailable"`
	}

	var pdbs []PDBInfo

	for _, singleNs := range namespaces {
		list, err := clientset.PolicyV1().PodDisruptionBudgets(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			minAvail := ""
			if item.Spec.MinAvailable != nil {
				minAvail = item.Spec.MinAvailable.String()
			}
			maxUnavail := ""
			if item.Spec.MaxUnavailable != nil {
				maxUnavail = item.Spec.MaxUnavailable.String()
			}
			pdbs = append(pdbs, PDBInfo{
				Name:       item.Name,
				Namespace:  item.Namespace,
				Age:        item.CreationTimestamp.Time.Format(time.RFC3339),
				MinAvail:   minAvail,
				MaxUnavail: maxUnavail,
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"pdbs": pdbs})
}
