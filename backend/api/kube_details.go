package api

import (
	"context"
	"sort"
	"time"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

// SupportedResources defines the kinds supported and their scope (Cluster vs Namespace)
var SupportedResources = map[string]string{
	"Pod":                            "Namespace",
	"Deployment":                     "Namespace",
	"ReplicaSet":                     "Namespace",
	"StatefulSet":                    "Namespace",
	"DaemonSet":                      "Namespace",
	"Job":                            "Namespace",
	"CronJob":                        "Namespace",
	"ReplicationController":          "Namespace",
	"Service":                        "Namespace",
	"Endpoints":                      "Namespace",
	"Ingress":                        "Namespace",
	"NetworkPolicy":                  "Namespace",
	"ConfigMap":                      "Namespace",
	"Secret":                         "Namespace",
	"ResourceQuota":                  "Namespace",
	"LimitRange":                     "Namespace",
	"HorizontalPodAutoscaler":        "Namespace",
	"PodDisruptionBudget":            "Namespace",
	"Lease":                          "Namespace",
	"ServiceAccount":                 "Namespace",
	"Role":                           "Namespace",
	"RoleBinding":                    "Namespace",
	"Node":                           "Cluster",
	"Namespace":                      "Cluster",
	"PersistentVolume":               "Cluster",
	"PersistentVolumeClaim":          "Namespace",
	"StorageClass":                   "Cluster",
	"ClusterRole":                    "Cluster",
	"ClusterRoleBinding":             "Cluster",
	"PriorityClass":                  "Cluster",
	"RuntimeClass":                   "Cluster",
	"IngressClass":                   "Cluster",
	"MutatingWebhookConfiguration":   "Cluster",
	"ValidatingWebhookConfiguration": "Cluster",
	"Event":                          "Namespace",
}

// GetResourceScopes returns the list of supported resources and their scopes
func GetResourceScopes(c *gin.Context) {
	c.JSON(200, gin.H{"scopes": SupportedResources})
}

// GetResourceDetails fetches detailed info (YAML manifest + events) for a resource
func GetResourceDetails(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")
	name := c.Query("name")
	kind := c.Query("kind") // e.g., "Pod", "Deployment", "Service" (Case sensitive matching SupportedResources)

	if name == "" || kind == "" {
		c.JSON(400, gin.H{"error": "name and kind are required"})
		return
	}

	scope, ok := SupportedResources[kind]
	if !ok {
		c.JSON(400, gin.H{"error": "Unsupported kind: " + kind})
		return
	}

	// For namespaced resources, namespace is required
	if scope == "Namespace" && ns == "" && kind != "Namespace" {
		c.JSON(400, gin.H{"error": "namespace is required for " + kind})
		return
	}

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	// 1. Fetch the resource manifest
	var resourceObj interface{}
	var resourceErr error

	switch kind {
	// Workloads
	case "Pod":
		resourceObj, resourceErr = clientset.CoreV1().Pods(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "Deployment":
		resourceObj, resourceErr = clientset.AppsV1().Deployments(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "ReplicaSet":
		resourceObj, resourceErr = clientset.AppsV1().ReplicaSets(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "StatefulSet":
		resourceObj, resourceErr = clientset.AppsV1().StatefulSets(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "DaemonSet":
		resourceObj, resourceErr = clientset.AppsV1().DaemonSets(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "Job":
		resourceObj, resourceErr = clientset.BatchV1().Jobs(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "CronJob":
		resourceObj, resourceErr = clientset.BatchV1().CronJobs(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "ReplicationController":
		resourceObj, resourceErr = clientset.CoreV1().ReplicationControllers(ns).Get(c.Request.Context(), name, metav1.GetOptions{})

	// Network
	case "Service":
		resourceObj, resourceErr = clientset.CoreV1().Services(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "Endpoints":
		resourceObj, resourceErr = clientset.CoreV1().Endpoints(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "Ingress":
		resourceObj, resourceErr = clientset.NetworkingV1().Ingresses(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "NetworkPolicy":
		resourceObj, resourceErr = clientset.NetworkingV1().NetworkPolicies(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "IngressClass":
		resourceObj, resourceErr = clientset.NetworkingV1().IngressClasses().Get(c.Request.Context(), name, metav1.GetOptions{})

	// Config
	case "ConfigMap":
		resourceObj, resourceErr = clientset.CoreV1().ConfigMaps(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "Secret":
		resourceObj, resourceErr = clientset.CoreV1().Secrets(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "ResourceQuota":
		resourceObj, resourceErr = clientset.CoreV1().ResourceQuotas(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "LimitRange":
		resourceObj, resourceErr = clientset.CoreV1().LimitRanges(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "HorizontalPodAutoscaler":
		resourceObj, resourceErr = clientset.AutoscalingV2().HorizontalPodAutoscalers(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "PodDisruptionBudget":
		resourceObj, resourceErr = clientset.PolicyV1().PodDisruptionBudgets(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "Lease":
		resourceObj, resourceErr = clientset.CoordinationV1().Leases(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "PriorityClass":
		resourceObj, resourceErr = clientset.SchedulingV1().PriorityClasses().Get(c.Request.Context(), name, metav1.GetOptions{})
	case "RuntimeClass":
		resourceObj, resourceErr = clientset.NodeV1().RuntimeClasses().Get(c.Request.Context(), name, metav1.GetOptions{})
	case "MutatingWebhookConfiguration":
		resourceObj, resourceErr = clientset.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(c.Request.Context(), name, metav1.GetOptions{})
	case "ValidatingWebhookConfiguration":
		resourceObj, resourceErr = clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(c.Request.Context(), name, metav1.GetOptions{})

	// Storage
	case "PersistentVolume":
		resourceObj, resourceErr = clientset.CoreV1().PersistentVolumes().Get(c.Request.Context(), name, metav1.GetOptions{})
	case "PersistentVolumeClaim":
		resourceObj, resourceErr = clientset.CoreV1().PersistentVolumeClaims(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "StorageClass":
		resourceObj, resourceErr = clientset.StorageV1().StorageClasses().Get(c.Request.Context(), name, metav1.GetOptions{})

	// Access
	case "ServiceAccount":
		resourceObj, resourceErr = clientset.CoreV1().ServiceAccounts(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "ClusterRole":
		resourceObj, resourceErr = clientset.RbacV1().ClusterRoles().Get(c.Request.Context(), name, metav1.GetOptions{})
	case "Role":
		resourceObj, resourceErr = clientset.RbacV1().Roles(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "ClusterRoleBinding":
		resourceObj, resourceErr = clientset.RbacV1().ClusterRoleBindings().Get(c.Request.Context(), name, metav1.GetOptions{})
	case "RoleBinding":
		resourceObj, resourceErr = clientset.RbacV1().RoleBindings(ns).Get(c.Request.Context(), name, metav1.GetOptions{})

	// Cluster Infrastructure
	case "Node":
		resourceObj, resourceErr = clientset.CoreV1().Nodes().Get(c.Request.Context(), name, metav1.GetOptions{})
	case "Namespace":
		resourceObj, resourceErr = clientset.CoreV1().Namespaces().Get(c.Request.Context(), name, metav1.GetOptions{})
	case "Event":
		resourceObj, resourceErr = clientset.CoreV1().Events(ns).Get(c.Request.Context(), name, metav1.GetOptions{})

	default:
		c.JSON(400, gin.H{"error": "Unsupported kind: " + kind})
		return
	}

	if resourceErr != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch resource: " + resourceErr.Error()})
		return
	}

	// Convert to YAML
	// Note: The objects returned by typed clients might miss TypeMeta (Kind/APIVersion) when marshaled.
	// We might need to manually set them or accept they are missing in the view.
	// Or utilize scheme.Scheme to populate.
	// For visualization, missing Kind/APIVersion in the YAML block is a known quirk of client-go typed gets.
	// We can manually patch it for the common ones or use dynamic client.
	// Let's try to set it via reflection or type assertion if critical, or use SetGroupVersionKind if it's a runtime.Object (it is).
	// Simple approach: marshal as is.

	// Attempt to set GVK if missing (common client-go issue)
	if obj, ok := resourceObj.(schema.ObjectKind); ok {
		if obj.GroupVersionKind().Empty() {
			// We can't easily deduce generic GVK without scheme,
			// but we know it from our switch context.
			// Manual fix for common types to make YAML look correct:
			switch kind {
			case "Pod":
				// Need to cast to the concrete struct to set TypeMeta
				// struct does not embed metav1.Object direct? It embeds TypeMeta.
				// Actually corev1.Pod embeds metav1.TypeMeta
				// We use a helper or just let it be.
				// Let's rely on JSON marshaling adding it if present, or just frontend knowing what it is.
			}
		}
	}

	yamlBytes, err := yaml.Marshal(resourceObj)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to marshal resource to YAML: " + err.Error()})
		return
	}

	// 2. Fetch associated Events
	fieldSelector := "involvedObject.name=" + name + ",involvedObject.kind=" + kind

	eventsNs := ns
	if scope == "Cluster" {
		// Cluster scoped resource events usually have an empty namespace in involvedObject.
		// Searching in "" (all namespaces) with fieldSelector will find them correctly.
		eventsNs = ""
	}

	eventList, err := clientset.CoreV1().Events(eventsNs).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector,
	})

	type EventSimple struct {
		Type     string `json:"type"`
		Reason   string `json:"reason"`
		Message  string `json:"message"`
		Count    int32  `json:"count"`
		LastSeen string `json:"last_seen"`
		Age      string `json:"age"`
	}

	var events []EventSimple
	if err == nil {
		for _, e := range eventList.Items {
			events = append(events, EventSimple{
				Type:     e.Type,
				Reason:   e.Reason,
				Message:  e.Message,
				Count:    e.Count,
				LastSeen: e.LastTimestamp.Time.Format(time.RFC3339),
				Age:      e.LastTimestamp.Time.Format(time.RFC3339),
			})
		}
		// Sort events by last seen desc
		sort.Slice(events, func(i, j int) bool {
			return events[i].LastSeen > events[j].LastSeen
		})
	} else {
		// Log error but don't fail entire request?
		// Just return empty events
	}

	// Ensure we don't pass configuration secrets if any?
	// User authorized for NS can read secrets in NS? Yes.
	// So dumping full YAML is fine for authorized user.

	c.JSON(200, gin.H{
		"manifest": string(yamlBytes),
		"events":   events,
		"raw":      resourceObj,
	})
}

// DeleteResource deletes a generic Kubernetes resource
func DeleteResource(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")
	name := c.Query("name")
	kind := c.Query("kind")

	if name == "" || kind == "" {
		c.JSON(400, gin.H{"error": "name and kind are required"})
		return
	}

	scope, ok := SupportedResources[kind]
	if !ok {
		c.JSON(400, gin.H{"error": "Unsupported kind: " + kind})
		return
	}

	if scope == "Namespace" && ns == "" && kind != "Namespace" {
		c.JSON(400, gin.H{"error": "namespace is required for " + kind})
		return
	}

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	var deleteErr error
	deleteOptions := metav1.DeleteOptions{}

	switch kind {
	case "Pod":
		deleteErr = clientset.CoreV1().Pods(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "Deployment":
		deleteErr = clientset.AppsV1().Deployments(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "ReplicaSet":
		deleteErr = clientset.AppsV1().ReplicaSets(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "StatefulSet":
		deleteErr = clientset.AppsV1().StatefulSets(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "DaemonSet":
		deleteErr = clientset.AppsV1().DaemonSets(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "Job":
		deleteErr = clientset.BatchV1().Jobs(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "CronJob":
		deleteErr = clientset.BatchV1().CronJobs(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "ReplicationController":
		deleteErr = clientset.CoreV1().ReplicationControllers(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "Service":
		deleteErr = clientset.CoreV1().Services(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "Endpoints":
		deleteErr = clientset.CoreV1().Endpoints(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "Ingress":
		deleteErr = clientset.NetworkingV1().Ingresses(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "NetworkPolicy":
		deleteErr = clientset.NetworkingV1().NetworkPolicies(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "IngressClass":
		deleteErr = clientset.NetworkingV1().IngressClasses().Delete(c.Request.Context(), name, deleteOptions)
	case "ConfigMap":
		deleteErr = clientset.CoreV1().ConfigMaps(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "Secret":
		deleteErr = clientset.CoreV1().Secrets(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "ResourceQuota":
		deleteErr = clientset.CoreV1().ResourceQuotas(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "LimitRange":
		deleteErr = clientset.CoreV1().LimitRanges(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "HorizontalPodAutoscaler":
		deleteErr = clientset.AutoscalingV2().HorizontalPodAutoscalers(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "PodDisruptionBudget":
		deleteErr = clientset.PolicyV1().PodDisruptionBudgets(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "Lease":
		deleteErr = clientset.CoordinationV1().Leases(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "PriorityClass":
		deleteErr = clientset.SchedulingV1().PriorityClasses().Delete(c.Request.Context(), name, deleteOptions)
	case "RuntimeClass":
		deleteErr = clientset.NodeV1().RuntimeClasses().Delete(c.Request.Context(), name, deleteOptions)
	case "MutatingWebhookConfiguration":
		deleteErr = clientset.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(c.Request.Context(), name, deleteOptions)
	case "ValidatingWebhookConfiguration":
		deleteErr = clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Delete(c.Request.Context(), name, deleteOptions)
	case "PersistentVolume":
		deleteErr = clientset.CoreV1().PersistentVolumes().Delete(c.Request.Context(), name, deleteOptions)
	case "PersistentVolumeClaim":
		deleteErr = clientset.CoreV1().PersistentVolumeClaims(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "StorageClass":
		deleteErr = clientset.StorageV1().StorageClasses().Delete(c.Request.Context(), name, deleteOptions)
	case "ServiceAccount":
		deleteErr = clientset.CoreV1().ServiceAccounts(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "ClusterRole":
		deleteErr = clientset.RbacV1().ClusterRoles().Delete(c.Request.Context(), name, deleteOptions)
	case "Role":
		deleteErr = clientset.RbacV1().Roles(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "ClusterRoleBinding":
		deleteErr = clientset.RbacV1().ClusterRoleBindings().Delete(c.Request.Context(), name, deleteOptions)
	case "RoleBinding":
		deleteErr = clientset.RbacV1().RoleBindings(ns).Delete(c.Request.Context(), name, deleteOptions)
	case "Node":
		deleteErr = clientset.CoreV1().Nodes().Delete(c.Request.Context(), name, deleteOptions)
	case "Namespace":
		deleteErr = clientset.CoreV1().Namespaces().Delete(c.Request.Context(), name, deleteOptions)
	case "Event":
		deleteErr = clientset.CoreV1().Events(ns).Delete(c.Request.Context(), name, deleteOptions)
	default:
		c.JSON(400, gin.H{"error": "Unsupported kind for deletion: " + kind})
		return
	}

	if deleteErr != nil {
		c.JSON(500, gin.H{"error": "Failed to delete resource: " + deleteErr.Error()})
		return
	}

	// Create audit log for deletion
	RecordAuditLog(c, "DELETE_KUBE_RESOURCE", gin.H{
		"context":   ctxName,
		"namespace": ns,
		"name":      name,
		"kind":      kind,
	})

	c.JSON(200, gin.H{"message": "Resource deleted successfully"})
}
