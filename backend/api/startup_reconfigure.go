package api

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"cloud-sentinel-k8s/db"
	"cloud-sentinel-k8s/models"
)

// ReconfigureAllAgentsOnStartup iterates over all GitLab agent configs and reconfigures them.
// This is useful when the data directory is not persisted and kubeconfig files need to be regenerated.
func ReconfigureAllAgentsOnStartup() {
	log.Println("[Startup] Starting reconfiguration of all GitLab agents...")

	// Get all agent configs that are marked as configured
	var agentConfigs []models.GitlabK8sAgentConfig
	if err := db.DB.Preload("GitlabConfig").Preload("GitlabConfig.User").Where("is_configured = ?", true).Find(&agentConfigs).Error; err != nil {
		log.Printf("[Startup] Failed to fetch agent configs: %v", err)
		return
	}

	if len(agentConfigs) == 0 {
		log.Println("[Startup] No configured agents found, skipping reconfiguration")
		return
	}

	log.Printf("[Startup] Found %d configured agent(s) to reconfigure", len(agentConfigs))

	for _, agentConfig := range agentConfigs {
		log.Printf("[Startup] Reconfiguring agent ID %s for GitLab host %s (User: %d)...",
			agentConfig.AgentID, agentConfig.GitlabConfig.Host, agentConfig.UserID)

		// Get storage namespace from GitlabConfig -> User
		var user models.User
		if err := db.DB.First(&user, agentConfig.UserID).Error; err != nil {
			log.Printf("[Startup] Failed to fetch user %d: %v", agentConfig.UserID, err)
			continue
		}

		glabConfigDir := GetUserGlabConfigDir(user.StorageNamespace)
		kubeConfigPath := GetUserKubeConfigPath(user.StorageNamespace)

		// Ensure directories exist
		if err := os.MkdirAll(glabConfigDir, 0777); err != nil {
			log.Printf("[Startup] Failed to create glab config directory: %v", err)
			continue
		}
		os.Chmod(glabConfigDir, 0777)

		kubeConfigDir := filepath.Dir(kubeConfigPath)
		if err := os.MkdirAll(kubeConfigDir, 0777); err != nil {
			log.Printf("[Startup] Failed to create kube config directory: %v", err)
			continue
		}
		os.Chmod(kubeConfigDir, 0777)

		// 1. Auth Login
		loginCmd := exec.Command("glab", "auth", "login", "--hostname", agentConfig.GitlabConfig.Host, "--token", agentConfig.GitlabConfig.Token)
		loginCmd.Env = append(os.Environ(), "GLAB_CONFIG_DIR="+glabConfigDir)
		if output, err := loginCmd.CombinedOutput(); err != nil {
			log.Printf("[Startup] Login failed for %s: %s", agentConfig.GitlabConfig.Host, string(output))
			continue
		}

		// 2. Update Kubeconfig
		protocol := "https://"
		if !agentConfig.GitlabConfig.IsHTTPS {
			protocol = "http://"
		}
		gitlabHost := protocol + agentConfig.GitlabConfig.Host

		agentCmd := exec.Command("glab", "cluster", "agent", "update-kubeconfig", "--agent", agentConfig.AgentID, "--repo", agentConfig.AgentRepo, "--use-context", "--cache-mode=no")
		agentCmd.Env = append(os.Environ(),
			"GLAB_CONFIG_DIR="+glabConfigDir,
			"KUBECONFIG="+kubeConfigPath,
			"GITLAB_HOST="+gitlabHost,
		)

		if output, err := agentCmd.CombinedOutput(); err != nil {
			log.Printf("[Startup] Failed to update kubeconfig for agent %s: %s", agentConfig.AgentID, string(output))
			continue
		}

		// Ensure the final kubeconfig has permissive permissions
		os.Chmod(kubeConfigPath, 0666)

		log.Printf("[Startup] Successfully reconfigured agent %s", agentConfig.AgentID)
	}

	log.Println("[Startup] Completed reconfiguration of all GitLab agents")
}
