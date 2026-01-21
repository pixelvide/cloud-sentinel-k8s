package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/cluster"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/helm"
)

func ListHelmReleases(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "_all" {
		namespace = ""
	}

	cs := c.MustGet("cluster").(*cluster.ClientSet)

	releases, err := helm.ListReleases(cs.Configuration, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Simplify the response for the frontend
	type Release struct {
		Name       string `json:"name"`
		Namespace  string `json:"namespace"`
		Revision   int    `json:"revision"`
		Status     string `json:"status"`
		Chart      string `json:"chart"`
		AppVersion string `json:"app_version"`
		Updated    string `json:"updated"`
	}

	var response []Release
	for _, r := range releases {
		response = append(response, Release{
			Name:       r.Name,
			Namespace:  r.Namespace,
			Revision:   r.Version,
			Status:     r.Info.Status.String(),
			Chart:      r.Chart.Metadata.Name + "-" + r.Chart.Metadata.Version,
			AppVersion: r.Chart.Metadata.AppVersion,
			Updated:    r.Info.LastDeployed.String(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"items": response})
}
