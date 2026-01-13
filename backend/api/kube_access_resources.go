package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetServiceAccounts lists service accounts for a given namespace and context
func GetServiceAccounts(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type SAInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Secrets   int    `json:"secrets"`
		Age       string `json:"age"`
	}

	var sas []SAInfo

	for _, singleNs := range namespaces {
		list, err := clientset.CoreV1().ServiceAccounts(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			sas = append(sas, SAInfo{
				Name:      item.Name,
				Namespace: item.Namespace,
				Secrets:   len(item.Secrets),
				Age:       item.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"serviceaccounts": sas})
}

// GetClusterRoles lists cluster-wide cluster roles
func GetClusterRoles(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	list, err := clientset.RbacV1().ClusterRoles().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type RoleInfo struct {
		Name string `json:"name"`
		Age  string `json:"age"`
	}

	var roles []RoleInfo
	for _, item := range list.Items {
		roles = append(roles, RoleInfo{
			Name: item.Name,
			Age:  item.CreationTimestamp.Time.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"clusterroles": roles})
}

// GetRoles lists namespace-scoped roles
func GetRoles(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type RoleInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Age       string `json:"age"`
	}

	var roles []RoleInfo

	for _, singleNs := range namespaces {
		list, err := clientset.RbacV1().Roles(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			roles = append(roles, RoleInfo{
				Name:      item.Name,
				Namespace: item.Namespace,
				Age:       item.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// GetClusterRoleBindings lists cluster-wide cluster role bindings
func GetClusterRoleBindings(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	list, err := clientset.RbacV1().ClusterRoleBindings().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type BindingInfo struct {
		Name string `json:"name"`
		Role string `json:"role"`
		Age  string `json:"age"`
	}

	var bindings []BindingInfo
	for _, item := range list.Items {
		bindings = append(bindings, BindingInfo{
			Name: item.Name,
			Role: item.RoleRef.Name,
			Age:  item.CreationTimestamp.Time.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"clusterrolebindings": bindings})
}

// GetRoleBindings lists namespace-scoped role bindings
func GetRoleBindings(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type BindingInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Role      string `json:"role"`
		Age       string `json:"age"`
	}

	var bindings []BindingInfo

	for _, singleNs := range namespaces {
		list, err := clientset.RbacV1().RoleBindings(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			bindings = append(bindings, BindingInfo{
				Name:      item.Name,
				Namespace: item.Namespace,
				Role:      item.RoleRef.Name,
				Age:       item.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"rolebindings": bindings})
}
