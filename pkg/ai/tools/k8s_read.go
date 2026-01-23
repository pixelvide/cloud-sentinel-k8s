package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pixelvide/cloud-sentinel-k8s/pkg/cluster"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/helm"
	openai "github.com/sashabaranov/go-openai"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type ClientSetKey struct{}

func GetClientSet(ctx context.Context) (*cluster.ClientSet, error) {
	cs, ok := ctx.Value(ClientSetKey{}).(*cluster.ClientSet)
	if !ok || cs == nil {
		klog.Warningf("K8s Tool: Kubernetes client not found in context (key: %T)", ClientSetKey{})
		return nil, fmt.Errorf("kubernetes client not found in context")
	}
	klog.V(2).Infof("K8s Tool: Found client for cluster %s", cs.Name)
	return cs, nil
}

// --- List Pods Tool ---

type ListPodsTool struct{}

func (t *ListPodsTool) Name() string { return "list_pods" }

func (t *ListPodsTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "list_pods",
			Description: "List pods in a namespace, optionally filtered by status",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"namespace": {
						"type": "string",
						"description": "The namespace to list pods from. If empty, lists from all namespaces."
					},
					"status_filter": {
						"type": "string",
						"enum": ["Running", "Pending", "Failed", "Succeeded", "Unknown"],
						"description": "Filter pods by status phase."
					}
				}
			}`),
		},
	}
}

func (t *ListPodsTool) Execute(ctx context.Context, args string) (string, error) {
	var params struct {
		Namespace    string `json:"namespace"`
		StatusFilter string `json:"status_filter"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}

	cs, err := GetClientSet(ctx)
	if err != nil {
		return "", err
	}

	pods, err := cs.K8sClient.ClientSet.CoreV1().Pods(params.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	var results []string
	for _, pod := range pods.Items {
		if params.StatusFilter != "" && string(pod.Status.Phase) != params.StatusFilter {
			continue
		}

		restarts := 0
		for _, status := range pod.Status.ContainerStatuses {
			restarts += int(status.RestartCount)
		}

		results = append(results, fmt.Sprintf("%s (Status: %s, Restarts: %d, IP: %s)",
			pod.Name, pod.Status.Phase, restarts, pod.Status.PodIP))
	}

	if len(results) == 0 {
		return "No pods found.", nil
	}

	// Limit output to prevent token overflow
	if len(results) > 50 {
		return strings.Join(results[:50], "\n") + fmt.Sprintf("\n... and %d more", len(results)-50), nil
	}

	return strings.Join(results, "\n"), nil
}

// --- Get Pod Logs Tool ---

type GetPodLogsTool struct{}

func (t *GetPodLogsTool) Name() string { return "get_pod_logs" }

func (t *GetPodLogsTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "get_pod_logs",
			Description: "Get logs from a specific pod",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"namespace": {
						"type": "string",
						"description": "The namespace of the pod."
					},
					"pod_name": {
						"type": "string",
						"description": "The name of the pod."
					},
					"container": {
						"type": "string",
						"description": "Optional container name."
					},
					"lines": {
						"type": "integer",
						"description": "Number of lines to retrieve (max 100). Defaults to 50."
					}
				},
				"required": ["namespace", "pod_name"]
			}`),
		},
	}
}

func (t *GetPodLogsTool) Execute(ctx context.Context, args string) (string, error) {
	var params struct {
		Namespace string `json:"namespace"`
		PodName   string `json:"pod_name"`
		Container string `json:"container"`
		Lines     int64  `json:"lines"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}

	if params.Lines <= 0 {
		params.Lines = 50
	}
	if params.Lines > 100 {
		params.Lines = 100
	}

	cs, err := GetClientSet(ctx)
	if err != nil {
		return "", err
	}

	opts := &corev1.PodLogOptions{
		TailLines: &params.Lines,
	}
	if params.Container != "" {
		opts.Container = params.Container
	}

	req := cs.K8sClient.ClientSet.CoreV1().Pods(params.Namespace).GetLogs(params.PodName, opts)
	logs, err := req.DoRaw(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %w", err)
	}

	return string(logs), nil
}

