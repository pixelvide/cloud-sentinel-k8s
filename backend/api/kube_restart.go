package api

import (
	"fmt"
	"net/http"
	"time"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RolloutRestartResource handles triggering a rollout restart for Deployment, DaemonSet, and StatefulSet
func RolloutRestartResource(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := c.Query("context")
	namespace := c.Query("namespace")
	name := c.Query("name")
	kind := c.Query("kind")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	ctx := c.Request.Context()
	restartAt := time.Now().Format(time.RFC3339)

	switch kind {
	case "Deployment":
		deployment, err := clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Deployment: " + err.Error()})
			return
		}

		if deployment.Spec.Template.Annotations == nil {
			deployment.Spec.Template.Annotations = make(map[string]string)
		}
		deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = restartAt

		_, err = clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restart Deployment: " + err.Error()})
			return
		}

	case "DaemonSet":
		daemonset, err := clientset.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get DaemonSet: " + err.Error()})
			return
		}

		if daemonset.Spec.Template.Annotations == nil {
			daemonset.Spec.Template.Annotations = make(map[string]string)
		}
		daemonset.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = restartAt

		_, err = clientset.AppsV1().DaemonSets(namespace).Update(ctx, daemonset, metav1.UpdateOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restart DaemonSet: " + err.Error()})
			return
		}

	case "StatefulSet":
		statefulset, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get StatefulSet: " + err.Error()})
			return
		}

		if statefulset.Spec.Template.Annotations == nil {
			statefulset.Spec.Template.Annotations = make(map[string]string)
		}
		statefulset.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = restartAt

		_, err = clientset.AppsV1().StatefulSets(namespace).Update(ctx, statefulset, metav1.UpdateOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restart StatefulSet: " + err.Error()})
			return
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unsupported kind for rollout restart: %s", kind)})
		return
	}

	// Record audit log
	RecordAuditLog(c, "Rollout Restart Resource", gin.H{
		"context":   ctxName,
		"namespace": namespace,
		"name":      name,
		"kind":      kind,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%s %s restart triggered", kind, name),
	})
}
