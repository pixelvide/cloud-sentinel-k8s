package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pixelvide/cloud-sentinel-k8s/pkg/cluster"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/helm"
	openai "github.com/sashabaranov/go-openai"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// --- List Resources Tool ---

type ListResourcesTool struct{}

func (t *ListResourcesTool) Name() string { return "list_resources" }

func (t *ListResourcesTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "list_resources",
			Description: "List Kubernetes resources of a specific kind in a namespace, optionally filtered by labels. This tool also support listing CustomResourceDefinition (CRD) and HelmRelease.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"kind": {
						"type": "string",
						"description": "The kind of resource to list (Pod, Service, Deployment, ReplicaSet, StatefulSet, DaemonSet, Job, CronJob, ConfigMap, Secret, Namespace, Node, Ingress, Event, CustomResourceDefinition, HelmRelease)."
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

	// Summarize results if they are too many
	summary := ""
	if len(results) > 50 {
		summary = fmt.Sprintf("Found %d %s. Showing first 50 results:\n\n", len(results), params.Kind)
		results = results[:50]
		results = append(results, fmt.Sprintf("\n... and %d more resources. Use more specific filters or a name filter to narrow down.", len(results)-50))
	} else {
		summary = fmt.Sprintf("Found %d %s:\n\n", len(results), params.Kind)
	}

	return summary + strings.Join(results, "\n"), nil
}

// Define a common function signature that fits the 'superset' of arguments
type listFunc func(tx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error)

func (t *ListResourcesTool) getListHandler(kind string) listFunc {
	handlers := map[string]listFunc{
		// pods
		"pod":  t.listPods,
		"pods": t.listPods,

		// services
		"service":  t.listServices,
		"services": t.listServices,
		"svc":      t.listServices,

		// deployments
		"deployment":  t.listDeployments,
		"deployments": t.listDeployments,
		"deploy":      t.listDeployments,

		// replica sets
		"replicaset":  t.listReplicaSets,
		"replicasets": t.listReplicaSets,
		"rs":          t.listReplicaSets,

		// stateful sets
		"statefulset":  t.listStatefulSets,
		"statefulsets": t.listStatefulSets,
		"sts":          t.listStatefulSets,

		// daemon sets
		"daemonset":  t.listDaemonSets,
		"daemonsets": t.listDaemonSets,
		"ds":         t.listDaemonSets,

		// jobs
		"job":  t.listJobs,
		"jobs": t.listJobs,

		// cron jobs
		"cronjob":  t.listCronJobs,
		"cronjobs": t.listCronJobs,
		"cj":       t.listCronJobs,

		// config maps
		"configmap":  t.listConfigMaps,
		"configmaps": t.listConfigMaps,
		"cm":         t.listConfigMaps,

		// secrets
		"secret":  t.listSecrets,
		"secrets": t.listSecrets,
		"sec":     t.listSecrets,

		// ingress
		"ingress":   t.listIngresses,
		"ingresses": t.listIngresses,
		"ing":       t.listIngresses,

		// events
		"event":  t.listEvents,
		"events": t.listEvents,
		"ev":     t.listEvents,

		// nodes
		"node":  t.listNodes,
		"nodes": t.listNodes,
		"no":    t.listNodes,

		// namespaces
		"namespace":  t.listNamespaces,
		"namespaces": t.listNamespaces,
		"ns":         t.listNamespaces,

		// persistent volumes
		"persistentvolume":  t.listPersistentVolumes,
		"persistentvolumes": t.listPersistentVolumes,
		"pv":                t.listPersistentVolumes,

		// persistent volume claims
		"persistentvolumeclaim":  t.listPersistentVolumeClaims,
		"persistentvolumeclaims": t.listPersistentVolumeClaims,
		"pvc":                    t.listPersistentVolumeClaims,

		// service accounts
		"serviceaccount":  t.listServiceAccounts,
		"serviceaccounts": t.listServiceAccounts,
		"sa":              t.listServiceAccounts,

		// endpoints
		"endpoint":  t.listEndpoints,
		"endpoints": t.listEndpoints,
		"ep":        t.listEndpoints,

		// endpointslices
		"endpointslice":  t.listEndpointSlices,
		"endpointslices": t.listEndpointSlices,

		// resource quotas
		"resourcequota":  t.listResourceQuotas,
		"resourcequotas": t.listResourceQuotas,
		"quota":          t.listResourceQuotas,

		// limit ranges
		"limitrange":  t.listLimitRanges,
		"limitranges": t.listLimitRanges,
		"limits":      t.listLimitRanges,

		// storage classes
		"storageclass":   t.listStorageClasses,
		"storageclasses": t.listStorageClasses,
		"sc":             t.listStorageClasses,

		// cluster roles
		"clusterrole":  t.listClusterRoles,
		"clusterroles": t.listClusterRoles,
		"cr":           t.listClusterRoles,

		// cluster role bindings
		"clusterrolebinding":  t.listClusterRoleBindings,
		"clusterrolebindings": t.listClusterRoleBindings,
		"crb":                 t.listClusterRoleBindings,

		// roles
		"role":  t.listRoles,
		"roles": t.listRoles,

		// role bindings
		"rolebinding":  t.listRoleBindings,
		"rolebindings": t.listRoleBindings,

		// gateway
		"gateway":  t.listGateways,
		"gateways": t.listGateways,
		"gw":       t.listGateways,

		// http route
		"httproute":  t.listHTTPRoutes,
		"httproutes": t.listHTTPRoutes,

		// horizontal pod autoscaler
		"horizontalpodautoscaler":  t.listHorizontalPodAutoscalers,
		"horizontalpodautoscalers": t.listHorizontalPodAutoscalers,
		"hpa":                      t.listHorizontalPodAutoscalers,

		// pod disruption budget
		"poddisruptionbudget":  t.listPodDisruptionBudgets,
		"poddisruptionbudgets": t.listPodDisruptionBudgets,
		"pdb":                  t.listPodDisruptionBudgets,

		// network policy
		"networkpolicy":   t.listNetworkPolicies,
		"networkpolicies": t.listNetworkPolicies,
		"netpol":          t.listNetworkPolicies,

		// helm release
		"helmrelease":  t.listHelmReleases,
		"helmreleases": t.listHelmReleases,
		"hr":           t.listHelmReleases,

		// custom resource definitions
		"customresourcedefinition":  t.listCRDs,
		"customresourcedefinitions": t.listCRDs,
		"crd":                       t.listCRDs,
		"crds":                      t.listCRDs,
	}

	return handlers[kind]
}

