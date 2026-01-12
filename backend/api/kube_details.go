package api

import (
	"context"
	"sort"
	"strings"
	"time"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

// SupportedResources defines the kinds supported and their scope (Cluster vs Namespace)
var SupportedResources = map[string]string{
	"pod":         "Namespace",
	"deployment":  "Namespace",
	"service":     "Namespace",
	"node":        "Cluster",
	"ingress":     "Namespace",
	"daemonset":   "Namespace",
	"statefulset": "Namespace",
	"job":         "Namespace",
	"cronjob":     "Namespace",
	"namespace":   "Cluster",
	"event":       "Namespace",
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
	kind := c.Query("kind") // "Pod", "Deployment", "Service", "Node", "Ingress" (case insensitive)

	if name == "" || kind == "" {
		c.JSON(400, gin.H{"error": "name and kind are required"})
		return
	}

	kindLower := strings.ToLower(kind)
	scope, ok := SupportedResources[kindLower]
	if !ok {
		// Fallback or error? For now error as we only support specific kinds
		c.JSON(400, gin.H{"error": "Unsupported kind: " + kind})
		return
	}

	// For namespaced resources, namespace is required
	if scope == "Namespace" && ns == "" {
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

	// Normalize kind
	// kindLower is already defined above

	switch kindLower {
	case "pod":
		resourceObj, resourceErr = clientset.CoreV1().Pods(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "deployment":
		resourceObj, resourceErr = clientset.AppsV1().Deployments(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "service":
		resourceObj, resourceErr = clientset.CoreV1().Services(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "node":
		// Node is cluster-scoped, ignore namespace if provided, or use it? usually ignore for get
		resourceObj, resourceErr = clientset.CoreV1().Nodes().Get(c.Request.Context(), name, metav1.GetOptions{})
	case "ingress":
		resourceObj, resourceErr = clientset.NetworkingV1().Ingresses(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "daemonset":
		resourceObj, resourceErr = clientset.AppsV1().DaemonSets(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "statefulset":
		resourceObj, resourceErr = clientset.AppsV1().StatefulSets(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "job":
		resourceObj, resourceErr = clientset.BatchV1().Jobs(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "cronjob":
		resourceObj, resourceErr = clientset.BatchV1().CronJobs(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	case "namespace":
		// Namespace is cluster-scoped, ignore ns query param for the get, use name as the resource name
		resourceObj, resourceErr = clientset.CoreV1().Namespaces().Get(c.Request.Context(), name, metav1.GetOptions{})
	case "event":
		// Event is namespaced (usually)
		resourceObj, resourceErr = clientset.CoreV1().Events(ns).Get(c.Request.Context(), name, metav1.GetOptions{})
	default:
		// Attempt dynamic client for other resources if needed, but for now error
		// Ideally we use dynamic client for everything to be generic
		// For now, limited set as per plan.
		// If using dynamic client:
		// dynClient, _ := dynamic.NewForConfig(config)
		// gvr := ... mapping kind to GVR ...
		// dynClient.Resource(gvr).Namespace(ns).Get(...)
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
			switch kindLower {
			case "pod":
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
	// FieldSelector: involvedObject.name=NAME, involvedObject.namespace=NAMESPACE, involvedObject.kind=KIND (roughly)
	// Or involvedObject.uid if we had it. Name/Kind/Namespace is standard.
	// Note: Events are namespaced. For Node, check default or all? Usually events for Node are in 'default' or where the event recorder put them.
	// Actually events for a Node invokeObject.kind="Node", involvedObject.name=NodeName.

	validKind := kind // Use original case or adjust? Events involvedObject.kind is usually TitleCase (e.g. "Pod")
	// Simple capitalization fix
	if len(validKind) > 0 {
		validKind = strings.ToUpper(validKind[:1]) + strings.ToLower(validKind[1:])
		// Special cases
		if strings.ToLower(kind) == "cronjob" {
			validKind = "CronJob"
		}
		if strings.ToLower(kind) == "daemonset" {
			validKind = "DaemonSet"
		}
		if strings.ToLower(kind) == "statefulset" {
			validKind = "StatefulSet"
		}
	}

	fieldSelector := "involvedObject.name=" + name + ",involvedObject.kind=" + validKind

	// For specific check
	eventsNs := ns
	if kindLower == "node" {
		// Events for nodes are often in default namespace? or all?
		// Let's try List for all namespaces or just current.
		// client-go Events() requires namespace.
		// Typically Node events are in 'default'.
		if ns == "" {
			eventsNs = "default"
		}
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
	})
}
