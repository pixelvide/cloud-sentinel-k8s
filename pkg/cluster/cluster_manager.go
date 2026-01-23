package cluster

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/pixelvide/cloud-sentinel-k8s/pkg/kube"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/model"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/prometheus"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/rbac"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/utils"
	"gorm.io/gorm"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
)

// ClientSet holds the clients for a cluster
type ClientSet struct {
	Name       string
	Version    string // Kubernetes version
	K8sClient  *kube.K8sClient
	PromClient *prometheus.Client

	Configuration *rest.Config

	DiscoveredPrometheusURL string
	config                  string
	prometheusURL           string
}

type UserClient struct {
	ClientSet  *ClientSet
	LastUsedAt time.Time
	Error      string
}

type ClusterManager struct {
	clusters       map[string]*ClientSet
	userClients    map[string]map[uint]*UserClient // clusterName -> userID -> UserClient
	activeUsers    map[uint]time.Time              // userID -> lastActiveAt
	errors         map[string]string
	defaultContext string
	mu             sync.RWMutex
	activeUsersMu  sync.RWMutex
}

func createClientSetInCluster(name, prometheusURL string, skipSystemSync bool) (*ClientSet, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return newClientSet(name, config, prometheusURL, skipSystemSync)
}

func createClientSetFromConfig(name, content, prometheusURL string, skipSystemSync bool) (*ClientSet, error) {
	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(content))
	if err != nil {
		klog.Warningf("Failed to create REST config for cluster %s: %v", name, err)
		return nil, err
	}
	cs, err := newClientSet(name, restConfig, prometheusURL, skipSystemSync)
	if err != nil {
		return nil, err
	}
	cs.config = content

	return cs, nil
}

func newClientSet(name string, k8sConfig *rest.Config, prometheusURL string, skipSystemSync bool) (*ClientSet, error) {
	cs := &ClientSet{
		Name:          name,
		Configuration: k8sConfig,
		prometheusURL: prometheusURL,
	}
	var err error
	cs.K8sClient, err = kube.NewClient(kube.ClientOptions{
		Config: k8sConfig,
	})
	if err != nil {
		klog.Warningf("Failed to create k8s client for cluster %s: %v", name, err)
		return nil, err
	}
	if prometheusURL == "" {
		if !skipSystemSync {
			prometheusURL = discoveryPrometheusURL(cs.K8sClient)
			if prometheusURL != "" {
				cs.DiscoveredPrometheusURL = prometheusURL
				klog.Infof("Discovered Prometheus URL for cluster %s: %s", name, cs.DiscoveredPrometheusURL)
			}
		} else {
			klog.V(2).Infof("Skipping Prometheus discovery for cluster %s (SkipSystemSync=true)", name)
		}
	}
	if prometheusURL != "" {
		var rt = http.DefaultTransport
		var err error
		if isClusterLocalURL(prometheusURL) {
			rt, err = createK8sProxyTransport(k8sConfig, prometheusURL)
			if err != nil {
				klog.Warningf("Failed to create k8s proxy transport for cluster %s: %v, using direct connection", name, err)
			} else {
				klog.Infof("Using k8s API proxy for Prometheus in cluster %s", name)
			}
		}
		cs.PromClient, err = prometheus.NewClientWithRoundTripper(prometheusURL, rt)
		if err != nil {
			klog.Warningf("Failed to create Prometheus client for cluster %s, some features may not work as expected, err: %v", name, err)
		}
	}
	if !skipSystemSync {
		v, err := cs.K8sClient.ClientSet.Discovery().ServerVersion()
		if err != nil {
			klog.Warningf("Failed to get server version for cluster %s: %v", name, err)
		} else {
			cs.Version = v.String()
		}
	} else {
		cs.Version = "unknown (skipped)"
		klog.V(2).Infof("Skipping server version check for cluster %s (SkipSystemSync=true)", name)
	}
	klog.Infof("Loaded K8s client for cluster: %s, version: %s", name, cs.Version)
	return cs, nil
}

func isClusterLocalURL(urlStr string) bool {
	return strings.Contains(urlStr, ".svc.cluster.local") || strings.Contains(urlStr, ".svc:")
}