// --- Describe Resource Tool ---

type DescribeResourceTool struct{}

func (t *DescribeResourceTool) Name() string { return "describe_resource" }

func (t *DescribeResourceTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "describe_resource",
			Description: "Get details (JSON) of a specific resource",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"namespace": {
						"type": "string",
						"description": "The namespace of the resource."
					},
					"kind": {
						"type": "string",
						"description": "The kind of resource (Pod, Deployment, Service, etc)."
					},
					"name": {
						"type": "string",
						"description": "The name of the resource."
					}
				},
				"required": ["namespace", "kind", "name"]
			}`),
		},
	}
}

func (t *DescribeResourceTool) Execute(ctx context.Context, args string) (string, error) {
	var params struct {
		Namespace string `json:"namespace"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}

	cs, err := GetClientSet(ctx)
	if err != nil {
		return "", err
	}

	var obj interface{}
	var getErr error

	switch strings.ToLower(params.Kind) {
	case "pod":
		obj, getErr = cs.K8sClient.ClientSet.CoreV1().Pods(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
	case "deployment":
		obj, getErr = cs.K8sClient.ClientSet.AppsV1().Deployments(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
	case "service":
		obj, getErr = cs.K8sClient.ClientSet.CoreV1().Services(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
	case "node":
		obj, getErr = cs.K8sClient.ClientSet.CoreV1().Nodes().Get(ctx, params.Name, metav1.GetOptions{})
	default:
		return "", fmt.Errorf("unsupported resource kind: %s", params.Kind)
	}

	if getErr != nil {
		return "", getErr
	}

	// Serialize to JSON for the LLM
	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// --- List Resources Tool ---

type ListResourcesTool struct{}

func (t *ListResourcesTool) Name() string { return "list_resources" }

func (t *ListResourcesTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "list_resources",
			Description: "List Kubernetes resources of a specific kind in a namespace, optionally filtered by labels. This tool also can be used to list the Helm Release by passing on of helmrelease|helmreleases|hr| as kind.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"kind": {
						"type": "string",
						"description": "The kind of resource to list (Pod, Service, Deployment, ReplicaSet, StatefulSet, DaemonSet, Job, CronJob, ConfigMap, Secret, Namespace, Node, Ingress, Event)."
					},
					"namespace": {
						"type": "string",
						"description": "The namespace to list resources from. If empty, lists from all namespaces (if applicable)."
					},
					"name_filter": {
						"type": "string",
						"description": "Optional name filter. Only resources containing this string in their name will be returned."
					},
					"label_selector": {
						"type": "string",
						"description": "Optional label selector to filter results (e.g., 'app=nginx')."
					}
				},
				"required": ["kind"]
			}`),
		},
	}
}

func (t *ListResourcesTool) Execute(ctx context.Context, args string) (string, error) {
	var params struct {
		Kind          string `json:"kind"`
		Namespace     string `json:"namespace"`
		NameFilter    string `json:"name_filter"`
		LabelSelector string `json:"label_selector"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}

	cs, err := GetClientSet(ctx)
	if err != nil {
		return "", err
	}

	opts := metav1.ListOptions{
		LabelSelector: params.LabelSelector,
	}

	var results []string
	kind := strings.ToLower(params.Kind)

	if kind == "" {
		// Default search across common kinds
		commonKinds := []string{"pod", "service", "deployment", "ingress"}
		var combinedResults []string
		for _, k := range commonKinds {
			res, err := t.listByKind(ctx, cs, k, params.Namespace, params.NameFilter, opts)
			if err == nil && len(res) > 0 {
				combinedResults = append(combinedResults, fmt.Sprintf("--- %s ---", strings.ToUpper(k)))
				combinedResults = append(combinedResults, res...)
			}
		}
		results = combinedResults
	} else {
		var err error
		results, err = t.listByKind(ctx, cs, kind, params.Namespace, params.NameFilter, opts)
		if err != nil {
			return "", err
		}
	}

	if len(results) == 0 {
		return fmt.Sprintf("No %s found.", params.Kind), nil
	}

	// Limit output to prevent token overflow
	if len(results) > 100 {
		return strings.Join(results[:100], "\n") + fmt.Sprintf("\n... and %d more", len(results)-100), nil
	}

	return strings.Join(results, "\n"), nil
}

