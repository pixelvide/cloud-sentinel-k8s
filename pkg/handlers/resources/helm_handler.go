package resources

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/cluster"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/common"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/helm"
	v3 "github.com/pixelvide/cloud-sentinel-k8s/pkg/helm/types/v3"
	"sigs.k8s.io/yaml"
)

type HelmHandler struct{}

func NewHelmHandler() *HelmHandler {
	return &HelmHandler{}
}

func (h *HelmHandler) List(c *gin.Context) {
	rawNamespace := c.Param("namespace")
	var namespaces []string
	if rawNamespace != "" && rawNamespace != "_all" {
		namespaces = strings.Split(rawNamespace, ",")
	}

	cs := c.MustGet("cluster").(*cluster.ClientSet)

	queryNamespace := ""
	if len(namespaces) == 1 {
		queryNamespace = namespaces[0]
	}

	releases, err := helm.ListReleases(cs.Configuration, queryNamespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var items []v3.HelmRelease
	for _, r := range releases {
		// Filter by namespace if multiple namespaces were requested
		if queryNamespace == "" && len(namespaces) > 1 {
			found := false
			for _, ns := range namespaces {
				if r.Namespace == ns {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		items = append(items, v3.HelmRelease{
			Name:       r.Name,
			Namespace:  r.Namespace,
			Revision:   r.Version,
			Status:     r.Info.Status.String(),
			Chart:      r.Chart.Metadata.Name + "-" + r.Chart.Metadata.Version,
			AppVersion: r.Chart.Metadata.AppVersion,
			Updated:    r.Info.LastDeployed.Time,
		})
	}

	// If items is nil, return empty array
	if items == nil {
		items = []v3.HelmRelease{}
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *HelmHandler) Get(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	cs := c.MustGet("cluster").(*cluster.ClientSet)
	release, err := helm.GetRelease(cs.Configuration, namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	valuesBytes, err := yaml.Marshal(release.Config)
	values := ""
	if err != nil {
		log.Printf("Error marshaling values for release %s: %v", name, err)
	} else {
		values = string(valuesBytes)
	}

	item := v3.HelmRelease{
		Name:       release.Name,
		Namespace:  release.Namespace,
		Revision:   release.Version,
		Status:     release.Info.Status.String(),
		Chart:      release.Chart.Metadata.Name + "-" + release.Chart.Metadata.Version,
		AppVersion: release.Chart.Metadata.AppVersion,
		Updated:    release.Info.LastDeployed.Time,
		Values:     values,
		Notes:      release.Info.Notes,
		Manifest:   release.Manifest,
	}

	c.JSON(http.StatusOK, item)
}

func (h *HelmHandler) Create(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *HelmHandler) Update(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *HelmHandler) Delete(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	cs := c.MustGet("cluster").(*cluster.ClientSet)
	if err := helm.UninstallRelease(cs.Configuration, namespace, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Release deleted"})
}

func (h *HelmHandler) Patch(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *HelmHandler) IsClusterScoped() bool {
	return false
}

func (h *HelmHandler) Searchable() bool {
	return true
}

func (h *HelmHandler) Search(c *gin.Context, query string, limit int64) ([]common.SearchResult, error) {
	// Basic search implementation can be added here if needed
	return []common.SearchResult{}, nil
}

func (h *HelmHandler) GetResource(c *gin.Context, namespace, name string) (interface{}, error) {
	cs := c.MustGet("cluster").(*cluster.ClientSet)
	return helm.GetRelease(cs.Configuration, namespace, name)
}

func (h *HelmHandler) registerCustomRoutes(group *gin.RouterGroup) {
	group.POST("/:namespace/:name/rollback", h.Rollback)
}

func (h *HelmHandler) Rollback(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var req struct {
		Revision int `json:"revision"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cs := c.MustGet("cluster").(*cluster.ClientSet)
	if err := helm.RollbackRelease(cs.Configuration, namespace, name, req.Revision); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rollback successful"})
}

func (h *HelmHandler) ListHistory(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	cs := c.MustGet("cluster").(*cluster.ClientSet)
	history, err := helm.GetReleaseHistory(cs.Configuration, namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var items []v3.HelmRelease
	for _, r := range history {
		items = append(items, v3.HelmRelease{
			Name:       r.Name,
			Namespace:  r.Namespace,
			Revision:   r.Version,
			Status:     r.Info.Status.String(),
			Chart:      r.Chart.Metadata.Name + "-" + r.Chart.Metadata.Version,
			AppVersion: r.Chart.Metadata.AppVersion,
			Updated:    r.Info.LastDeployed.Time,
			// History doesn't typically need values/manifest for the list view
		})
	}

	// Sort logic could be added here (Helm typically returns sorted by revision)
	// Reverse order (newest first)
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *HelmHandler) Describe(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *HelmHandler) GetAnalysis(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}