func createK8sProxyTransport(k8sConfig *rest.Config, prometheusURL string) (*k8sProxyTransport, error) {
	parsedURL, err := url.Parse(prometheusURL)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(parsedURL.Host, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid cluster local URL format")
	}
	svcName := parts[0]
	namespace := parts[1]

	transport, err := rest.TransportFor(k8sConfig)
	if err != nil {
		return nil, err
	}

	transportWrapper := &k8sProxyTransport{
		transport:    transport,
		apiServerURL: k8sConfig.Host,
		namespace:    namespace,
		svcName:      svcName,
		scheme:       parsedURL.Scheme,
	}
	transportWrapper.port = parsedURL.Port()
	if transportWrapper.port == "" {
		if parsedURL.Scheme == "https" {
			transportWrapper.port = "443"
		} else {
			transportWrapper.port = "80"
		}
	}

	return transportWrapper, nil
}

type k8sProxyTransport struct {
	transport    http.RoundTripper
	apiServerURL string
	namespace    string
	svcName      string
	scheme       string
	port         string
}

func (t *k8sProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	proxyURL, err := url.Parse(t.apiServerURL)
	if err != nil {
		return nil, err
	}
	req.URL.Scheme = proxyURL.Scheme
	req.URL.Host = proxyURL.Host

	servicePath := fmt.Sprintf("/api/v1/namespaces/%s/services/%s:%s/proxy", t.namespace, t.svcName, t.port)
	req.URL.Path = servicePath + req.URL.Path

	return t.transport.RoundTrip(req)
}

func (cm *ClusterManager) GetActiveClusters() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	names := make([]string, 0, len(cm.clusters))
	for name := range cm.clusters {
		names = append(names, name)
	}
	return names
}

func (cm *ClusterManager) GetCluster(name string) (*ClientSet, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if cs, ok := cm.clusters[name]; ok {
		return cs, nil
	}
	return nil, fmt.Errorf("cluster %s not found", name)
}

func (cm *ClusterManager) GetClientSet(clusterName string, user *model.User) (*ClientSet, error) {
	klog.V(2).Infof("GetClientSet called for cluster: %s, user: %v", clusterName, user)

	cm.mu.RLock()
	if clusterName == "" {
		if cm.defaultContext != "" {
			ctx := cm.defaultContext
			cm.mu.RUnlock()
			return cm.GetClientSet(ctx, user)
		}
		// If no default context is set, return the first available shared cluster
		for _, cs := range cm.clusters {
			cm.mu.RUnlock()
			return cs, nil
		}
		cm.mu.RUnlock()
		return nil, fmt.Errorf("no clusters available")
	}

	// Check user client first if applicable
	if user != nil {
		if userMap, ok := cm.userClients[clusterName]; ok {
			if uc, ok := userMap[user.ID]; ok {
				if uc.ClientSet != nil {
					uc.LastUsedAt = time.Now()
					cm.mu.RUnlock()
					return uc.ClientSet, nil
				}
				if uc.Error != "" {
					cm.mu.RUnlock()
					return nil, fmt.Errorf("user-specific cluster client error: %s", uc.Error)
				}
			}
		}
	}

	cs, ok := cm.clusters[clusterName]
	if ok {
		cm.mu.RUnlock()
		return cs, nil
	}
	cm.mu.RUnlock()

	// If not found in shared or user cache, it might need user-level auth and wasn't synced yet
	if user != nil {
		cluster, err := model.GetClusterByName(clusterName)
		if err == nil && cluster.Enable && cluster.SkipSystemSync {
			cm.mu.Lock()
			// Double check
			if userMap, ok := cm.userClients[clusterName]; ok {
				if uc, ok := userMap[user.ID]; ok && uc.ClientSet != nil {
					uc.LastUsedAt = time.Now()
					cm.mu.Unlock()
					return uc.ClientSet, nil
				}
			} else {
				cm.userClients[clusterName] = make(map[uint]*UserClient)
			}

			klog.Infof("Creating on-demand user client for user %d in cluster %s", user.ID, clusterName)
			uc, err := buildUserClientSet(cluster, user)
			if err != nil {
				cm.mu.Unlock()
				return nil, err
			}
			cm.userClients[clusterName][user.ID] = uc
			cm.mu.Unlock()
			return uc.ClientSet, nil
		}
	}

	return nil, fmt.Errorf("cluster not found or not initialized: %s", clusterName)
}

