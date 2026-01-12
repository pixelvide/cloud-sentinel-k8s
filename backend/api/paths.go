package api

import (
	"path/filepath"
)

// GetUserGlabConfigDir returns the directory path for a user's glab configuration
func GetUserGlabConfigDir(storageNamespace string) string {
	return filepath.Join("/data", storageNamespace, ".config", "glab-cli")
}

// GetUserKubeConfigPath returns the path to a user's kubeconfig file
func GetUserKubeConfigPath(storageNamespace string) string {
	return filepath.Join("/data", storageNamespace, ".config", "kube", "config")
}
