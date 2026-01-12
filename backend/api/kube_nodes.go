package api

import (
	"net/http"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetNodes lists nodes for a given context
func GetNodes(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := c.Query("context")
	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	list, err := clientset.CoreV1().Nodes().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type NodeInfo struct {
		Name              string            `json:"name"`
		Status            string            `json:"status"`
		Roles             []string          `json:"roles"`
		KubeletVersion    string            `json:"kubelet_version"`
		OS                string            `json:"os"`
		Architecture      string            `json:"architecture"`
		CPUCapacity       string            `json:"cpu_capacity"`
		MemoryCapacity    string            `json:"memory_capacity"`
		CPUAllocatable    string            `json:"cpu_allocatable"`
		MemoryAllocatable string            `json:"memory_allocatable"`
		Labels            map[string]string `json:"labels"`
		Age               string            `json:"age"`
	}

	var nodes []NodeInfo
	for _, node := range list.Items {
		// Determine status
		status := "Unknown"
		for _, cond := range node.Status.Conditions {
			if cond.Type == "Ready" {
				if cond.Status == "True" {
					status = "Ready"
				} else {
					status = "NotReady"
				}
				break
			}
		}

		// Extract roles from labels
		var roles []string
		for key := range node.Labels {
			if key == "node-role.kubernetes.io/master" || key == "node-role.kubernetes.io/control-plane" {
				roles = append(roles, "control-plane")
			} else if key == "node-role.kubernetes.io/worker" {
				roles = append(roles, "worker")
			} else if len(key) > 24 && key[:24] == "node-role.kubernetes.io/" {
				roles = append(roles, key[24:])
			}
		}
		if len(roles) == 0 {
			roles = append(roles, "worker")
		}

		nodes = append(nodes, NodeInfo{
			Name:              node.Name,
			Status:            status,
			Roles:             roles,
			KubeletVersion:    node.Status.NodeInfo.KubeletVersion,
			OS:                node.Status.NodeInfo.OperatingSystem,
			Architecture:      node.Status.NodeInfo.Architecture,
			CPUCapacity:       node.Status.Capacity.Cpu().String(),
			MemoryCapacity:    node.Status.Capacity.Memory().String(),
			CPUAllocatable:    node.Status.Allocatable.Cpu().String(),
			MemoryAllocatable: node.Status.Allocatable.Memory().String(),
			Labels:            node.Labels,
			Age:               node.CreationTimestamp.String(),
		})
	}
	c.JSON(http.StatusOK, gin.H{"nodes": nodes})
}