func ImportClustersFromKubeconfig(kubeconfig *clientcmdapi.Config) int64 {
	if len(kubeconfig.Contexts) == 0 {
		return 0
	}

	importedCount := 0
	for contextName, context := range kubeconfig.Contexts {
		config := clientcmdapi.NewConfig()
		skipSystemSync := false
		config.Contexts = map[string]*clientcmdapi.Context{
			contextName: context,
		}
		config.CurrentContext = contextName
		config.Clusters = map[string]*clientcmdapi.Cluster{
			context.Cluster: kubeconfig.Clusters[context.Cluster],
		}
		authInfo := kubeconfig.AuthInfos[context.AuthInfo]
		if authInfo != nil && authInfo.Exec != nil {
			authInfo, _, skipSystemSync = processAuthInfo(authInfo)
		}
		config.AuthInfos = map[string]*clientcmdapi.AuthInfo{
			context.AuthInfo: authInfo,
		}
		configStr, err := clientcmd.Write(*config)
		if err != nil {
			continue
		}

		// Sanitize cluster name: replace / with -
		sanitizedName := strings.ReplaceAll(contextName, "/", "-")

		cluster := model.Cluster{
			Name:           sanitizedName,
			Config:         model.SecretString(configStr),
			IsDefault:      contextName == kubeconfig.CurrentContext,
			SkipSystemSync: skipSystemSync,
		}
		if _, err := model.GetClusterByName(sanitizedName); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := model.AddCluster(&cluster); err != nil {
					continue
				}
				importedCount++
				klog.Infof("Imported cluster success: %s (original context: %s)", sanitizedName, contextName)
			}
			continue
		}
	}
	return int64(importedCount)
}

func processAuthInfo(authInfo *clientcmdapi.AuthInfo) (*clientcmdapi.AuthInfo, bool, bool) {
	if strings.Contains(authInfo.Exec.Command, "glab") {
		return processGlabAuth(authInfo), true, true
	} else if strings.Contains(authInfo.Exec.Command, "aws") || strings.Contains(authInfo.Exec.Command, "aws-iam-authenticator") {
		return processAWSAuth(authInfo), true, true
	}
	return authInfo, false, false
}

func processGlabAuth(authInfo *clientcmdapi.AuthInfo) *clientcmdapi.AuthInfo {
	// Create a copy to avoid modifying the original kubeconfig
	copiedAuthInfo := authInfo.DeepCopy()
	if copiedAuthInfo.Exec.Command != "glab" {
		copiedAuthInfo.Exec.Command = "glab"
	}

	// Normalize --cache-mode to "none"
	for i, arg := range copiedAuthInfo.Exec.Args {
		if arg == "--cache-mode" && i+1 < len(copiedAuthInfo.Exec.Args) {
			if copiedAuthInfo.Exec.Args[i+1] != "no" {
				copiedAuthInfo.Exec.Args[i+1] = "no"
			}
		} else if strings.HasPrefix(arg, "--cache-mode=") {
			if arg != "--cache-mode=no" {
				copiedAuthInfo.Exec.Args[i] = "--cache-mode=no"
			}
		}
	}
	return copiedAuthInfo
}

