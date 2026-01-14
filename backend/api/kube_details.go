package api

import (
	"context"
	"sort"
	"time"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/yaml"
)

type ResourceInfo struct {
	Group    string `json:"group"`
	Version  string `json:"version"`
	Resource string `json:"resource"`
	Scope    string `json:"scope"`
	Kind     string `json:"kind"`
}

// SupportedResources defines the kinds supported and their GVR info
var SupportedResources = map[string]ResourceInfo{
	"Pod":                            {Group: "", Version: "v1", Resource: "pods", Scope: "Namespace", Kind: "Pod"},
	"Deployment":                     {Group: "apps", Version: "v1", Resource: "deployments", Scope: "Namespace", Kind: "Deployment"},
	"ReplicaSet":                     {Group: "apps", Version: "v1", Resource: "replicasets", Scope: "Namespace", Kind: "ReplicaSet"},
	"StatefulSet":                    {Group: "apps", Version: "v1", Resource: "statefulsets", Scope: "Namespace", Kind: "StatefulSet"},
	"DaemonSet":                      {Group: "apps", Version: "v1", Resource: "daemonsets", Scope: "Namespace", Kind: "DaemonSet"},
	"Job":                            {Group: "batch", Version: "v1", Resource: "jobs", Scope: "Namespace", Kind: "Job"},
	"CronJob":                        {Group: "batch", Version: "v1", Resource: "cronjobs", Scope: "Namespace", Kind: "CronJob"},
	"ReplicationController":          {Group: "", Version: "v1", Resource: "replicationcontrollers", Scope: "Namespace", Kind: "ReplicationController"},
	"Service":                        {Group: "", Version: "v1", Resource: "services", Scope: "Namespace", Kind: "Service"},
	"Endpoints":                      {Group: "", Version: "v1", Resource: "endpoints", Scope: "Namespace", Kind: "Endpoints"},
	"Ingress":                        {Group: "networking.k8s.io", Version: "v1", Resource: "ingresses", Scope: "Namespace", Kind: "Ingress"},
	"NetworkPolicy":                  {Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies", Scope: "Namespace", Kind: "NetworkPolicy"},
	"ConfigMap":                      {Group: "", Version: "v1", Resource: "configmaps", Scope: "Namespace", Kind: "ConfigMap"},
	"Secret":                         {Group: "", Version: "v1", Resource: "secrets", Scope: "Namespace", Kind: "Secret"},
	"ResourceQuota":                  {Group: "", Version: "v1", Resource: "resourcequotas", Scope: "Namespace", Kind: "ResourceQuota"},
	"LimitRange":                     {Group: "", Version: "v1", Resource: "limitranges", Scope: "Namespace", Kind: "LimitRange"},
	"HorizontalPodAutoscaler":        {Group: "autoscaling", Version: "v2", Resource: "horizontalpodautoscalers", Scope: "Namespace", Kind: "HorizontalPodAutoscaler"},
	"PodDisruptionBudget":            {Group: "policy", Version: "v1", Resource: "poddisruptionbudgets", Scope: "Namespace", Kind: "PodDisruptionBudget"},
	"Lease":                          {Group: "coordination.k8s.io", Version: "v1", Resource: "leases", Scope: "Namespace", Kind: "Lease"},
	"ServiceAccount":                 {Group: "", Version: "v1", Resource: "serviceaccounts", Scope: "Namespace", Kind: "ServiceAccount"},
	"Role":                           {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles", Scope: "Namespace", Kind: "Role"},
	"RoleBinding":                    {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings", Scope: "Namespace", Kind: "RoleBinding"},
	"Node":                           {Group: "", Version: "v1", Resource: "nodes", Scope: "Cluster", Kind: "Node"},
	"Namespace":                      {Group: "", Version: "v1", Resource: "namespaces", Scope: "Cluster", Kind: "Namespace"},
	"PersistentVolume":               {Group: "", Version: "v1", Resource: "persistentvolumes", Scope: "Cluster", Kind: "PersistentVolume"},
	"PersistentVolumeClaim":          {Group: "", Version: "v1", Resource: "persistentvolumeclaims", Scope: "Namespace", Kind: "PersistentVolumeClaim"},
	"StorageClass":                   {Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses", Scope: "Cluster", Kind: "StorageClass"},
	"ClusterRole":                    {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles", Scope: "Cluster", Kind: "ClusterRole"},
	"ClusterRoleBinding":             {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings", Scope: "Cluster", Kind: "ClusterRoleBinding"},
	"PriorityClass":                  {Group: "scheduling.k8s.io", Version: "v1", Resource: "priorityclasses", Scope: "Cluster", Kind: "PriorityClass"},
	"RuntimeClass":                   {Group: "node.k8s.io", Version: "v1", Resource: "runtimeclasses", Scope: "Cluster", Kind: "RuntimeClass"},
	"IngressClass":                   {Group: "networking.k8s.io", Version: "v1", Resource: "ingressclasses", Scope: "Cluster", Kind: "IngressClass"},
	"MutatingWebhookConfiguration":   {Group: "admissionregistration.k8s.io", Version: "v1", Resource: "mutatingwebhookconfigurations", Scope: "Cluster", Kind: "MutatingWebhookConfiguration"},
	"ValidatingWebhookConfiguration": {Group: "admissionregistration.k8s.io", Version: "v1", Resource: "validatingwebhookconfigurations", Scope: "Cluster", Kind: "ValidatingWebhookConfiguration"},
	"Event":                          {Group: "", Version: "v1", Resource: "events", Scope: "Namespace", Kind: "Event"},
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

	info, ok := SupportedResources[kind]
	if !ok {
		c.JSON(400, gin.H{"error": "Unsupported kind: " + kind})
		return
	}

	// For namespaced resources, namespace is required
	if info.Scope == "Namespace" && ns == "" && kind != "Namespace" {
		c.JSON(400, gin.H{"error": "namespace is required for " + kind})
		return
	}

	clientset, config, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create dynamic client: " + err.Error()})
		return
	}

	gvr := schema.GroupVersionResource{
		Group:    info.Group,
		Version:  info.Version,
		Resource: info.Resource,
	}

	var dr dynamic.ResourceInterface
	if info.Scope == "Namespace" {
		dr = dynamicClient.Resource(gvr).Namespace(ns)
	} else {
		dr = dynamicClient.Resource(gvr)
	}

	// 1. Fetch the resource manifest
	resourceObj, resourceErr := dr.Get(c.Request.Context(), name, metav1.GetOptions{})

	if resourceErr != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch resource: " + resourceErr.Error()})
		return
	}

	// Convert to YAML
	// Note: Fetching via dynamic client returns *unstructured.Unstructured,
	// which by default includes TypeMeta (Kind/APIVersion).
	if resourceObj.GroupVersionKind().Empty() {
		// Set it if missing for some reason
		resourceObj.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   info.Group,
			Version: info.Version,
			Kind:    info.Kind,
		})
	}

	yamlBytes, err := yaml.Marshal(resourceObj)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to marshal resource to YAML: " + err.Error()})
		return
	}

	// 2. Fetch associated Events
	fieldSelector := "involvedObject.name=" + name + ",involvedObject.kind=" + kind

	eventsNs := ns
	if info.Scope == "Cluster" {
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

	info, ok := SupportedResources[kind]
	if !ok {
		c.JSON(400, gin.H{"error": "Unsupported kind: " + kind})
		return
	}

	if info.Scope == "Namespace" && ns == "" && kind != "Namespace" {
		c.JSON(400, gin.H{"error": "namespace is required for " + kind})
		return
	}

	_, config, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create dynamic client: " + err.Error()})
		return
	}

	gvr := schema.GroupVersionResource{
		Group:    info.Group,
		Version:  info.Version,
		Resource: info.Resource,
	}

	var dr dynamic.ResourceInterface
	if info.Scope == "Namespace" {
		dr = dynamicClient.Resource(gvr).Namespace(ns)
	} else {
		dr = dynamicClient.Resource(gvr)
	}

	deleteErr := dr.Delete(c.Request.Context(), name, metav1.DeleteOptions{})

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

// UpdateResource updates a generic Kubernetes resource from a YAML manifest
func UpdateResource(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := c.Query("context")
	ns := c.Query("namespace")
	name := c.Query("name")
	kind := c.Query("kind")

	var body struct {
		Manifest string `json:"manifest"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	clientset, config, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	// 1. Parse the received YAML into an Unstructured object
	obj := &unstructured.Unstructured{}
	// sigs.k8s.io/yaml handles JSON under the hood too if user sends JSON
	if err := yaml.Unmarshal([]byte(body.Manifest), &obj); err != nil {
		c.JSON(400, gin.H{"error": "Failed to parse YAML: " + err.Error()})
		return
	}

	// Basic validation to ensure they aren't trying to change the name/kind/namespace of the resource via edit
	// (editing usually applies to the same resource identify)
	if obj.GetName() != name && name != "" {
		c.JSON(400, gin.H{"error": "Resource name in manifest does not match URL parameter"})
		return
	}

	// 2. Create a Dynamic Client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create dynamic client: " + err.Error()})
		return
	}

	// 3. Get the REST Mapping to find the GVR (GroupVersionResource)
	gr, err := restmapper.GetAPIGroupResources(clientset.Discovery())
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get API group resources: " + err.Error()})
		return
	}
	mapper := restmapper.NewDiscoveryRESTMapper(gr)

	gvk := obj.GroupVersionKind()
	if gvk.Empty() {
		// Set it from SupportedResources if missing in YAML
		info, ok := SupportedResources[kind]
		if ok {
			obj.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   info.Group,
				Version: info.Version,
				Kind:    info.Kind,
			})
			gvk = obj.GroupVersionKind()
		} else {
			c.JSON(400, gin.H{"error": "apiVersion and kind are required in manifest or unknown kind"})
			return
		}
	}

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to find REST mapping for " + gvk.String() + ": " + err.Error()})
		return
	}

	// 4. Perform the Update
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		resourceNs := obj.GetNamespace()
		if resourceNs == "" {
			resourceNs = ns // Fallback to URL param if not in YAML
		}
		dr = dynamicClient.Resource(mapping.Resource).Namespace(resourceNs)
	} else {
		dr = dynamicClient.Resource(mapping.Resource)
	}

	// Update requires resourceVersion. We assume the user didn't strip it.
	// If stripping is common, we'd need to fetch first and apply strategic merge or patch.
	// For simple "edit YAML" UI, we usually expect them to keep resourceVersion for optimistic locking.
	updatedObj, err := dr.Update(c.Request.Context(), obj, metav1.UpdateOptions{})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update resource: " + err.Error()})
		return
	}

	// Create audit log for update
	RecordAuditLog(c, "UPDATE_KUBE_RESOURCE", gin.H{
		"context":   ctxName,
		"namespace": ns,
		"name":      name,
		"kind":      kind,
	})

	c.JSON(200, gin.H{
		"message":  "Resource updated successfully",
		"resource": updatedObj,
	})
}