func (t *ListResourcesTool) listByKind(ctx context.Context, cs *cluster.ClientSet, kind, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	var results []string
	var err error

	switch kind {
	case "pod", "pods":
		results, err = t.listPods(ctx, cs, ns, filter, opts)
	case "service", "services", "svc":
		results, err = t.listServices(ctx, cs, ns, filter, opts)
	case "deployment", "deployments", "deploy":
		results, err = t.listDeployments(ctx, cs, ns, filter, opts)
	case "replicaset", "replicasets", "rs":
		results, err = t.listReplicaSets(ctx, cs, ns, filter, opts)
	case "statefulset", "statefulsets", "sts":
		results, err = t.listStatefulSets(ctx, cs, ns, filter, opts)
	case "daemonset", "daemonsets", "ds":
		results, err = t.listDaemonSets(ctx, cs, ns, filter, opts)
	case "job", "jobs":
		results, err = t.listJobs(ctx, cs, ns, filter, opts)
	case "cronjob", "cronjobs":
		results, err = t.listCronJobs(ctx, cs, ns, filter, opts)
	case "configmap", "configmaps", "cm":
		results, err = t.listConfigMaps(ctx, cs, ns, filter, opts)
	case "secret", "secrets":
		results, err = t.listSecrets(ctx, cs, ns, filter, opts)
	case "namespace", "namespaces", "ns":
		results, err = t.listNamespaces(ctx, cs, filter, opts)
	case "node", "nodes", "no":
		results, err = t.listNodes(ctx, cs, filter, opts)
	case "ingress", "ingresses", "ing":
		results, err = t.listIngresses(ctx, cs, ns, filter, opts)
	case "event", "events", "ev":
		results, err = t.listEvents(ctx, cs, ns, filter, opts)
	case "helmrelease", "helmreleases", "hr":
		results, err = t.listHelmReleases(ctx, cs, ns, filter)
	case "persistentvolume", "persistentvolumes", "pv":
		results, err = t.listPersistentVolumes(ctx, cs, filter, opts)
	case "persistentvolumeclaim", "persistentvolumeclaims", "pvc":
		results, err = t.listPersistentVolumeClaims(ctx, cs, ns, filter, opts)
	case "serviceaccount", "serviceaccounts", "sa":
		results, err = t.listServiceAccounts(ctx, cs, ns, filter, opts)
	case "endpoint", "endpoints", "ep":
		results, err = t.listEndpoints(ctx, cs, ns, filter, opts)
	case "endpointslice", "endpointslices":
		results, err = t.listEndpointSlices(ctx, cs, ns, filter, opts)
	case "resourcequota", "resourcequotas", "quota":
		results, err = t.listResourceQuotas(ctx, cs, ns, filter, opts)
	case "limitrange", "limitranges", "limits":
		results, err = t.listLimitRanges(ctx, cs, ns, filter, opts)
	case "horizontalpodautoscaler", "horizontalpodautoscalers", "hpa":
		results, err = t.listHorizontalPodAutoscalers(ctx, cs, ns, filter, opts)
	case "poddisruptionbudget", "poddisruptionbudgets", "pdb":
		results, err = t.listPodDisruptionBudgets(ctx, cs, ns, filter, opts)
	case "networkpolicy", "networkpolicies", "netpol":
		results, err = t.listNetworkPolicies(ctx, cs, ns, filter, opts)
	case "storageclass", "storageclasses", "sc":
		results, err = t.listStorageClasses(ctx, cs, filter, opts)
	case "role", "roles":
		results, err = t.listRoles(ctx, cs, ns, filter, opts)
	case "rolebinding", "rolebindings":
		results, err = t.listRoleBindings(ctx, cs, ns, filter, opts)
	case "clusterrole", "clusterroles":
		results, err = t.listClusterRoles(ctx, cs, filter, opts)
	case "clusterrolebinding", "clusterrolebindings":
		results, err = t.listClusterRoleBindings(ctx, cs, filter, opts)
	case "gateway", "gateways", "gw":
		results, err = t.listGateways(ctx, cs, ns, filter, opts)
	case "httproute", "httproutes":
		results, err = t.listHTTPRoutes(ctx, cs, ns, filter, opts)
	default:
		return nil, fmt.Errorf("unsupported resource kind: %s", kind)
	}
	return results, err
}

