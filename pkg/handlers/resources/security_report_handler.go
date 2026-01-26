package resources

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/cluster"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/model"
	appsv1 "k8s.io/api/apps/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecurityReportHandler struct{}

func NewSecurityReportHandler() *SecurityReportHandler {
	return &SecurityReportHandler{}
}

var vulnerabilityReportKind = schema.GroupVersionKind{
	Group:   "aquasecurity.github.io",
	Version: "v1alpha1",
	Kind:    "VulnerabilityReport",
}

var clusterVulnerabilityReportKind = schema.GroupVersionKind{
	Group:   "aquasecurity.github.io",
	Version: "v1alpha1",
	Kind:    "ClusterVulnerabilityReport",
}

// CheckStatus checks if the Trivy Operator is installed by looking for the CRD
func (h *SecurityReportHandler) CheckStatus(c *gin.Context) {
	cs := c.MustGet("cluster").(*cluster.ClientSet)

	// Check if the CRD exists
	var crd apiextensionsv1.CustomResourceDefinition
	err := cs.K8sClient.Get(c.Request.Context(), client.ObjectKey{Name: "vulnerabilityreports.aquasecurity.github.io"}, &crd)

	installed := err == nil
	c.JSON(http.StatusOK, model.SecurityStatusResponse{TrivyInstalled: installed})
}

// ListReports fetches vulnerability reports, optionally filtered by workload
func (h *SecurityReportHandler) ListReports(c *gin.Context) {
	cs := c.MustGet("cluster").(*cluster.ClientSet)
	namespace := c.Query("namespace")
	workloadKind := c.Query("workloadKind") // e.g. Pod, Deployment
	workloadName := c.Query("workloadName")

	if namespace == "" && workloadKind != "Node" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is required for namespaced resources"})
		return
	}

	// 1. Check if CRD exists first to avoid confusing errors
	var crd apiextensionsv1.CustomResourceDefinition
	if err := cs.K8sClient.Get(c.Request.Context(), client.ObjectKey{Name: "vulnerabilityreports.aquasecurity.github.io"}, &crd); err != nil {
		// If CRD not found, try ClusterVulnerabilityReport just in case, or return empty
		c.JSON(http.StatusOK, model.VulnerabilityReportList{Items: []model.VulnerabilityReport{}})
		return
	}

	// 2. List Reports
	var list unstructured.UnstructuredList
	opts := []client.ListOption{}

	if workloadKind == "Node" {
		list.SetGroupVersionKind(clusterVulnerabilityReportKind)
		// For ClusterVulnerabilityReport, no namespace
	} else {
		list.SetGroupVersionKind(vulnerabilityReportKind)
		opts = append(opts, client.InNamespace(namespace))
	}

	// Trivy Operator labels reports with the workload details
	// labels: trivy-operator.resource.kind, trivy-operator.resource.name

	// Special handling for Deployment: Trivy attaches reports to the ReplicaSet
	switch {
	case workloadKind == "Deployment":
		var rsList appsv1.ReplicaSetList
		if err := cs.K8sClient.List(c.Request.Context(), &rsList, client.InNamespace(namespace)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to list recyclasets: %v", err)})
			return
		}

		var targetRSNames []string
		for _, rs := range rsList.Items {
			for _, owner := range rs.OwnerReferences {
				if owner.Kind == "Deployment" && owner.Name == workloadName {
					// Found a ReplicaSet owned by this Deployment
					targetRSNames = append(targetRSNames, rs.Name)
					break
				}
			}
		}

		if len(targetRSNames) == 0 {
			// No RS found (or no RS owned by this deployment yet), return empty
			c.JSON(http.StatusOK, model.VulnerabilityReportList{Items: []model.VulnerabilityReport{}})
			return
		}

		// List ALL reports for ReplicaSets in this namespace, then filter in memory
		// This is efficient enough for typical namespaces
		labels := client.MatchingLabels{
			"trivy-operator.resource.kind": "ReplicaSet",
		}
		opts = append(opts, labels)

		// We can't easily set label selector for multiple names ("OR"), so we filter after list
		if err := cs.K8sClient.List(c.Request.Context(), &list, opts...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to list vulnerability reports: %v", err)})
			return
		}

		// Filter list to keep only those belonging to our target RSs
		filteredItems := []unstructured.Unstructured{}
		for _, item := range list.Items {
			lbls := item.GetLabels()
			reportResourceName := lbls["trivy-operator.resource.name"]
			for _, target := range targetRSNames {
				if reportResourceName == target {
					filteredItems = append(filteredItems, item)
					break
				}
			}
		}
		list.Items = filteredItems

	case workloadKind != "" && workloadName != "":
		labels := client.MatchingLabels{
			"trivy-operator.resource.kind": workloadKind,
			"trivy-operator.resource.name": workloadName,
		}
		opts = append(opts, labels)

		if err := cs.K8sClient.List(c.Request.Context(), &list, opts...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to list vulnerability reports: %v", err)})
			return
		}

	default:
		// No specific workload filter? Just list with existing opts (namespace only)
		if err := cs.K8sClient.List(c.Request.Context(), &list, opts...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to list vulnerability reports: %v", err)})
			return
		}
	}

	// Skip the original List call since we handled it inside the branches
	// Proceed to conversion

	// 3. Convert to typed Helper models
	reports := make([]model.VulnerabilityReport, 0, len(list.Items))
	for _, u := range list.Items {
		var report model.VulnerabilityReport
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
			continue // skip malformed
		}
		reports = append(reports, report)
	}

	c.JSON(http.StatusOK, model.VulnerabilityReportList{Items: reports})
}

