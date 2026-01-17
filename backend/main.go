package main

import (
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog/v2"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	"cloud-sentinel-k8s/api"
	"cloud-sentinel-k8s/auth"
	"cloud-sentinel-k8s/pkg/middleware"
	"cloud-sentinel-k8s/pkg/models"
	"cloud-sentinel-k8s/pkg/utils"
)

var Version = "dev"

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	if klog.V(1).Enabled() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize services
	models.InitDB()
	auth.InitOIDC()
	auth.InitJWT()

	// Reconfigure all GitLab K8s agents on startup (for non-persistent data directories)
	go api.ReconfigureAllAgentsOnStartup()
	go api.ReconfigureAllKubeConfigsOnStartup()

	// Set app version from build tag
	api.SetAppVersion(Version)

	// Subpath support
	base := os.Getenv("CLOUD_SENTINEL_K8S_BASE")
	if base != "" && !strings.HasPrefix(base, "/") {
		base = "/" + base
	}
	base = strings.TrimSuffix(base, "/")

	r := gin.New()
	r.Use(middleware.Metrics())
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	setupAPIRouter(r, base)
	setupStatic(r, base)

	klog.Infof("Server running on :8080")
	r.Run(":8080")
}

func setupStatic(r *gin.Engine, base string) {
	if base != "" && base != "/" {
		r.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusFound, base+"/")
		})
	}

	// Serve static files from the "static" directory
	r.Static(base+"/assets", "./static/assets")

	// SPA Handler: Serve static files if they exist, otherwise serve index.html
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, base+"/api") {
			c.JSON(404, gin.H{"error": "API route not found"})
			return
		}

		if base != "" && strings.HasPrefix(path, base) {
			path = strings.TrimPrefix(path, base)
		}

		// Check if file exists in static folder (e.g. /sw.js, /favicon.ico)
		// filepath.Join handles path sanitization
		filePath := filepath.Join("./static", filepath.Clean(path))
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			c.File(filePath)
			return
		}

		// Fallback to SPA index.html for all other non-API routes
		content, err := os.ReadFile("./static/index.html")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read index.html"})
			return
		}

		htmlContent := utils.InjectBase(string(content), base)
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, htmlContent)
	})
}

