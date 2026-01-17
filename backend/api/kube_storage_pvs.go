package api

import (
	"net/http"
	"strings"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetPVs lists cluster-wide PersistentVolumes
func GetPVs(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := GetKubeContext(c)

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	list, err := clientset.CoreV1().PersistentVolumes().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	storageClassFilter := c.Query("storageClass")

	type PVInfo struct {
		Name          string `json:"name"`
		Capacity      string `json:"capacity"`
		AccessModes   string `json:"access_modes"`
		ReclaimPolicy string `json:"reclaim_policy"`
		Status        string `json:"status"`
		Claim         string `json:"claim"`
		StorageClass  string `json:"storage_class"`
		Reason        string `json:"reason"`
		Age           string `json:"age"`
	}

	var pvs []PVInfo
	for _, item := range list.Items {
		if storageClassFilter != "" && item.Spec.StorageClassName != storageClassFilter {
			continue
		}

		capacity := ""
		if val, ok := item.Spec.Capacity["storage"]; ok {
			capacity = val.String()
		}

		var modes []string
		for _, m := range item.Spec.AccessModes {
			modes = append(modes, string(m))
		}

		claim := ""
		if item.Spec.ClaimRef != nil {
			claim = item.Spec.ClaimRef.Namespace + "/" + item.Spec.ClaimRef.Name
		}

		pvs = append(pvs, PVInfo{
			Name:          item.Name,
			Capacity:      capacity,
			AccessModes:   strings.Join(modes, ","),
			ReclaimPolicy: string(item.Spec.PersistentVolumeReclaimPolicy),
			Status:        string(item.Status.Phase),
			Claim:         claim,
			StorageClass:  item.Spec.StorageClassName,
			Reason:        item.Status.Reason,
			Age:           item.CreationTimestamp.Time.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"pvs": pvs})
}