// GetClusterSummary aggregates vulnerabilities across the entire cluster (or filtered namespace)
func (h *SecurityReportHandler) GetClusterSummary(c *gin.Context) {
	cs := c.MustGet("cluster").(*cluster.ClientSet)

	// Check CRD
	var crd apiextensionsv1.CustomResourceDefinition
	if err := cs.K8sClient.Get(c.Request.Context(), client.ObjectKey{Name: "vulnerabilityreports.aquasecurity.github.io"}, &crd); err != nil {
		c.JSON(http.StatusOK, model.ClusterSecuritySummary{})
		return
	}

	var list unstructured.UnstructuredList
	list.SetGroupVersionKind(vulnerabilityReportKind)

	if err := cs.K8sClient.List(c.Request.Context(), &list); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	summary := model.ClusterSecuritySummary{
		ScannedImages: len(list.Items),
	}

	for _, u := range list.Items {
		var report model.VulnerabilityReport
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
			continue
		}

		s := report.Report.Summary
		summary.TotalVulnerabilities.CriticalCount += s.CriticalCount
		summary.TotalVulnerabilities.HighCount += s.HighCount
		summary.TotalVulnerabilities.MediumCount += s.MediumCount
		summary.TotalVulnerabilities.LowCount += s.LowCount
		summary.TotalVulnerabilities.UnknownCount += s.UnknownCount

		if s.CriticalCount > 0 || s.HighCount > 0 || s.MediumCount > 0 || s.LowCount > 0 {
			summary.VulnerableImages++
		}
	}

	// 2. Aggregate by Workload for Top List
	workloadMap := make(map[string]*model.WorkloadSummary)

	for _, u := range list.Items {
		lbls := u.GetLabels()
		kind := lbls["trivy-operator.resource.kind"]
		name := lbls["trivy-operator.resource.name"]
		namespace := u.GetNamespace()

		if kind == "" || name == "" {
			continue
		}

		key := fmt.Sprintf("%s/%s/%s", namespace, kind, name)
		if _, exists := workloadMap[key]; !exists {
			workloadMap[key] = &model.WorkloadSummary{
				Namespace: namespace,
				Kind:      kind,
				Name:      name,
			}
		}

		// Extract summary from unstructured again (or reuse if I had stored it, but parsing twice is fine for now/low scale)
		var report model.VulnerabilityReport
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &report); err != nil {
			continue
		}
		s := report.Report.Summary

		w := workloadMap[key]
		w.Vulnerabilities.CriticalCount += s.CriticalCount
		w.Vulnerabilities.HighCount += s.HighCount
		w.Vulnerabilities.MediumCount += s.MediumCount
		w.Vulnerabilities.LowCount += s.LowCount
		w.Vulnerabilities.UnknownCount += s.UnknownCount
	}

	// Converts map to slice
	var workloads []model.WorkloadSummary
	for _, w := range workloadMap {
		workloads = append(workloads, *w)
	}

	// Sort
	sort.Slice(workloads, func(i, j int) bool {
		// Critical > High > Medium > Low
		if workloads[i].Vulnerabilities.CriticalCount != workloads[j].Vulnerabilities.CriticalCount {
			return workloads[i].Vulnerabilities.CriticalCount > workloads[j].Vulnerabilities.CriticalCount
		}
		if workloads[i].Vulnerabilities.HighCount != workloads[j].Vulnerabilities.HighCount {
			return workloads[i].Vulnerabilities.HighCount > workloads[j].Vulnerabilities.HighCount
		}
		if workloads[i].Vulnerabilities.MediumCount != workloads[j].Vulnerabilities.MediumCount {
			return workloads[i].Vulnerabilities.MediumCount > workloads[j].Vulnerabilities.MediumCount
		}
		return workloads[i].Vulnerabilities.LowCount > workloads[j].Vulnerabilities.LowCount
	})

	// Take top 10
	if len(workloads) > 10 {
		summary.TopVulnerableWorkloads = workloads[:10]
	} else {
		summary.TopVulnerableWorkloads = workloads
	}

	c.JSON(http.StatusOK, summary)
}

func (h *SecurityReportHandler) RegisterRoutes(group *gin.RouterGroup) {
	securityParams := group.Group("/security")
	securityParams.GET("/status", h.CheckStatus)
	securityParams.GET("/reports", h.ListReports)
	securityParams.GET("/summary", h.GetClusterSummary)
}
