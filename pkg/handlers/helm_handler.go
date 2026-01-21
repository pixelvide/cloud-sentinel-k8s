package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/cluster"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/helm"
)

func ListHelmReleases(c *gin.Context) {
	rawNamespace := c.Param("namespace")
	var namespaces []string
	if rawNamespace != "" && rawNamespace != "_all" {
		namespaces = strings.Split(rawNamespace, ",")
	}

	cs := c.MustGet("cluster").(*cluster.ClientSet)

	// Determine the namespace to query Helm with.
	// If single namespace, query efficiently.
	// If multiple or all, query all and filter.
	queryNamespace := ""
	if len(namespaces) == 1 {
		queryNamespace = namespaces[0]
	}

	releases, err := helm.ListReleases(cs.Configuration, queryNamespace)
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

func DeleteHelmRelease(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace and name are required"})
		return
	}

	cs := c.MustGet("cluster").(*cluster.ClientSet)

	if err := helm.UninstallRelease(cs.Configuration, namespace, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Release deleted"})
}
