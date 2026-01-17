package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetStorageClasses lists cluster-wide StorageClasses
func GetStorageClasses(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := GetKubeContext(c)

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
		IsDefault         bool   `json:"is_default"`
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

		isDefault := false
		if item.Annotations != nil {
			if item.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
				isDefault = true
			} else if item.Annotations["storageclass.beta.kubernetes.io/is-default-class"] == "true" {
				isDefault = true
			}
		}

		classes = append(classes, StorageClassInfo{
			Name:              item.Name,
			Provisioner:       item.Provisioner,
			ReclaimPolicy:     reclaim,
			VolumeBindingMode: binding,
			IsDefault:         isDefault,
			Age:               item.CreationTimestamp.Time.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"storageclasses": classes})
}