func (t *ListResourcesTool) listPods(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.CoreV1().Pods(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s (Status: %s)", item.Namespace, item.Name, item.Status.Phase))
	}
	return results, nil
}

func (t *ListResourcesTool) listServices(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.CoreV1().Services(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s (Type: %s, ClusterIP: %s)", item.Namespace, item.Name, item.Spec.Type, item.Spec.ClusterIP))
	}
	return results, nil
}

func (t *ListResourcesTool) listDeployments(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.AppsV1().Deployments(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s (Ready: %d/%d)", item.Namespace, item.Name, item.Status.ReadyReplicas, item.Status.Replicas))
	}
	return results, nil
}

func (t *ListResourcesTool) listReplicaSets(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.AppsV1().ReplicaSets(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s (Replicas: %d)", item.Namespace, item.Name, item.Status.Replicas))
	}
	return results, nil
}

func (t *ListResourcesTool) listStatefulSets(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.AppsV1().StatefulSets(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s (Ready: %d/%d)", item.Namespace, item.Name, item.Status.ReadyReplicas, item.Status.Replicas))
	}
	return results, nil
}

func (t *ListResourcesTool) listDaemonSets(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.AppsV1().DaemonSets(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s (Desired: %d, Ready: %d)", item.Namespace, item.Name, item.Status.DesiredNumberScheduled, item.Status.NumberReady))
	}
	return results, nil
}

func (t *ListResourcesTool) listJobs(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.BatchV1().Jobs(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s (Succeeded: %d)", item.Namespace, item.Name, item.Status.Succeeded))
	}
	return results, nil
}

func (t *ListResourcesTool) listCronJobs(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.BatchV1().CronJobs(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s (Schedule: %s)", item.Namespace, item.Name, item.Spec.Schedule))
	}
	return results, nil
}

func (t *ListResourcesTool) listConfigMaps(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.CoreV1().ConfigMaps(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s", item.Namespace, item.Name))
	}
	return results, nil
}

func (t *ListResourcesTool) listSecrets(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.CoreV1().Secrets(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s (Type: %s)", item.Namespace, item.Name, item.Type))
	}
	return results, nil
}

func (t *ListResourcesTool) listNamespaces(ctx context.Context, cs *cluster.ClientSet, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.CoreV1().Namespaces().List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s (Status: %s)", item.Name, item.Status.Phase))
	}
	return results, nil
}

func (t *ListResourcesTool) listNodes(ctx context.Context, cs *cluster.ClientSet, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.CoreV1().Nodes().List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		status := "NotReady"
		for _, cond := range item.Status.Conditions {
			if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
				status = "Ready"
				break
			}
		}
		results = append(results, fmt.Sprintf("%s (Status: %s, Version: %s)", item.Name, status, item.Status.NodeInfo.KubeletVersion))
	}
	return results, nil
}

