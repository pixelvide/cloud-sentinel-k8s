package api

import (
	"net/http"
	"time"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetConfigMaps lists config maps for a given namespace and context
func GetConfigMaps(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type ConfigMapInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Age       string `json:"age"`
		DataCount int    `json:"data_count"`
	}

	var configMaps []ConfigMapInfo

	for _, singleNs := range namespaces {
		list, err := clientset.CoreV1().ConfigMaps(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			configMaps = append(configMaps, ConfigMapInfo{
				Name:      item.Name,
				Namespace: item.Namespace,
				Age:       item.CreationTimestamp.Time.Format(time.RFC3339),
				DataCount: len(item.Data),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"configmaps": configMaps})
}

// GetSecrets lists secrets for a given namespace and context
func GetSecrets(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type SecretInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Age       string `json:"age"`
		Type      string `json:"type"`
		DataCount int    `json:"data_count"`
	}

	var secrets []SecretInfo

	for _, singleNs := range namespaces {
		list, err := clientset.CoreV1().Secrets(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			secrets = append(secrets, SecretInfo{
				Name:      item.Name,
				Namespace: item.Namespace,
				Age:       item.CreationTimestamp.Time.Format(time.RFC3339),
				Type:      string(item.Type),
				DataCount: len(item.Data),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"secrets": secrets})
}

// GetResourceQuotas lists resource quotas for a given namespace and context
func GetResourceQuotas(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type ResourceQuotaInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Age       string `json:"age"`
	}

	var resourceQuotas []ResourceQuotaInfo

	for _, singleNs := range namespaces {
		list, err := clientset.CoreV1().ResourceQuotas(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			resourceQuotas = append(resourceQuotas, ResourceQuotaInfo{
				Name:      item.Name,
				Namespace: item.Namespace,
				Age:       item.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"resourcequotas": resourceQuotas})
}

// GetLimitRanges lists limit ranges for a given namespace and context
func GetLimitRanges(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type LimitRangeInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Age       string `json:"age"`
	}

	var limitRanges []LimitRangeInfo

	for _, singleNs := range namespaces {
		list, err := clientset.CoreV1().LimitRanges(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			limitRanges = append(limitRanges, LimitRangeInfo{
				Name:      item.Name,
				Namespace: item.Namespace,
				Age:       item.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"limitranges": limitRanges})
}

// GetHPAs lists horizontal pod autoscalers for a given namespace and context
func GetHPAs(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type HPAInfo struct {
		Name         string `json:"name"`
		Namespace    string `json:"namespace"`
		Age          string `json:"age"`
		Reference    string `json:"reference"`
		MinReplicas  int32  `json:"min_replicas"`
		MaxReplicas  int32  `json:"max_replicas"`
		CurrReplicas int32  `json:"curr_replicas"`
	}

	var hpas []HPAInfo

	for _, singleNs := range namespaces {
		list, err := clientset.AutoscalingV2().HorizontalPodAutoscalers(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			minReplicas := int32(0)
			if item.Spec.MinReplicas != nil {
				minReplicas = *item.Spec.MinReplicas
			}
			hpas = append(hpas, HPAInfo{
				Name:         item.Name,
				Namespace:    item.Namespace,
				Age:          item.CreationTimestamp.Time.Format(time.RFC3339),
				Reference:    item.Spec.ScaleTargetRef.Kind + "/" + item.Spec.ScaleTargetRef.Name,
				MinReplicas:  minReplicas,
				MaxReplicas:  item.Spec.MaxReplicas,
				CurrReplicas: item.Status.CurrentReplicas,
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"hpas": hpas})
}

// GetPDBs lists pod disruption budgets for a given namespace and context
func GetPDBs(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type PDBInfo struct {
		Name       string `json:"name"`
		Namespace  string `json:"namespace"`
		Age        string `json:"age"`
		MinAvail   string `json:"min_available"`
		MaxUnavail string `json:"max_unavailable"`
	}

	var pdbs []PDBInfo

	for _, singleNs := range namespaces {
		list, err := clientset.PolicyV1().PodDisruptionBudgets(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			minAvail := ""
			if item.Spec.MinAvailable != nil {
				minAvail = item.Spec.MinAvailable.String()
			}
			maxUnavail := ""
			if item.Spec.MaxUnavailable != nil {
				maxUnavail = item.Spec.MaxUnavailable.String()
			}
			pdbs = append(pdbs, PDBInfo{
				Name:       item.Name,
				Namespace:  item.Namespace,
				Age:        item.CreationTimestamp.Time.Format(time.RFC3339),
				MinAvail:   minAvail,
				MaxUnavail: maxUnavail,
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"pdbs": pdbs})
}

// GetPriorityClasses lists priority classes for a given context (cluster-wide)
func GetPriorityClasses(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	list, err := clientset.SchedulingV1().PriorityClasses().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type PCInfo struct {
		Name  string `json:"name"`
		Value int32  `json:"value"`
		Age   string `json:"age"`
	}

	var pcs []PCInfo
	for _, item := range list.Items {
		pcs = append(pcs, PCInfo{
			Name:  item.Name,
			Value: item.Value,
			Age:   item.CreationTimestamp.Time.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"priorityclasses": pcs})
}

// GetRuntimeClasses lists runtime classes for a given context (cluster-wide)
func GetRuntimeClasses(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	list, err := clientset.NodeV1().RuntimeClasses().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type RCInfo struct {
		Name    string `json:"name"`
		Handler string `json:"handler"`
		Age     string `json:"age"`
	}

	var rcs []RCInfo
	for _, item := range list.Items {
		rcs = append(rcs, RCInfo{
			Name:    item.Name,
			Handler: item.Handler,
			Age:     item.CreationTimestamp.Time.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"runtimeclasses": rcs})
}

// GetLeases lists leases for a given namespace and context
func GetLeases(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ns := c.Query("namespace")
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	namespaces := ParseNamespaces(ns)

	type LeaseInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Age       string `json:"age"`
		Holder    string `json:"holder"`
	}

	var leases []LeaseInfo

	for _, singleNs := range namespaces {
		list, err := clientset.CoordinationV1().Leases(singleNs).List(c.Request.Context(), metav1.ListOptions{})
		if err != nil {
			if len(namespaces) == 1 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			continue
		}

		for _, item := range list.Items {
			holder := ""
			if item.Spec.HolderIdentity != nil {
				holder = *item.Spec.HolderIdentity
			}
			leases = append(leases, LeaseInfo{
				Name:      item.Name,
				Namespace: item.Namespace,
				Age:       item.CreationTimestamp.Time.Format(time.RFC3339),
				Holder:    holder,
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"leases": leases})
}

// GetMutatingWebhooks lists mutating webhook configurations for a given context (cluster-wide)
func GetMutatingWebhooks(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	list, err := clientset.AdmissionregistrationV1().MutatingWebhookConfigurations().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type WebhookInfo struct {
		Name     string `json:"name"`
		Age      string `json:"age"`
		Webhooks int    `json:"webhooks_count"`
	}

	var hooks []WebhookInfo
	for _, item := range list.Items {
		hooks = append(hooks, WebhookInfo{
			Name:     item.Name,
			Age:      item.CreationTimestamp.Time.Format(time.RFC3339),
			Webhooks: len(item.Webhooks),
		})
	}
	c.JSON(http.StatusOK, gin.H{"mutatingwebhooks": hooks})
}

// GetValidatingWebhooks lists validating webhook configurations for a given context (cluster-wide)
func GetValidatingWebhooks(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := c.Query("context")

	clientset, _, err := GetClientInfo(user.StorageNamespace, ctxName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config: " + err.Error()})
		return
	}

	list, err := clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type WebhookInfo struct {
		Name     string `json:"name"`
		Age      string `json:"age"`
		Webhooks int    `json:"webhooks_count"`
	}

	var hooks []WebhookInfo
	for _, item := range list.Items {
		hooks = append(hooks, WebhookInfo{
			Name:     item.Name,
			Age:      item.CreationTimestamp.Time.Format(time.RFC3339),
			Webhooks: len(item.Webhooks),
		})
	}
	c.JSON(http.StatusOK, gin.H{"validatingwebhooks": hooks})
}