func (t *ListResourcesTool) listByKind(ctx context.Context, cs *cluster.ClientSet, kind, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	var results []string
	var err error

	// 1. Look up the handler
	handler := t.getListHandler(kind)
	if handler != nil {
		// 2. Call the handler
		results, err = handler(ctx, cs, ns, filter, opts)
	} else {
		// 3. Try dynamic listing
		results, err = t.listDynamicResources(ctx, cs, kind, ns, filter, opts)
	}

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (t *ListResourcesTool) listPods(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	var list corev1.PodList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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
	var list corev1.ServiceList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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
	var list appsv1.DeploymentList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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
	var list appsv1.ReplicaSetList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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
	var list appsv1.StatefulSetList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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
	var list appsv1.DaemonSetList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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
	var list batchv1.JobList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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
	var list batchv1.CronJobList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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
	var list corev1.ConfigMapList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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
	var list corev1.SecretList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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

func (t *ListResourcesTool) listNamespaces(ctx context.Context, cs *cluster.ClientSet, _ string, filter string, opts metav1.ListOptions) ([]string, error) {
	var list corev1.NamespaceList
	if err := listK8sObject(ctx, cs, "", opts, &list); err != nil {
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

func (t *ListResourcesTool) listNodes(ctx context.Context, cs *cluster.ClientSet, _ string, filter string, opts metav1.ListOptions) ([]string, error) {
	var list corev1.NodeList
	if err := listK8sObject(ctx, cs, "", opts, &list); err != nil {
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
	var list networkingv1.IngressList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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
	var list corev1.EventList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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

func (t *ListResourcesTool) listHelmReleases(_ context.Context, cs *cluster.ClientSet, ns, filter string, _ metav1.ListOptions) ([]string, error) {
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

func (t *ListResourcesTool) listPersistentVolumes(ctx context.Context, cs *cluster.ClientSet, _ string, filter string, opts metav1.ListOptions) ([]string, error) {
	var list corev1.PersistentVolumeList
	if err := listK8sObject(ctx, cs, "", opts, &list); err != nil {
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
	var list corev1.PersistentVolumeClaimList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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
	var list corev1.ServiceAccountList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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
	var list corev1.EndpointsList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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
	var list discoveryv1.EndpointSliceList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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
	var list corev1.ResourceQuotaList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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
	var list corev1.LimitRangeList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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
	var list autoscalingv2.HorizontalPodAutoscalerList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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
	var list policyv1.PodDisruptionBudgetList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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
	var list networkingv1.NetworkPolicyList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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

func (t *ListResourcesTool) listStorageClasses(ctx context.Context, cs *cluster.ClientSet, _ string, filter string, opts metav1.ListOptions) ([]string, error) {
	var list storagev1.StorageClassList
	if err := listK8sObject(ctx, cs, "", opts, &list); err != nil {
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
	var list rbacv1.RoleList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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
	var list rbacv1.RoleBindingList
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
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

func (t *ListResourcesTool) listClusterRoles(ctx context.Context, cs *cluster.ClientSet, _ string, filter string, opts metav1.ListOptions) ([]string, error) {
	var list rbacv1.ClusterRoleList
	if err := listK8sObject(ctx, cs, "", opts, &list); err != nil {
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

func (t *ListResourcesTool) listClusterRoleBindings(ctx context.Context, cs *cluster.ClientSet, _ string, filter string, opts metav1.ListOptions) ([]string, error) {
	var list rbacv1.ClusterRoleBindingList
	if err := listK8sObject(ctx, cs, "", opts, &list); err != nil {
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
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if !shouldIncludeResource(item.Name, item.Namespace, ns, filter) {
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
	if err := listK8sObject(ctx, cs, ns, opts, &list); err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if !shouldIncludeResource(item.Name, item.Namespace, ns, filter) {
			continue
		}

		parentRefs := len(item.Spec.ParentRefs)
		results = append(results, fmt.Sprintf("%s/%s (ParentRefs: %d, Rules: %d)",
			item.Namespace, item.Name, parentRefs, len(item.Spec.Rules)))
	}
	return results, nil
}

func (t *ListResourcesTool) listCRDs(ctx context.Context, cs *cluster.ClientSet, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	// CRDs are cluster-scoped, so we ignore namespace in the request but respect it if provided in options (which shouldn't happen for CRDs usually)
	var list apiextensionsv1.CustomResourceDefinitionList
	if err := listK8sObject(ctx, cs, "", opts, &list); err != nil {
		return nil, err
	}
	var results []string
	for _, item := range list.Items {
		if !shouldIncludeResource(item.Name, "", "", filter) {
			continue
		}

		results = append(results, fmt.Sprintf("%s (Group: %s, Version: %s, Scope: %s)",
			item.Name, item.Spec.Group, item.Spec.Versions[0].Name, item.Spec.Scope))
	}
	return results, nil
}

func (t *ListResourcesTool) listDynamicResources(ctx context.Context, cs *cluster.ClientSet, kind, ns, filter string, opts metav1.ListOptions) ([]string, error) {
	// 1. Verify availability and resolve GVK
	gvk, err := resolveGVK(cs, kind)
	if err != nil {
		return nil, err
	}

	// 2. Build list options
	listUpdates, err := buildListOptions(ns, opts)
	if err != nil {
		return nil, err
	}

	// 3. List resources
	var uList unstructured.UnstructuredList
	uList.SetGroupVersionKind(gvk)

	if err := cs.K8sClient.List(ctx, &uList, listUpdates...); err != nil {
		return nil, err
	}

	var results []string
	for _, item := range uList.Items {
		if !shouldIncludeResource(item.GetName(), item.GetNamespace(), ns, filter) {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s (Kind: %s)", item.GetNamespace(), item.GetName(), item.GetKind()))
	}
	return results, nil
}
