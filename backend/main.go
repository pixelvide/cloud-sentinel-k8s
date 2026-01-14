package main

import (
	"log"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"cloud-sentinel-k8s/api"
	"cloud-sentinel-k8s/auth"
	"cloud-sentinel-k8s/db"
)

func main() {
	// Initialize services
	db.InitDB()
	auth.InitOIDC()
	auth.InitJWT()

	// Reconfigure all GitLab K8s agents on startup (for non-persistent data directories)
	go api.ReconfigureAllAgentsOnStartup()

	r := gin.Default()

	config := cors.DefaultConfig()
	// Read CORS origins from environment variable, default to localhost:3000
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = frontendURL
	}
	config.AllowOrigins = strings.Split(corsOrigins, ",")
	config.AllowCredentials = true
	r.Use(cors.New(config))

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	authGroup := r.Group("/api/v1/auth")
	{
		authGroup.GET("/login", auth.LoginHandler)
		authGroup.GET("/callback", auth.CallbackHandler)
		authGroup.GET("/logout", auth.LogoutHandler)
	}

	apiGroup := r.Group("/api/v1")
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
			}
			contextMappings := settings.Group("/context-mappings")
			{
				contextMappings.GET("", api.ListContextMappings)
				contextMappings.POST("", api.UpsertContextMapping)
				contextMappings.DELETE("/:id", api.DeleteContextMapping)
			}
			settings.GET("/audit-logs", api.GetMyAuditLogs)
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
			kubeGroup.GET("/resource", api.GetResourceDetails) // <--- New endpoint
			kubeGroup.DELETE("/resource", api.DeleteResource)
			kubeGroup.GET("/dashboard", api.GetDashboardSummary)

			// WS Handler
			kubeGroup.GET("/exec", api.HandleExec)
			kubeGroup.GET("/logs", api.HandleLogs)
		}
	}

	log.Println("Server running on :8080")
	r.Run(":8080")
}
