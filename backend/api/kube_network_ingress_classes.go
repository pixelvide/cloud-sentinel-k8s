package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetIngressClasses lists cluster-wide ingress classes
func GetIngressClasses(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := GetKubeContext(c)

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	list, err := clientset.NetworkingV1().IngressClasses().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type IngressClassInfo struct {
		Name       string `json:"name"`
		Controller string `json:"controller"`
		Age        string `json:"age"`
	}

	var classes []IngressClassInfo
	for _, item := range list.Items {
		classes = append(classes, IngressClassInfo{
			Name:       item.Name,
			Controller: item.Spec.Controller,
			Age:        item.CreationTimestamp.Time.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"ingressclasses": classes})
}
