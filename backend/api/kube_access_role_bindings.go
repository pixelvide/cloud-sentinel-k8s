package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetRoleBindings lists namespace-scoped role bindings
func GetRoleBindings(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := GetKubeContext(c)

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type BindingInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Role      string `json:"role"`
		Age       string `json:"age"`
	}

	var bindings []BindingInfo

	for _, singleNs := range namespaces {
		list, err := clientset.RbacV1().RoleBindings(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			bindings = append(bindings, BindingInfo{
				Name:      item.Name,
				Namespace: item.Namespace,
				Role:      item.RoleRef.Name,
				Age:       item.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"rolebindings": bindings})
}