func setupAPIRouter(r *gin.Engine, base string) {
	g := r.Group(base)

	g.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	g.GET("/metrics", gin.WrapH(promhttp.HandlerFor(prometheus.Gatherers{
		prometheus.DefaultGatherer,
		ctrlmetrics.Registry,
	}, promhttp.HandlerOpts{})))

	authGroup := g.Group("/api/v1/auth")
	{
		authGroup.GET("/login", auth.LoginHandler)
		authGroup.GET("/callback", auth.CallbackHandler)
		authGroup.GET("/logout", auth.LogoutHandler)
	}

	apiGroup := g.Group("/api/v1")
	apiGroup.Use(auth.AuthMiddleware())
	{
		apiGroup.GET("/me", api.GetMe)
		settings := apiGroup.Group("/settings")
		{
			gitlab := settings.Group("/gitlab")
			{
				gitlab.GET("", api.ListGitlabConfigs)
				gitlab.POST("", api.CreateGitlabConfig)
				gitlab.POST("/:id/validate", api.ValidateGitlabConfig)

				agents := gitlab.Group("/agents")
				{
					agents.GET("", api.ListGitlabAgentConfigs)
					agents.POST("", api.CreateGitlabAgentConfig)
					agents.DELETE("/:id", api.DeleteGitlabAgentConfig)
					agents.POST("/:id/configure", api.ConfigureGitlabAgent)
				}
				gitlab.PUT("/:id", api.UpdateGitlabConfig)
				gitlab.DELETE("/:id", api.DeleteGitlabConfig)
			}
			kube := settings.Group("/kube")
			{
				kube.GET("", api.GetKubeConfig)
				kube.POST("", api.UpdateKubeConfig)
				kube.POST("/validate", api.ValidateKubeConfig)
				kube.POST("/context", api.SetCurrentContext)

				configs := kube.Group("/configs")
				{
					configs.GET("", api.ListKubeConfigs)
					configs.POST("", api.UploadKubeConfig)
					configs.DELETE("/:id", api.DeleteKubeConfig)
				}
			}
			awsCfg := settings.Group("/aws")
			{
				awsCfg.GET("", api.ListAWSConfigs)
				awsCfg.POST("", api.CreateAWSConfig)
				awsCfg.DELETE("/:id", api.DeleteAWSConfig)

				eks := awsCfg.Group("/eks")
				{
					eks.POST("/clusters", api.ListEKSClusters)
					eks.POST("/import", api.ImportEKSClusters)
				}
			}
			contextMappings := settings.Group("/context-mappings")
			{
				contextMappings.GET("", api.ListContextMappings)
				contextMappings.POST("", api.UpsertContextMapping)
				contextMappings.DELETE("/:id", api.DeleteContextMapping)
			}
			settings.GET("/audit-logs", api.GetMyAuditLogs)
			settings.GET("/version", api.GetAppVersion)
		}

		// K8s API group - all Kubernetes related endpoints
		kubeGroup := apiGroup.Group("/kube")
		{
			// List valid contexts from kubeconfig with display name mappings
			kubeGroup.GET("/contexts", api.GetContexts)

			kubeGroup.GET("/namespaces", api.GetNamespaces)
			kubeGroup.GET("/nodes", api.GetNodes)
			kubeGroup.POST("/nodes/cordon", api.ToggleCordon)
			kubeGroup.POST("/nodes/drain", api.DrainNode)
			kubeGroup.GET("/events", api.GetEvents)

			kubeGroup.GET("/pods", api.GetPods)
			kubeGroup.GET("/deployments", api.GetDeployments)
			kubeGroup.GET("/replica-sets", api.GetReplicaSets)
			kubeGroup.GET("/replication-controllers", api.GetReplicationControllers)
			kubeGroup.GET("/jobs", api.GetJobs)
			kubeGroup.GET("/cron-jobs", api.GetCronJobs)
			kubeGroup.POST("/cron-jobs/suspend", api.ToggleCronJobSuspend)
			kubeGroup.POST("/cron-jobs/trigger", api.TriggerCronJob)
			kubeGroup.GET("/daemon-sets", api.GetDaemonSets)
			kubeGroup.GET("/stateful-sets", api.GetStatefulSets)

			kubeGroup.GET("/services", api.GetServices)
			kubeGroup.GET("/ingresses", api.GetIngresses)
			kubeGroup.GET("/endpoints", api.GetEndpoints)
			kubeGroup.GET("/ingress-classes", api.GetIngressClasses)
			kubeGroup.GET("/network-policies", api.GetNetworkPolicies)
			kubeGroup.GET("/port-forwards", api.GetPortForwards)

			kubeGroup.GET("/config-maps", api.GetConfigMaps)
			kubeGroup.GET("/secrets", api.GetSecrets)
			kubeGroup.GET("/resource-quotas", api.GetResourceQuotas)
			kubeGroup.GET("/limit-ranges", api.GetLimitRanges)
			kubeGroup.GET("/hpas", api.GetHPAs)
			kubeGroup.GET("/pdbs", api.GetPDBs)
			kubeGroup.GET("/priority-classes", api.GetPriorityClasses)
			kubeGroup.GET("/runtime-classes", api.GetRuntimeClasses)
			kubeGroup.GET("/leases", api.GetLeases)
			kubeGroup.GET("/mutating-webhooks", api.GetMutatingWebhooks)
			kubeGroup.GET("/validating-webhooks", api.GetValidatingWebhooks)

			kubeGroup.GET("/pvcs", api.GetPVCs)
			kubeGroup.GET("/pvs", api.GetPVs)
			kubeGroup.GET("/storage-classes", api.GetStorageClasses)

			kubeGroup.GET("/service-accounts", api.GetServiceAccounts)
			kubeGroup.GET("/cluster-roles", api.GetClusterRoles)
			kubeGroup.GET("/roles", api.GetRoles)
			kubeGroup.GET("/cluster-role-bindings", api.GetClusterRoleBindings)
			kubeGroup.GET("/role-bindings", api.GetRoleBindings)

			kubeGroup.GET("/scopes", api.GetResourceScopes)
			kubeGroup.GET("/resource", api.GetResourceDetails)
			kubeGroup.PUT("/resource", api.UpdateResource)
			kubeGroup.DELETE("/resource", api.DeleteResource)
			kubeGroup.POST("/resource/scale", api.ScaleResource)
			kubeGroup.POST("/resource/restart", api.RolloutRestartResource)
			kubeGroup.GET("/dashboard", api.GetDashboardSummary)
			kubeGroup.GET("/helm/releases", api.ListHelmReleases)

			// Metrics
			kubeGroup.GET("/metrics/pods", api.GetPodMetrics)

			// CRDs
			kubeGroup.GET("/crds", api.GetCustomResourceDefinitions)
			kubeGroup.GET("/crds/:crd_name/resources", api.GetCustomResources)
			kubeGroup.GET("/crds/:crd_name/resources/:name", api.GetCustomResourceDetails)

			// WS Handler
			kubeGroup.GET("/exec", api.HandleExec)
			kubeGroup.GET("/logs", api.HandleLogs)
		}
	}
}