func processAWSAuth(authInfo *clientcmdapi.AuthInfo) *clientcmdapi.AuthInfo {
	copiedAuthInfo := authInfo.DeepCopy()

	region := ""
	clusterID := ""
	var filteredArgs []string

	// Extract region and cluster ID from args, and filter them out if we're converting from 'aws eks'
	isAwsEks := strings.HasSuffix(copiedAuthInfo.Exec.Command, "aws")
	for i := 0; i < len(copiedAuthInfo.Exec.Args); i++ {
		arg := copiedAuthInfo.Exec.Args[i]
		switch {
		case (arg == "--region") && i+1 < len(copiedAuthInfo.Exec.Args):
			region = copiedAuthInfo.Exec.Args[i+1]
			i++
		case strings.HasPrefix(arg, "--region="):
			region = strings.TrimPrefix(arg, "--region=")
		case (arg == "--cluster-name" || arg == "--cluster-id") && i+1 < len(copiedAuthInfo.Exec.Args):
			clusterID = copiedAuthInfo.Exec.Args[i+1]
			i++
		case strings.HasPrefix(arg, "--cluster-name="):
			clusterID = strings.TrimPrefix(arg, "--cluster-name=")
		case strings.HasPrefix(arg, "--cluster-id="):
			clusterID = strings.TrimPrefix(arg, "--cluster-id=")
		case !isAwsEks:
			filteredArgs = append(filteredArgs, arg)
		}
	}

	if isAwsEks {
		filteredArgs = []string{"token"}
		if clusterID != "" {
			filteredArgs = append(filteredArgs, "-i", clusterID)
		}
	}
	copiedAuthInfo.Exec.Args = filteredArgs

	// Handle Environment Variables
	hasRegion := false
	hasStsRegional := false
	for _, env := range copiedAuthInfo.Exec.Env {
		if env.Name == "AWS_REGION" {
			hasRegion = true
		}
		if env.Name == "AWS_STS_REGIONAL_ENDPOINTS" {
			hasStsRegional = true
		}
	}

	if region != "" && !hasRegion {
		copiedAuthInfo.Exec.Env = append(copiedAuthInfo.Exec.Env, clientcmdapi.ExecEnvVar{
			Name:  "AWS_REGION",
			Value: region,
		})
	}

	if !hasStsRegional {
		copiedAuthInfo.Exec.Env = append(copiedAuthInfo.Exec.Env, clientcmdapi.ExecEnvVar{
			Name:  "AWS_STS_REGIONAL_ENDPOINTS",
			Value: "regional",
		})
	}

	copiedAuthInfo.Exec.Command = "aws-iam-authenticator"
	return copiedAuthInfo
}

var (
	syncNow = make(chan struct{}, 1)
)

func syncClusters(cm *ClusterManager) error {
	klog.Infof("Starting logs cluster sync")
	clusters, err := model.ListClusters()
	if err != nil {
		klog.Warningf("list cluster err: %v", err)
		time.Sleep(5 * time.Second)
		return err
	}
	klog.Infof("Found %d clusters from database", len(clusters))

	now := time.Now()
	cm.activeUsersMu.RLock()
	activeUserIDs := cm.getActiveUserIDs(now)
	cm.activeUsersMu.RUnlock()

	dbClusterMap := make(map[string]*model.Cluster)
	for _, cluster := range clusters {
		dbClusterMap[cluster.Name] = cluster
		cm.updateClusterStatus(cluster, activeUserIDs, now)
	}

	cm.cleanupDeletedClusters(dbClusterMap)

	klog.Infof("Cluster sync completed, active: %d shared, %d clusters with user-level sync", len(cm.clusters), len(cm.userClients))
	return nil
}

// getActiveUserIDs returns a list of active user IDs.
// Caller is responsible for locking activeUsersMu if needed, but this function does not lock it itself?
// Wait, looking at usage:
// In syncClusters, I'm locking it outside.
// Let's make this function NOT lock, but assume caller holds the lock OR it's safe.
// Actually, to be safe and consistent, let's make it internal helper where caller handles lock.
func (cm *ClusterManager) getActiveUserIDs(now time.Time) []uint {
	var activeUserIDs []uint
	for userID, lastActiveAt := range cm.activeUsers {
		if now.Sub(lastActiveAt) <= 30*time.Minute {
			activeUserIDs = append(activeUserIDs, userID)
		}
	}
	return activeUserIDs
}

func (cm *ClusterManager) updateClusterStatus(cluster *model.Cluster, activeUserIDs []uint, now time.Time) {
	if !cluster.Enable || len(activeUserIDs) == 0 {
		cm.stopClusterSync(cluster)
		return
	}

	if !cluster.SkipSystemSync {
		cm.handleSharedSync(cluster)
	} else {
		cm.handleUserLevelSync(cluster, activeUserIDs, now)
	}
}

