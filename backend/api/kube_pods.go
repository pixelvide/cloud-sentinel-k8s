package api

import (
	"net/http"
	"strings"
	"time"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetPods lists pods for a given namespace and context
func GetPods(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	if ns == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace required"})
		return
	}

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	// Split namespaces by comma
	namespaces := strings.Split(ns, ",")

	type PodInfo struct {
		Name           string   `json:"name"`
		Containers     []string `json:"containers"`
		InitContainers []string `json:"init_containers"`
		Status         string   `json:"status"`
		Namespace      string   `json:"namespace"`
		Age            string   `json:"age"`
		QoS            string   `json:"qos"`
	}

	var pods []PodInfo

	for _, singleNs := range namespaces {
		singleNs = strings.TrimSpace(singleNs)
		if singleNs == "" {
			continue
		}

		list, err := clientset.CoreV1().Pods(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, p := range list.Items {
			var containers []string
			var initContainers []string

			for _, cn := range p.Spec.InitContainers {
				initContainers = append(initContainers, cn.Name)
			}
			for _, cn := range p.Spec.Containers {
				containers = append(containers, cn.Name)
			}
			pods = append(pods, PodInfo{
				Name:           p.Name,
				Containers:     containers,
				InitContainers: initContainers,
				Status:         string(p.Status.Phase),
				Namespace:      p.Namespace,
				Age:            p.CreationTimestamp.Time.Format(time.RFC3339),
				QoS:            string(p.Status.QOSClass),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"pods": pods})
}
