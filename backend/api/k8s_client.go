package api

import (
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// GetClientInfo returns a kubernetes clientset and rest config for a specific context and user storage namespace
func GetClientInfo(storageNamespace string, contextName string) (*kubernetes.Clientset, *rest.Config, error) {
	if storageNamespace == "" {
		return nil, nil, os.ErrNotExist // Enforce isolation
	}

	kubeconfig := GetUserKubeConfigPath(storageNamespace)
	if _, err := os.Stat(kubeconfig); err != nil {
		return nil, nil, err
	}

	// Load raw config to process contexts
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return nil, nil, err
	}

	// If no context provided, use current
	if contextName == "" {
		contextName = config.CurrentContext
	}

	// Sanitize config: Replace "force-keyring" with "no" in glab exec args
	for _, authInfo := range config.AuthInfos {
		if authInfo.Exec != nil {
			for i, arg := range authInfo.Exec.Args {
				if arg == "force-keyring" {
					log.Printf("Sanitizing kubeconfig: replacing 'force-keyring' with 'no' for auth info")
					authInfo.Exec.Args[i] = "no"
				}
			}
			if storageNamespace != "" {
				glabConfigDir := filepath.Join("/data", storageNamespace, ".config", "glab-cli")
				authInfo.Exec.Env = append(authInfo.Exec.Env, clientcmdapi.ExecEnvVar{
					Name:  "GLAB_CONFIG_DIR",
					Value: glabConfigDir,
				})
			}
		}
	}

	clientConfig := clientcmd.NewNonInteractiveClientConfig(*config, contextName, &clientcmd.ConfigOverrides{}, nil)
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, nil, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, err
	}

	return clientset, restConfig, nil
}
