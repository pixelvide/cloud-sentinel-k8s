package api

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetEvents lists events for a given namespace and context
func GetEvents(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")
	limitStr := c.Query("limit")

	if ns == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace required"})
		return
	}

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := strings.Split(ns, ",")

	type EventInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Type      string `json:"type"`
		Reason    string `json:"reason"`
		Message   string `json:"message"`
		Object    string `json:"object"`
		Count     int32  `json:"count"`
		FirstSeen string `json:"first_seen"`
		LastSeen  string `json:"last_seen"`
	}

	var events []EventInfo

	for _, singleNs := range namespaces {
		singleNs = strings.TrimSpace(singleNs)
		if singleNs == "" {
			continue
		}

		list, err := clientset.CoreV1().Events(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, e := range list.Items {
			events = append(events, EventInfo{
				Name:      e.Name,
				Namespace: e.Namespace,
				Type:      e.Type,
				Reason:    e.Reason,
				Message:   e.Message,
				Object:    e.InvolvedObject.Kind + "/" + e.InvolvedObject.Name,
				Count:     e.Count,
				FirstSeen: e.FirstTimestamp.String(),
				LastSeen:  e.LastTimestamp.String(),
			})
		}
	}

	// Sort by LastSeen descending
	sort.Slice(events, func(i, j int) bool {
		return events[i].LastSeen > events[j].LastSeen
	})

	// Apply limit if specified
	if limitStr != "" {
		var limit int
		if _, err := fmt.Sscanf(limitStr, "%d", &limit); err == nil && limit > 0 && limit < len(events) {
			events = events[:limit]
		}
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}
