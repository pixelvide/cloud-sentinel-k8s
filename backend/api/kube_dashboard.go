package api

import (
	"net/http"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetDashboardSummary returns counts of various resources
func GetDashboardSummary(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := c.Query("context")

	if ctxName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "context required"})
		return
	}

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	type Summary struct {
		Nodes       int `json:"nodes"`
		Namespaces  int `json:"namespaces"`
		Pods        int `json:"pods"`
		Deployments int `json:"deployments"`
		Services    int `json:"services"`
		Ingresses   int `json:"ingresses"`
		Jobs        int `json:"jobs"`
		CronJobs    int `json:"cronjobs"`
	}

	var summary Summary

	// Count Nodes
	nodeList, err := clientset.CoreV1().Nodes().List(c.Request.Context(), metav1.ListOptions{})
	if err == nil {
		summary.Nodes = len(nodeList.Items)
	}

	// Count Namespaces
	nsList, err := clientset.CoreV1().Namespaces().List(c.Request.Context(), metav1.ListOptions{})
	if err == nil {
		summary.Namespaces = len(nsList.Items)
	}

	// Count Pods (all namespaces)
	podList, err := clientset.CoreV1().Pods("").List(c.Request.Context(), metav1.ListOptions{})
	if err == nil {
		summary.Pods = len(podList.Items)
	}

	// Count Deployments (all namespaces)
	deplList, err := clientset.AppsV1().Deployments("").List(c.Request.Context(), metav1.ListOptions{})
	if err == nil {
		summary.Deployments = len(deplList.Items)
	}

	// Count Services (all namespaces)
	svcList, err := clientset.CoreV1().Services("").List(c.Request.Context(), metav1.ListOptions{})
	if err == nil {
		summary.Services = len(svcList.Items)
	}

	// Count Ingresses (all namespaces)
	ingList, err := clientset.NetworkingV1().Ingresses("").List(c.Request.Context(), metav1.ListOptions{})
	if err == nil {
		summary.Ingresses = len(ingList.Items)
	}

	// Count Jobs (all namespaces)
	jobList, err := clientset.BatchV1().Jobs("").List(c.Request.Context(), metav1.ListOptions{})
	if err == nil {
		summary.Jobs = len(jobList.Items)
	}

	// Count CronJobs (all namespaces)
	cronList, err := clientset.BatchV1().CronJobs("").List(c.Request.Context(), metav1.ListOptions{})
	if err == nil {
		summary.CronJobs = len(cronList.Items)
	}

	c.JSON(http.StatusOK, summary)
}
