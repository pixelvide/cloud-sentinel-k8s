package api

import (
	"context"
	"log"
	"sort"
	"time"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

// CRDInfo simplifies the CRD structure for the frontend
type CRDInfo struct {
	Name       string   `json:"name"`
	Group      string   `json:"group"`
	Version    string   `json:"version"`
	Scope      string   `json:"scope"`
	Kind       string   `json:"kind"`
	Resource   string   `json:"resource"` // Plural
	ShortNames []string `json:"short_names"`
	Categories []string `json:"categories"`
}

// GetCustomResourceDefinitions lists all CRDs in the cluster
func GetCustomResourceDefinitions(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := GetKubeContext(c)

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

	gvr := schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}
	list, err := dynamicClient.Resource(gvr).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to list CRDs: " + err.Error()})
		return
	}

	var crds []CRDInfo
	for _, item := range list.Items {
		spec, _, _ := unstructured.NestedMap(item.Object, "spec")
		group, _, _ := unstructured.NestedString(spec, "group")
		names, _, _ := unstructured.NestedMap(spec, "names")
		kind, _, _ := unstructured.NestedString(names, "kind")
		plural, _, _ := unstructured.NestedString(names, "plural")
		scope, _, _ := unstructured.NestedString(spec, "scope")
		shortNames, _, _ := unstructured.NestedStringSlice(names, "shortNames")
		categories, _, _ := unstructured.NestedStringSlice(names, "categories")

		versions, _, _ := unstructured.NestedSlice(spec, "versions")
		var version string
		for _, v := range versions {
			vMap, ok := v.(map[string]interface{})
			if !ok {
				continue
			}
			if stored, ok := vMap["storage"].(bool); ok && stored {
				version, _, _ = unstructured.NestedString(vMap, "name")
				break
			}
		}
		if version == "" && len(versions) > 0 {
			if vMap, ok := versions[0].(map[string]interface{}); ok {
				version, _, _ = unstructured.NestedString(vMap, "name")
			}
		}

		crds = append(crds, CRDInfo{
			Name:       item.GetName(),
			Group:      group,
			Version:    version,
			Scope:      scope,
			Kind:       kind,
			Resource:   plural,
			ShortNames: shortNames,
			Categories: categories,
		})
	}

	sort.Slice(crds, func(i, j int) bool {
		return crds[i].Name < crds[j].Name
	})

	c.JSON(200, gin.H{"items": crds})
}

// GetCustomResources lists resources for a specific CRD
func GetCustomResources(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := GetKubeContext(c)
	ns := c.Query("namespace")
	crdName := c.Param("crd_name")

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

	// Fetch CRD to discover GVR
	crdGVR := schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}
	crdObj, err := dynamicClient.Resource(crdGVR).Get(c.Request.Context(), crdName, metav1.GetOptions{})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch CRD: " + err.Error()})
		return
	}

	spec, _, _ := unstructured.NestedMap(crdObj.Object, "spec")
	group, _, _ := unstructured.NestedString(spec, "group")
	names, _, _ := unstructured.NestedMap(spec, "names")
	plural, _, _ := unstructured.NestedString(names, "plural")
	kind, _, _ := unstructured.NestedString(names, "kind")
	scope, _, _ := unstructured.NestedString(spec, "scope")

	versions, _, _ := unstructured.NestedSlice(spec, "versions")
	var version string
	for _, v := range versions {
		vMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if stored, ok := vMap["storage"].(bool); ok && stored {
			version, _, _ = unstructured.NestedString(vMap, "name")
			break
		}
	}

	gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: plural}
	var list *unstructured.UnstructuredList

	if ns == "__all__" {
		ns = ""
	}

	if scope == "Namespaced" && ns != "" {
		list, err = dynamicClient.Resource(gvr).Namespace(ns).List(c.Request.Context(), metav1.ListOptions{})
	} else {
		// If scope is Cluster, ns is ignored. If scope is Namespaced and ns is empty, list all namespaces
		list, err = dynamicClient.Resource(gvr).List(c.Request.Context(), metav1.ListOptions{})
	}

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to list resources: " + err.Error()})
		return
	}

	type ResourceSimple struct {
		Name      string            `json:"name"`
		Namespace string            `json:"namespace"`
		Created   string            `json:"created"`
		Age       string            `json:"age"`
		Labels    map[string]string `json:"labels"`
	}

	var resources []ResourceSimple
	for _, item := range list.Items {
		resources = append(resources, ResourceSimple{
			Name:      item.GetName(),
			Namespace: item.GetNamespace(),
			Created:   item.GetCreationTimestamp().Time.Format(time.RFC3339),
			Age:       item.GetCreationTimestamp().Time.Format(time.RFC3339),
			Labels:    item.GetLabels(),
		})
	}

	// Sort by name
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})

	c.JSON(200, gin.H{
		"items": resources,
		"crd": gin.H{
			"scope":    scope,
			"group":    group,
			"version":  version,
			"kind":     kind,
			"resource": plural,
		},
	})
}

// GetCustomResourceDetails fetches detailed info for a single CR
func GetCustomResourceDetails(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := GetKubeContext(c)
	ns := c.Query("namespace")
	crdName := c.Param("crd_name")
	name := c.Param("name")

	clientset, config, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	log.Printf("[DEBUG] Fetching CRD details. CRD: %s, Name: %s, Namespace: %s, Context: %s", crdName, name, ns, ctxName)

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create dynamic client: " + err.Error()})
		return
	}

	// Fetch CRD info
	crdGVR := schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}
	crdObj, err := dynamicClient.Resource(crdGVR).Get(c.Request.Context(), crdName, metav1.GetOptions{})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch CRD: " + err.Error()})
		return
	}

	spec, _, _ := unstructured.NestedMap(crdObj.Object, "spec")
	group, _, _ := unstructured.NestedString(spec, "group")
	names, _, _ := unstructured.NestedMap(spec, "names")
	plural, _, _ := unstructured.NestedString(names, "plural")
	kind, _, _ := unstructured.NestedString(names, "kind")
	scope, _, _ := unstructured.NestedString(spec, "scope")

	versions, _, _ := unstructured.NestedSlice(spec, "versions")
	var version string
	for _, v := range versions {
		vMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if stored, ok := vMap["storage"].(bool); ok && stored {
			version, _, _ = unstructured.NestedString(vMap, "name")
			break
		}
	}

	gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: plural}

	log.Printf("[DEBUG] Resolved GVR: %+v, Scope: %s", gvr, scope)

	var dr dynamic.ResourceInterface
	if scope == "Namespaced" {
		dr = dynamicClient.Resource(gvr).Namespace(ns)
	} else {
		dr = dynamicClient.Resource(gvr)
	}

	resourceObj, err := dr.Get(c.Request.Context(), name, metav1.GetOptions{})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch resource: " + err.Error()})
		return
	}

	// YAML
	yamlBytes, err := yaml.Marshal(resourceObj)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to marshal to YAML: " + err.Error()})
		return
	}

	// Events
	fieldSelector := "involvedObject.name=" + name + ",involvedObject.kind=" + kind
	eventsNs := ns
	if scope == "Cluster" {
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
		sort.Slice(events, func(i, j int) bool {
			return events[i].LastSeen > events[j].LastSeen
		})
	}

	c.JSON(200, gin.H{
		"manifest": string(yamlBytes),
		"events":   events,
		"raw":      resourceObj,
	})
}
