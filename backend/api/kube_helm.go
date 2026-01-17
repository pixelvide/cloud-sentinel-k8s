package api

import (
	"cloud-sentinel-k8s/pkg/models"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// ListHelmReleases lists all Helm releases across all namespaces
func ListHelmReleases(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	ctxName := GetKubeContext(c)
	ns := c.Query("namespace")

	// Get the shared client config which handles auth sanitization (e.g. glab)
	clientConfig, err := GetClientConfig(user.StorageNamespace, ctxName)
	if err != nil {
		log.Printf("[Helm] Error getting client config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get kubernetes config: " + err.Error()})
		return
	}
	log.Printf("[Helm] Listing releases. Context: %s, Namespace: %s, StorageNS: %s", ctxName, ns, user.StorageNamespace)

	settings := cli.New()

	// Helm's action configuration needs a specific namespace or empty for all.
	// If a namespace is selected (and not all namespaces), use it.
	targetNamespace := settings.Namespace()
	if ns == "__all__" {
		targetNamespace = ""
	} else if ns != "" {
		targetNamespace = ns
	}

	actionConfig := new(action.Configuration)

	// Create our custom getter that wraps the shared client config
	getter := &CustomRESTClientGetter{ClientConfig: clientConfig}

	if err := actionConfig.Init(getter, targetNamespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		log.Printf("[Helm] Error initializing action config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize Helm configuration: " + err.Error()})
		return
	}

	client := action.NewList(actionConfig)
	if ns == "" || ns == "__all__" {
		client.AllNamespaces = true
	} else {
		client.AllNamespaces = false
	}
	client.SetStateMask() // Default is to list deployed releases

	releases, err := client.Run()
	if err != nil {
		log.Printf("[Helm] Error running list: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list Helm releases: " + err.Error()})
		return
	}
	log.Printf("[Helm] Found %d releases", len(releases))

	// Transform to a simpler structure if needed
	type ReleaseSimple struct {
		Name         string `json:"name"`
		Namespace    string `json:"namespace"`
		Revision     int    `json:"revision"`
		Updated      string `json:"updated"`
		Status       string `json:"status"`
		ChartName    string `json:"chart_name"`
		ChartVersion string `json:"chart_version"`
		AppVersion   string `json:"app_version"`
	}

	var result []ReleaseSimple
	for _, r := range releases {
		result = append(result, ReleaseSimple{
			Name:         r.Name,
			Namespace:    r.Namespace,
			Revision:     r.Version,
			Updated:      r.Info.LastDeployed.String(),
			Status:       r.Info.Status.String(),
			ChartName:    r.Chart.Metadata.Name,
			ChartVersion: r.Chart.Metadata.Version,
			AppVersion:   r.Chart.Metadata.AppVersion,
		})
	}

	c.JSON(http.StatusOK, gin.H{"releases": result})
}

// CustomRESTClientGetter implements genericclioptions.RESTClientGetter
type CustomRESTClientGetter struct {
	ClientConfig clientcmd.ClientConfig
}

func (c *CustomRESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	return c.ClientConfig.ClientConfig()
}

func (c *CustomRESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	config, err := c.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	// The default discovery client is not cached, but for this short-lived request it's likely fine.
	// If performance is an issue, we should cache this too.
	// However, ToDiscoveryClient expects CachedDiscoveryInterface.
	// discovery.NewDiscoveryClientForConfig returns *DiscoveryClient which implements DiscoveryInterface but not CachedDiscoveryInterface.
	// We need to wrap it in a memory cache.
	// Since we don't have memory cache imported easily, let's see what standard tools do.
	// Usually: memory.NewMemCacheClient(discoveryClient)
	// But we'd need "k8s.io/client-go/discovery/cached/memory"

	// Actually, for now, let's try just returning the discovery client if it satisfies the interface,
	// checking the interface definition:
	// type CachedDiscoveryInterface interface { DiscoveryInterface; Fresh(); Invalidate() }
	// *DiscoveryClient does NOT implement this.

	// So we must use a cache wrapper.
	d, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	// We will need to import "k8s.io/client-go/discovery/cached/memory"
	return memory.NewMemCacheClient(d), nil
}

func (c *CustomRESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := c.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient, func(msg string) { log.Println(msg) })
	return expander, nil
}

func (c *CustomRESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return c.ClientConfig
}