func (t *ListResourcesTool) listIngresses(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.NetworkingV1().Ingresses(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s", item.Namespace, item.Name))
	}
	return results, nil
}

func (t *ListResourcesTool) listEvents(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.CoreV1().Events(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.InvolvedObject.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s (%s: %s)", item.Namespace, item.InvolvedObject.Name, item.Type, item.Message))
	}
	return results, nil
}

// --- New Resource Listing Helper Methods ---

func (t *ListResourcesTool) listHelmReleases(ctx context.Context, cs *cluster.ClientSet, ns, filter string) ([]string, error) {
	releases, err := helm.ListReleases(cs.Configuration, ns)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, r := range releases {
		if filter != "" && !strings.Contains(strings.ToLower(r.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s (Chart: %s-%s, Status: %s)",
			r.Namespace, r.Name, r.Chart.Metadata.Name, r.Chart.Metadata.Version, r.Info.Status.String()))
	}
	return results, nil
}

func (t *ListResourcesTool) listPersistentVolumes(ctx context.Context, cs *cluster.ClientSet, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.CoreV1().PersistentVolumes().List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		capacity := item.Spec.Capacity.Storage().String()
		results = append(results, fmt.Sprintf("%s (Capacity: %s, Status: %s, StorageClass: %s)",
			item.Name, capacity, item.Status.Phase, item.Spec.StorageClassName))
	}
	return results, nil
}

func (t *ListResourcesTool) listPersistentVolumeClaims(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.CoreV1().PersistentVolumeClaims(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		capacity := ""
		if item.Spec.Resources.Requests != nil {
			if storage, ok := item.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
				capacity = storage.String()
			}
		}
		storageClass := ""
		if item.Spec.StorageClassName != nil {
			storageClass = *item.Spec.StorageClassName
		}
		results = append(results, fmt.Sprintf("%s/%s (Capacity: %s, Status: %s, StorageClass: %s)",
			item.Namespace, item.Name, capacity, item.Status.Phase, storageClass))
	}
	return results, nil
}

func (t *ListResourcesTool) listServiceAccounts(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.CoreV1().ServiceAccounts(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s (Secrets: %d)", item.Namespace, item.Name, len(item.Secrets)))
	}
	return results, nil
}

func (t *ListResourcesTool) listEndpoints(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.CoreV1().Endpoints(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		addressCount := 0
		for _, subset := range item.Subsets {
			addressCount += len(subset.Addresses)
		}
		results = append(results, fmt.Sprintf("%s/%s (Addresses: %d)", item.Namespace, item.Name, addressCount))
	}
	return results, nil
}

func (t *ListResourcesTool) listEndpointSlices(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.DiscoveryV1().EndpointSlices(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		addressType := string(item.AddressType)
		results = append(results, fmt.Sprintf("%s/%s (AddressType: %s, Endpoints: %d)",
			item.Namespace, item.Name, addressType, len(item.Endpoints)))
	}
	return results, nil
}

func (t *ListResourcesTool) listResourceQuotas(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.CoreV1().ResourceQuotas(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s", item.Namespace, item.Name))
	}
	return results, nil
}

func (t *ListResourcesTool) listLimitRanges(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.CoreV1().LimitRanges(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s", item.Namespace, item.Name))
	}
	return results, nil
}

func (t *ListResourcesTool) listHorizontalPodAutoscalers(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.AutoscalingV2().HorizontalPodAutoscalers(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		minReplicas := int32(1)
		if item.Spec.MinReplicas != nil {
			minReplicas = *item.Spec.MinReplicas
		}
		results = append(results, fmt.Sprintf("%s/%s (Target: %s, Min: %d, Max: %d, Current: %d)",
			item.Namespace, item.Name, item.Spec.ScaleTargetRef.Name, minReplicas, item.Spec.MaxReplicas, item.Status.CurrentReplicas))
	}
	return results, nil
}

