package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetPVCs lists PersistentVolumeClaims for a given namespace and context
func GetPVCs(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

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
				AccessModes:  joinStrings(modes, ","),
				StorageClass: class,
				Age:          item.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"pvcs": pvcs})
}

// GetPVs lists cluster-wide PersistentVolumes
func GetPVs(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := c.Query("context")

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
			AccessModes:   joinStrings(modes, ","),
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

// GetStorageClasses lists cluster-wide StorageClasses
func GetStorageClasses(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	list, err := clientset.StorageV1().StorageClasses().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type StorageClassInfo struct {
		Name              string `json:"name"`
		Provisioner       string `json:"provisioner"`
		ReclaimPolicy     string `json:"reclaim_policy"`
		VolumeBindingMode string `json:"volume_binding_mode"`
		Age               string `json:"age"`
	}

	var classes []StorageClassInfo
	for _, item := range list.Items {
		reclaim := ""
		if item.ReclaimPolicy != nil {
			reclaim = string(*item.ReclaimPolicy)
		}
		binding := ""
		if item.VolumeBindingMode != nil {
			binding = string(*item.VolumeBindingMode)
		}

		classes = append(classes, StorageClassInfo{
			Name:              item.Name,
			Provisioner:       item.Provisioner,
			ReclaimPolicy:     reclaim,
			VolumeBindingMode: binding,
			Age:               item.CreationTimestamp.Time.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"storageclasses": classes})
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	res := strs[0]
	for i := 1; i < len(strs); i++ {
		res += sep + strs[i]
	}
	return res
}
