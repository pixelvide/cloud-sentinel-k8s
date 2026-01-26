package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/pixelvide/cloud-sentinel-k8s/pkg/cluster"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func buildListOptions(ns string, opts metav1.ListOptions) ([]client.ListOption, error) {
	var listUpdates []client.ListOption
	if ns != "" {
		listUpdates = append(listUpdates, client.InNamespace(ns))
	}
	if opts.LabelSelector != "" {
		selector, err := labels.Parse(opts.LabelSelector)
		if err != nil {
			return nil, fmt.Errorf("invalid label selector: %w", err)
		}
		listUpdates = append(listUpdates, client.MatchingLabelsSelector{Selector: selector})
	}
	return listUpdates, nil
}

func listK8sObject[L client.ObjectList](ctx context.Context, cs *cluster.ClientSet, ns string, opts metav1.ListOptions, list L) error {
	listUpdates, err := buildListOptions(ns, opts)
	if err != nil {
		return err
	}
	return cs.K8sClient.List(ctx, list, listUpdates...)
}

func shouldIncludeResource(name, itemNs, requestNs, filter string) bool {
	if requestNs != "" && itemNs != requestNs {
		return false
	}
	if filter != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(filter)) {
		return false
	}
	return true
}

func resolveGVK(cs *cluster.ClientSet, kind string) (schema.GroupVersionKind, error) {
	// Discovery
	apiResourceLists, err := cs.K8sClient.ClientSet.Discovery().ServerPreferredResources()
	if err != nil {
		return schema.GroupVersionKind{}, fmt.Errorf("failed to discover resources: %w", err)
	}

	var bestMatch *metav1.APIResource
	var bestMatchGV string

	lowerKind := strings.ToLower(kind)
	found := false

	// Find the resource
	for _, list := range apiResourceLists {
		for _, resource := range list.APIResources {
			if strings.ToLower(resource.Kind) == lowerKind || strings.ToLower(resource.Name) == lowerKind || strings.ToLower(resource.SingularName) == lowerKind || utils.ContainsString(resource.ShortNames, lowerKind) {
				r := resource
				bestMatch = &r
				bestMatchGV = list.GroupVersion
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return schema.GroupVersionKind{}, fmt.Errorf("unsupported resource kind: %s (CRD not found or not available)", kind)
	}

	gv, err := schema.ParseGroupVersion(bestMatchGV)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}

	return gv.WithKind(bestMatch.Kind), nil
}