func (cm *ClusterManager) stopClusterSync(cluster *model.Cluster) {
	if cs, ok := cm.clusters[cluster.Name]; ok {
		klog.Infof("Stopping shared sync for cluster %s (disabled or no active users)", cluster.Name)
		delete(cm.clusters, cluster.Name)
		cs.K8sClient.Stop(cluster.Name)
	}
	if userMap, ok := cm.userClients[cluster.Name]; ok {
		for userID, uc := range userMap {
			klog.Infof("Stopping user client sync for user %d in cluster %s", userID, cluster.Name)
			if uc.ClientSet != nil {
				uc.ClientSet.K8sClient.Stop(fmt.Sprintf("%s-%d", cluster.Name, userID))
			}
		}
		delete(cm.userClients, cluster.Name)
	}
}

func (cm *ClusterManager) handleSharedSync(cluster *model.Cluster) {
	cm.mu.RLock()
	current, currentExist := cm.clusters[cluster.Name]
	cm.mu.RUnlock()

	if shouldUpdateCluster(current, cluster) {
		klog.Infof("Updating/Adding shared cluster %s", cluster.Name)
		clientSet, err := buildClientSet(cluster)

		cm.mu.Lock()
		if err != nil {
			klog.Errorf("Failed to build shared k8s client for cluster %s: %v", cluster.Name, err)
			cm.errors[cluster.Name] = err.Error()
			cm.mu.Unlock()
			return
		}

		if currentExist {
			// Stop the old one
			if old, ok := cm.clusters[cluster.Name]; ok {
				old.K8sClient.Stop(cluster.Name)
			}
		}

		delete(cm.errors, cluster.Name)
		cm.clusters[cluster.Name] = clientSet
		cm.mu.Unlock()
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()
	// Stop any user clients for this cluster as it's now shared
	if userMap, ok := cm.userClients[cluster.Name]; ok {
		for userID, uc := range userMap {
			if uc.ClientSet != nil {
				uc.ClientSet.K8sClient.Stop(fmt.Sprintf("%s-%d", cluster.Name, userID))
			}
		}
		delete(cm.userClients, cluster.Name)
	}
}

func (cm *ClusterManager) handleUserLevelSync(cluster *model.Cluster, activeUserIDs []uint, now time.Time) {
	cm.mu.Lock()
	// Stop shared client if it was previously shared
	if cs, ok := cm.clusters[cluster.Name]; ok {
		delete(cm.clusters, cluster.Name)
		cs.K8sClient.Stop(cluster.Name)
	}

	if _, ok := cm.userClients[cluster.Name]; !ok {
		cm.userClients[cluster.Name] = make(map[uint]*UserClient)
	}
	cm.mu.Unlock()

	activeUserSet := make(map[uint]bool)

	for _, userID := range activeUserIDs {
		user, err := model.GetUserByID(uint64(userID))
		if err != nil || !rbac.CanAccessCluster(*user, cluster.Name) {
			continue
		}
		activeUserSet[userID] = true

		cm.mu.RLock()
		userMap := cm.userClients[cluster.Name]
		uc, exists := userMap[userID]
		needsUpdate := !exists || shouldUpdateUserClient(uc, cluster)
		cm.mu.RUnlock()

		if needsUpdate {
			klog.Infof("Updating/Adding user cluster client for user %d in cluster %s", userID, cluster.Name)
			newUc, err := buildUserClientSet(cluster, user)

			cm.mu.Lock()
			// Re-fetch map in case it changed
			userMap = cm.userClients[cluster.Name]
			if userMap == nil {
				cm.userClients[cluster.Name] = make(map[uint]*UserClient)
				userMap = cm.userClients[cluster.Name]
			}

			if existingUc, ok := userMap[userID]; ok && existingUc.ClientSet != nil {
				existingUc.ClientSet.K8sClient.Stop(fmt.Sprintf("%s-%d", cluster.Name, userID))
			}

			if err != nil {
				klog.Errorf("Failed to build user client for user %d in cluster %s: %v", userID, cluster.Name, err)
				userMap[userID] = &UserClient{LastUsedAt: now, Error: err.Error()}
			} else {
				userMap[userID] = newUc
			}
			cm.mu.Unlock()
		} else {
			cm.mu.Lock()
			if userMap, ok := cm.userClients[cluster.Name]; ok {
				if uc, ok := userMap[userID]; ok {
					uc.LastUsedAt = now
				}
			}
			cm.mu.Unlock()
		}
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()
	if userMap, ok := cm.userClients[cluster.Name]; ok {
		for userID, uc := range userMap {
			if !activeUserSet[userID] {
				klog.Infof("Removing inactive user client for user %d in cluster %s", userID, cluster.Name)
				if uc.ClientSet != nil {
					uc.ClientSet.K8sClient.Stop(fmt.Sprintf("%s-%d", cluster.Name, userID))
				}
				delete(userMap, userID)
			}
		}
	}
}

func (cm *ClusterManager) cleanupDeletedClusters(dbClusterMap map[string]*model.Cluster) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	for name, cs := range cm.clusters {
		if _, ok := dbClusterMap[name]; !ok {
			klog.Infof("Removing shared cluster %s (deleted from DB)", name)
			delete(cm.clusters, name)
			cs.K8sClient.Stop(name)
		}
	}
	for name, userMap := range cm.userClients {
		if _, ok := dbClusterMap[name]; !ok {
			klog.Infof("Removing user clusters for %s (deleted from DB)", name)
			for userID, uc := range userMap {
				if uc.ClientSet != nil {
					uc.ClientSet.K8sClient.Stop(fmt.Sprintf("%s-%d", name, userID))
				}
			}
			delete(cm.userClients, name)
		}
	}
}

func shouldUpdateUserClient(uc *UserClient, cluster *model.Cluster) bool {
	if uc.ClientSet == nil {
		return true // It had an error before
	}
	if uc.ClientSet.config != string(cluster.Config) {
		return true
	}
	if uc.ClientSet.prometheusURL != cluster.PrometheusURL {
		return true
	}
	return false
}

func buildUserClientSet(cluster *model.Cluster, user *model.User) (*UserClient, error) {
	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(cluster.Config))
	if err != nil {
		return nil, err
	}

	userConfig, err := model.GetUserConfig(user.ID)
	if err != nil {
		return nil, err
	}

	if restConfig.ExecProvider != nil {
		if strings.Contains(restConfig.ExecProvider.Command, "glab") {
			glabConfigDir, err := utils.GetUserGlabConfigDir(userConfig.StorageNamespace)
			if err != nil {
				return nil, err
			}
			restConfig.ExecProvider.Env = append(restConfig.ExecProvider.Env,
				clientcmdapi.ExecEnvVar{Name: "GLAB_CONFIG_DIR", Value: glabConfigDir},
			)
		}

		if strings.Contains(restConfig.ExecProvider.Command, "aws") || strings.Contains(restConfig.ExecProvider.Command, "aws-iam-authenticator") {
			awsCredsPath := utils.GetUserAWSCredentialsPath(userConfig.StorageNamespace)
			restConfig.ExecProvider.Env = append(restConfig.ExecProvider.Env,
				clientcmdapi.ExecEnvVar{Name: "AWS_SHARED_CREDENTIALS_FILE", Value: awsCredsPath},
			)
		}
	}

	// Create new client with cache ENABLED (user wants sync)
	k8sClient, err := kube.NewClient(kube.ClientOptions{
		Config:       restConfig,
		DisableCache: false,
	})
	if err != nil {
		return nil, err
	}

	cs := &ClientSet{
		Name:          cluster.Name,
		Configuration: restConfig,
		prometheusURL: cluster.PrometheusURL,
		K8sClient:     k8sClient,
		config:        string(cluster.Config),
	}

	// Discovery and Prometheus discovery (optional, could be improved)
	v, err := cs.K8sClient.ClientSet.Discovery().ServerVersion()
	if err == nil {
		cs.Version = v.String()
		klog.Infof("buildUserClientSet: Successfully fetched version %s for cluster %s", cs.Version, cluster.Name)
	} else {
		klog.Warningf("buildUserClientSet: Failed to fetch version for cluster %s: %v", cluster.Name, err)
	}

	return &UserClient{
		ClientSet:  cs,
		LastUsedAt: time.Now(),
	}, nil
}

