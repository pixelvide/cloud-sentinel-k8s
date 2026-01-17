package api

import (
	"fmt"
	"net/http"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ScaleResource handles scaling of Deployment, StatefulSet, and ReplicaSet
func ScaleResource(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := GetKubeContext(c)
	namespace := c.Query("namespace")
	name := c.Query("name")
	kind := c.Query("kind")

	var input struct {
		Replicas int32 `json:"replicas"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	ctx := c.Request.Context()
	var currentReplicas int32

	switch kind {
	case "Deployment":
		scale, err := clientset.AppsV1().Deployments(namespace).GetScale(ctx, name, metav1.GetOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get scale: " + err.Error()})
			return
		}
		currentReplicas = scale.Spec.Replicas
		scale.Spec.Replicas = input.Replicas
		_, err = clientset.AppsV1().Deployments(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update scale: " + err.Error()})
			return
		}

	case "StatefulSet":
		scale, err := clientset.AppsV1().StatefulSets(namespace).GetScale(ctx, name, metav1.GetOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get scale: " + err.Error()})
			return
		}
		currentReplicas = scale.Spec.Replicas
		scale.Spec.Replicas = input.Replicas
		_, err = clientset.AppsV1().StatefulSets(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update scale: " + err.Error()})
			return
		}

	case "ReplicaSet":
		scale, err := clientset.AppsV1().ReplicaSets(namespace).GetScale(ctx, name, metav1.GetOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get scale: " + err.Error()})
			return
		}
		currentReplicas = scale.Spec.Replicas
		scale.Spec.Replicas = input.Replicas
		_, err = clientset.AppsV1().ReplicaSets(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update scale: " + err.Error()})
			return
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unsupported kind for scaling: %s", kind)})
		return
	}

	// Record audit log
	RecordAuditLog(c, "Scale Resource", gin.H{
		"context":   ctxName,
		"namespace": namespace,
		"name":      name,
		"kind":      kind,
		"from":      currentReplicas,
		"to":        input.Replicas,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%s %s scaled from %d to %d", kind, name, currentReplicas, input.Replicas),
	})
}
