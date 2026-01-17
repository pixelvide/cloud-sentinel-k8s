package api

import (
	"fmt"
	"net/http"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// GetNodes lists nodes for a given context
func GetNodes(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := GetKubeContext(c)
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
			Age:               node.CreationTimestamp.Time.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"nodes": nodes})
}

// ToggleCordon toggles the unschedulable state of a node
func ToggleCordon(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := GetKubeContext(c)
	nodeName := c.Query("name")

	var input struct {
		Unschedulable bool `json:"unschedulable"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	patch := []byte(fmt.Sprintf(`{"spec":{"unschedulable":%t}}`, input.Unschedulable))
	_, err = clientset.CoreV1().Nodes().Patch(c.Request.Context(), nodeName, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to patch node: " + err.Error()})
		return
	}

	// Record audit log
	action := "Cordon Node"
	if !input.Unschedulable {
		action = "Uncordon Node"
	}
	RecordAuditLog(c, action, gin.H{
		"context":       ctxName,
		"node":          nodeName,
		"unschedulable": input.Unschedulable,
	})

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Node %s updated", nodeName)})
}

// DrainNode cordons and evicts pods from a node
func DrainNode(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := GetKubeContext(c)
	nodeName := c.Query("name")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	// 1. Cordon the node first
	patch := []byte(`{"spec":{"unschedulable":true}}`)
	_, err = clientset.CoreV1().Nodes().Patch(c.Request.Context(), nodeName, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cordon node: " + err.Error()})
		return
	}

	// 2. List pods on this node
	podList, err := clientset.CoreV1().Pods("").List(c.Request.Context(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list pods: " + err.Error()})
		return
	}

	evictedCount := 0
	skippedCount := 0
	var errors []string

	for _, pod := range podList.Items {
		// Skip DaemonSet pods
		isDaemonSet := false
		for _, owner := range pod.OwnerReferences {
			if owner.Kind == "DaemonSet" {
				isDaemonSet = true
				break
			}
		}
		if isDaemonSet {
			skippedCount++
			continue
		}

		// Evict pod
		eviction := &policyv1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
		}
		err := clientset.PolicyV1().Evictions(pod.Namespace).Evict(c.Request.Context(), eviction)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to evict pod %s/%s: %v", pod.Namespace, pod.Name, err))
		} else {
			evictedCount++
		}
	}

	// Record audit log
	RecordAuditLog(c, "Drain Node", gin.H{
		"context": ctxName,
		"node":    nodeName,
		"evicted": evictedCount,
		"skipped": skippedCount,
		"errors":  errors,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Node %s draining started", nodeName),
		"evicted": evictedCount,
		"skipped": skippedCount,
		"errors":  errors,
	})
}