// shouldUpdateCluster decides whether the cached ClientSet needs to be updated
// based on the desired state from the database.
func shouldUpdateCluster(cs *ClientSet, cluster *model.Cluster) bool {
	// enable/disable toggle
	if (cs == nil && cluster.Enable) || (cs != nil && !cluster.Enable) {
		klog.Infof("Cluster %s status changed, updating, enabled -> %v", cluster.Name, cluster.Enable)
		return true
	}
	if cs == nil && !cluster.Enable {
		return false
	}

	if cs == nil || cs.K8sClient == nil || cs.K8sClient.ClientSet == nil {
		return true
	}

	// kubeconfig change
	if cs.config != string(cluster.Config) {
		klog.Infof("Kubeconfig changed for cluster %s, updating", cluster.Name)
		return true
	}

	// prometheus URL change
	if cs.prometheusURL != cluster.PrometheusURL {
		klog.Infof("Prometheus URL changed for cluster %s, updating", cluster.Name)
		return true
	}

	// k8s version change
	// If SkipSystemSync is true, we skip the version check to avoid auth errors on user-only clusters
	if cluster.SkipSystemSync {
		return false
	}

	// TODO: Replace direct ClientSet.Discovery() call with a small DiscoveryInterface.
	// current code depends on *kubernetes.Clientset, which is hard to mock in tests.
	version, err := cs.K8sClient.ClientSet.Discovery().ServerVersion()
	if err != nil {
		klog.Warningf("Failed to get server version for cluster %s: %v", cluster.Name, err)
	} else if version.String() != cs.Version {
		klog.Infof("Server version changed for cluster %s, updating, old: %s, new: %s", cluster.Name, cs.Version, version.String())
		return true
	}

	return false
}