func (t *ListResourcesTool) listPodDisruptionBudgets(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.PolicyV1().PodDisruptionBudgets(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		minAvailable := "N/A"
		maxUnavailable := "N/A"
		if item.Spec.MinAvailable != nil {
			minAvailable = item.Spec.MinAvailable.String()
		}
		if item.Spec.MaxUnavailable != nil {
			maxUnavailable = item.Spec.MaxUnavailable.String()
		}
		results = append(results, fmt.Sprintf("%s/%s (MinAvailable: %s, MaxUnavailable: %s)",
			item.Namespace, item.Name, minAvailable, maxUnavailable))
	}
	return results, nil
}

func (t *ListResourcesTool) listNetworkPolicies(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.NetworkingV1().NetworkPolicies(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		policyTypes := ""
		for i, pt := range item.Spec.PolicyTypes {
			if i > 0 {
				policyTypes += ","
			}
			policyTypes += string(pt)
		}
		results = append(results, fmt.Sprintf("%s/%s (PolicyTypes: %s)", item.Namespace, item.Name, policyTypes))
	}
	return results, nil
}

func (t *ListResourcesTool) listStorageClasses(ctx context.Context, cs *cluster.ClientSet, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.StorageV1().StorageClasses().List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		isDefault := ""
		if item.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			isDefault = " (default)"
		}
		results = append(results, fmt.Sprintf("%s (Provisioner: %s)%s", item.Name, item.Provisioner, isDefault))
	}
	return results, nil
}

func (t *ListResourcesTool) listRoles(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.RbacV1().Roles(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s (Rules: %d)", item.Namespace, item.Name, len(item.Rules)))
	}
	return results, nil
}

func (t *ListResourcesTool) listRoleBindings(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.RbacV1().RoleBindings(ns).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s (RoleRef: %s, Subjects: %d)",
			item.Namespace, item.Name, item.RoleRef.Name, len(item.Subjects)))
	}
	return results, nil
}

func (t *ListResourcesTool) listClusterRoles(ctx context.Context, cs *cluster.ClientSet, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.RbacV1().ClusterRoles().List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s (Rules: %d)", item.Name, len(item.Rules)))
	}
	return results, nil
}

func (t *ListResourcesTool) listClusterRoleBindings(ctx context.Context, cs *cluster.ClientSet, filter string, opts metav1.ListOptions) ([]string, error) {
	list, err := cs.K8sClient.ClientSet.RbacV1().ClusterRoleBindings().List(ctx, opts)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		results = append(results, fmt.Sprintf("%s (RoleRef: %s, Subjects: %d)", item.Name, item.RoleRef.Name, len(item.Subjects)))
	}
	return results, nil
}

func (t *ListResourcesTool) listGateways(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	var list gatewayapiv1.GatewayList
	if err := cs.K8sClient.List(ctx, &list); err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if ns != "" && item.Namespace != ns {
			continue
		}
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		gatewayClassName := string(item.Spec.GatewayClassName)
		results = append(results, fmt.Sprintf("%s/%s (GatewayClass: %s, Listeners: %d)",
			item.Namespace, item.Name, gatewayClassName, len(item.Spec.Listeners)))
	}
	return results, nil
}

func (t *ListResourcesTool) listHTTPRoutes(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	var list gatewayapiv1.HTTPRouteList
	if err := cs.K8sClient.List(ctx, &list); err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if ns != "" && item.Namespace != ns {
			continue
		}
		if filter != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(filter)) {
			continue
		}
		parentRefs := len(item.Spec.ParentRefs)
		results = append(results, fmt.Sprintf("%s/%s (ParentRefs: %d, Rules: %d)",
			item.Namespace, item.Name, parentRefs, len(item.Spec.Rules)))
	}
	return results, nil
}

// --- Get Cluster Info Tool ---

