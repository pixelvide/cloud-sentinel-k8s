package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetEndpoints lists endpoints for a given namespace and context
func GetEndpoints(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type EndpointInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Endpoints string `json:"endpoints"`
		Age       string `json:"age"`
	}

	var endpoints []EndpointInfo

	for _, singleNs := range namespaces {
		list, err := clientset.CoreV1().Endpoints(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			// Format endpoints briefly (e.g., "10.1.2.3:80, 10.1.2.4:80")
			var epStr string
			for _, subset := range item.Subsets {
				for _, addr := range subset.Addresses {
					for _, port := range subset.Ports {
						if epStr != "" {
							epStr += ", "
						}
						epStr += addr.IP + ":" + string(port.Port)
					}
				}
			}

			endpoints = append(endpoints, EndpointInfo{
				Name:      item.Name,
				Namespace: item.Namespace,
				Endpoints: epStr,
				Age:       item.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"endpoints": endpoints})
}

// GetIngressClasses lists cluster-wide ingress classes
func GetIngressClasses(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := c.Query("context")

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

// GetNetworkPolicies lists network policies for a given namespace and context
func GetNetworkPolicies(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type NetworkPolicyInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Age       string `json:"age"`
	}

	var policies []NetworkPolicyInfo

	for _, singleNs := range namespaces {
		list, err := clientset.NetworkingV1().NetworkPolicies(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			policies = append(policies, NetworkPolicyInfo{
				Name:      item.Name,
				Namespace: item.Namespace,
				Age:       item.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"networkpolicies": policies})
}

// GetPortForwards returns a placeholder list of port forwards
func GetPortForwards(c *gin.Context) {
	// For now, return an empty list as we haven't implemented port forwarding management yet
	c.JSON(http.StatusOK, gin.H{"portforwards": []interface{}{}})
}