func buildClientSet(cluster *model.Cluster) (*ClientSet, error) {
	if cluster.InCluster {
		return createClientSetInCluster(cluster.Name, cluster.PrometheusURL, cluster.SkipSystemSync)
	}
	return createClientSetFromConfig(cluster.Name, string(cluster.Config), cluster.PrometheusURL, cluster.SkipSystemSync)
}

func NewClusterManager() (*ClusterManager, error) {
	cm := new(ClusterManager)
	cm.clusters = make(map[string]*ClientSet)
	cm.userClients = make(map[string]map[uint]*UserClient)
	cm.activeUsers = make(map[uint]time.Time)
	cm.errors = make(map[string]string)

	// Start cleanup routine
	go cm.startCleanupRoutine()

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := syncClusters(cm); err != nil {
					klog.Warningf("Failed to sync clusters: %v", err)
				}
			case <-syncNow:
				if err := syncClusters(cm); err != nil {
					klog.Warningf("Failed to sync clusters: %v", err)
				}
			}
		}
	}()

	if err := syncClusters(cm); err != nil {
		klog.Warningf("Failed to sync clusters: %v", err)
	}
	return cm, nil
}

func (cm *ClusterManager) UpdateUserActivity(userID uint) {
	cm.activeUsersMu.RLock()
	lastActive, exists := cm.activeUsers[userID]
	cm.activeUsersMu.RUnlock()

	if exists && time.Since(lastActive) < 5*time.Minute {
		return
	}

	cm.activeUsersMu.Lock()
	defer cm.activeUsersMu.Unlock()
	cm.activeUsers[userID] = time.Now()
	// Trigger sync immediately for this user if not already running
	select {
	case syncNow <- struct{}{}:
	default:
	}
}

func (cm *ClusterManager) startCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()

	// 30 minutes TTL
	ttl := 30 * time.Minute

	for range ticker.C {
		// Cleanup active users
		cm.activeUsersMu.Lock()
		now := time.Now()
		for userID, lastActiveAt := range cm.activeUsers {
			if now.Sub(lastActiveAt) > ttl {
				delete(cm.activeUsers, userID)
			}
		}
		cm.activeUsersMu.Unlock()

		// Cleanup user clients
		cm.mu.Lock()
		for clusterName, users := range cm.userClients {
			for userID, client := range users {
				if time.Since(client.LastUsedAt) > ttl {
					klog.V(2).Infof("Evicting user client for user %d in cluster %s (last used: %v)", userID, clusterName, client.LastUsedAt)
					if client.ClientSet != nil {
						client.ClientSet.K8sClient.Stop(fmt.Sprintf("%s-%d", clusterName, userID))
					}
					delete(users, userID)
				}
			}
			// Clean up empty cluster maps
			if len(users) == 0 {
				delete(cm.userClients, clusterName)
			}
		}
		cm.mu.Unlock()
	}
}