type GetClusterInfoTool struct{}

func (t *GetClusterInfoTool) Name() string { return "get_cluster_info" }

func (t *GetClusterInfoTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "get_cluster_info",
			Description: "Get general information about the Kubernetes cluster, including server version and capacity (nodes, CPU, memory).",
			Parameters:  json.RawMessage(`{"type": "object", "properties": {}}`),
		},
	}
}

func (t *GetClusterInfoTool) Execute(ctx context.Context, args string) (string, error) {
	cs, err := GetClientSet(ctx)
	if err != nil {
		return "", err
	}

	// 1. Get Version
	version, err := cs.K8sClient.ClientSet.Discovery().ServerVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get cluster version: %w", err)
	}

	// 2. Get Nodes for detailed info
	nodes, err := cs.K8sClient.ClientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list nodes: %w", err)
	}

	var totalCPU int64
	var totalMem int64
	var controlPlaneCount int
	var workerCount int
	var readyCount int
	var notReadyCount int
	var platform string

	for _, node := range nodes.Items {
		totalCPU += node.Status.Capacity.Cpu().MilliValue()
		totalMem += node.Status.Capacity.Memory().Value()

		// Check if control plane node (either label indicates control plane)
		_, isControlPlane := node.Labels["node-role.kubernetes.io/control-plane"]
		_, isMaster := node.Labels["node-role.kubernetes.io/master"]
		if isControlPlane || isMaster {
			controlPlaneCount++
		} else {
			workerCount++
		}

		// Check node readiness
		isReady := false
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
				isReady = true
				break
			}
		}
		if isReady {
			readyCount++
		} else {
			notReadyCount++
		}

		// Capture platform info from first node
		if platform == "" {
			platform = fmt.Sprintf("%s/%s", node.Status.NodeInfo.OperatingSystem, node.Status.NodeInfo.Architecture)
		}
	}

	// 3. Get Namespace count
	namespaces, err := cs.K8sClient.ClientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	namespaceCount := 0
	if err == nil {
		namespaceCount = len(namespaces.Items)
	}

	// 4. Get Pod summary
	pods, err := cs.K8sClient.ClientSet.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	podStats := make(map[string]int)
	totalPods := 0
	if err == nil {
		totalPods = len(pods.Items)
		for _, pod := range pods.Items {
			podStats[string(pod.Status.Phase)]++
		}
	}

	// Build the info string
	var sb strings.Builder
	sb.WriteString("Cluster Information:\n")
	sb.WriteString(fmt.Sprintf("- Kubernetes Version: %s\n", version.GitVersion))
	sb.WriteString(fmt.Sprintf("- Platform: %s\n", platform))
	sb.WriteString("\nNode Summary:\n")
	sb.WriteString(fmt.Sprintf("- Total Nodes: %d\n", len(nodes.Items)))
	sb.WriteString(fmt.Sprintf("- Control Plane Nodes: %d\n", controlPlaneCount))
	sb.WriteString(fmt.Sprintf("- Worker Nodes: %d\n", workerCount))
	sb.WriteString(fmt.Sprintf("- Ready Nodes: %d\n", readyCount))
	sb.WriteString(fmt.Sprintf("- Not Ready Nodes: %d\n", notReadyCount))
	sb.WriteString("\nCapacity:\n")
	sb.WriteString(fmt.Sprintf("- Total CPU: %dm\n", totalCPU))
	sb.WriteString(fmt.Sprintf("- Total Memory: %d MiB\n", totalMem/(1024*1024)))
	sb.WriteString(fmt.Sprintf("\nNamespaces: %d\n", namespaceCount))
	sb.WriteString(fmt.Sprintf("\nPod Summary (Total: %d):\n", totalPods))
	for phase, count := range podStats {
		sb.WriteString(fmt.Sprintf("- %s: %d\n", phase, count))
	}

	return sb.String(), nil
}
