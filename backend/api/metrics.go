package api

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"

	"cloud-sentinel-k8s/metrics"
	"cloud-sentinel-k8s/pkg/models"
)

// GetPodMetrics returns CPU/Memory usage history for a pod.
// It tries to use Prometheus (future) or falls back to metrics-server cache.
func GetPodMetrics(c *gin.Context) {
	// 1. Resolve context/client
	user := c.MustGet("user").(*models.User)
	storageNamespace := user.StorageNamespace

	// Get context using helper (checks header first)
	kubeContext := GetKubeContext(c)

	// 2. Initialize Metrics Client wrapper
	k8sClient, _, err := GetClientInfo(storageNamespace, kubeContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get k8s client: " + err.Error()})
		return
	}

	metricsClient, err := GetMetricsClient(storageNamespace, kubeContext)
	if err != nil {
		// If metrics client fails (e.g. metrics-server not installed), we still might want to try?
		// But our wrapper needs it.
		// For now, let's log and error.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metrics client: " + err.Error()})
		return
	}
	if metricsClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Metrics server not available"})
		return
	}

	// In a real optimized app, this 'mc' should be cached per cluster/user session
	// to maintain the sliding window cache!
	//
	// PROBLEM: NewClient creates a new cache every request.
	// We need a global/persistent cache for the fallback to work (sliding window).
	//
	// Solution: We need a singleton or long-lived instance of metrics.Client per cluster.
	//
	// Let's implement a global cache manager for metrics clients in this file.
	mc := GetGlobalMetricsClient(storageNamespace, kubeContext, k8sClient, metricsClient)

	// 3. Get Parameters
	namespace := c.Query("namespace")
	podName := c.Query("podName")
	if namespace == "" || podName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace and podName query params are required"})
		return
	}

	// 4. Fetch Metrics
	podMetrics, err := mc.GetPodMetrics(c.Request.Context(), namespace, podName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch metrics: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, podMetrics)
}

// --- Global Metrics Client Cache ---

var (
	metricsClients = make(map[string]*metrics.Client)
	metricsMutex   sync.RWMutex
)

func GetGlobalMetricsClient(storageNamespace, contextName string, k8sClient kubernetes.Interface, metricsClient metricsclientset.Interface) *metrics.Client {
	key := storageNamespace + "::" + contextName

	metricsMutex.RLock()
	if c, ok := metricsClients[key]; ok {
		metricsMutex.RUnlock()
		return c
	}
	metricsMutex.RUnlock()

	metricsMutex.Lock()
	defer metricsMutex.Unlock()

	// Double check
	if c, ok := metricsClients[key]; ok {
		return c
	}

	c := metrics.NewClient(k8sClient, metricsClient)
	metricsClients[key] = c
	return c
}
