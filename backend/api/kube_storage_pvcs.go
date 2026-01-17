package api

import (
	"net/http"
	"strings"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetPVCs lists PersistentVolumeClaims for a given namespace and context
func GetPVCs(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := GetKubeContext(c)

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type PVCInfo struct {
		Name         string `json:"name"`
		Namespace    string `json:"namespace"`
		Status       string `json:"status"`
		Volume       string `json:"volume"`
		Capacity     string `json:"capacity"`
		AccessModes  string `json:"access_modes"`
		StorageClass string `json:"storage_class"`
		Age          string `json:"age"`
	}

	var pvcs []PVCInfo

	for _, singleNs := range namespaces {
		list, err := clientset.CoreV1().PersistentVolumeClaims(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			capacity := ""
			if val, ok := item.Status.Capacity["storage"]; ok {
				capacity = val.String()
			}

			var modes []string
			for _, m := range item.Spec.AccessModes {
				modes = append(modes, string(m))
			}

			class := ""
			if item.Spec.StorageClassName != nil {
				class = *item.Spec.StorageClassName
			}

			pvcs = append(pvcs, PVCInfo{
				Name:         item.Name,
				Namespace:    item.Namespace,
				Status:       string(item.Status.Phase),
				Volume:       item.Spec.VolumeName,
				Capacity:     capacity,
				AccessModes:  strings.Join(modes, ","),
				StorageClass: class,
				Age:          item.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"pvcs": pvcs})
}
