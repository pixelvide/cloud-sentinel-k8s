package api

import (
	"fmt"
	"net/http"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetPods lists pods for a given namespace and context
func GetPods(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := GetKubeContext(c)

	// ns optional, empty means all namespaces

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	// Parse namespaces
	namespaces := ParseNamespaces(ns)

	type PodInfo struct {
		Name           string   `json:"name"`
		Containers     []string `json:"containers"`
		InitContainers []string `json:"init_containers"`
		Status         string   `json:"status"`
		Namespace      string   `json:"namespace"`
		Age            string   `json:"age"`
		QoS            string   `json:"qos"`
		Ready          string   `json:"ready"`
		Restarts       int32    `json:"restarts"`
		Node           string   `json:"node"`
		IP             string   `json:"ip"`
		ControlledBy   string   `json:"controlled_by,omitempty"`
	}

	var pods []PodInfo

	selector := c.Query("selector")
	fieldSelector := c.Query("fieldSelector")

	for _, singleNs := range namespaces {
		// Note: singleNs can be empty if searchAll is true
		listOpts := metav1.ListOptions{}
		if selector != "" {
			listOpts.LabelSelector = selector
		}
		if fieldSelector != "" {
			listOpts.FieldSelector = fieldSelector
		}

		list, err := clientset.CoreV1().Pods(singleNs).List(c.Request.Context(), listOpts)
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

			// Calculate ready/restarts
			readyCount := 0
			totalCount := len(p.Spec.Containers)
			var restarts int32 = 0

			for _, cs := range p.Status.ContainerStatuses {
				if cs.Ready {
					readyCount++
				}
				restarts += cs.RestartCount
			}

			controlledBy := ""
			if len(p.OwnerReferences) > 0 {
				controlledBy = fmt.Sprintf("%s/%s", p.OwnerReferences[0].Kind, p.OwnerReferences[0].Name)
			}

			pods = append(pods, PodInfo{
				Name:           p.Name,
				Containers:     containers,
				InitContainers: initContainers,
				Status:         string(p.Status.Phase),
				Namespace:      p.Namespace,
				Age:            p.CreationTimestamp.Time.Format(time.RFC3339),
				QoS:            string(p.Status.QOSClass),
				Ready:          fmt.Sprintf("%d/%d", readyCount, totalCount),
				Restarts:       restarts,
				Node:           p.Spec.NodeName,
				IP:             p.Status.PodIP,
				ControlledBy:   controlledBy,
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"pods": pods})
}
