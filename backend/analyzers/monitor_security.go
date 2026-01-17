package analyzers

import (
	"cloud-sentinel-k8s/pkg/models"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// ImmutableTagAnalyzer detects usage of 'latest' tag or missing tags
type ImmutableTagAnalyzer struct{}

func (i *ImmutableTagAnalyzer) Name() string { return "ImmutableTags" }

func (i *ImmutableTagAnalyzer) Analyze(obj *unstructured.Unstructured, client dynamic.Interface, clusterID string) []models.Anomaly {
	kind := obj.GetKind()
	supportedKinds := map[string]bool{
		"Deployment":  true,
		"StatefulSet": true,
		"DaemonSet":   true,
		"Pod":         true,
	}

	if !supportedKinds[kind] {
		return nil
	}

	// Helper to extract containers from different resource types
	var containers []interface{}
	var initContainers []interface{}
	var foundC, foundIC bool
	var errC, errIC error

	if kind == "Pod" {
		containers, foundC, errC = unstructured.NestedSlice(obj.Object, "spec", "containers")
		initContainers, foundIC, errIC = unstructured.NestedSlice(obj.Object, "spec", "initContainers")
	} else {
		// Workloads (Deployment, StatefulSet, DaemonSet) store containers in spec.template.spec.containers
		containers, foundC, errC = unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "containers")
		initContainers, foundIC, errIC = unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "initContainers")
	}

	if (errC != nil || !foundC) && (errIC != nil || !foundIC) {
		return nil
	}

	var anomalies []models.Anomaly
	allContainers := append(containers, initContainers...)

	for _, c := range allContainers {
		container, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		name, _, _ := unstructured.NestedString(container, "name")
		image, found, _ := unstructured.NestedString(container, "image")

		if found {
			// Check if tag is 'latest' or missing (which implies latest)
			// Examples: "nginx", "nginx:latest", "my-registry.io/img:latest"
			isLatest := false
			if strings.HasSuffix(image, ":latest") {
				isLatest = true
			} else if !strings.Contains(image, ":") {
				// No tag specified usually means latest, unless it has a digest (sha256:...)
				// "ubuntu@sha256:..." is valid and immutable.
				if !strings.Contains(image, "@") {
					isLatest = true
				}
			}

			if isLatest {
				anomalies = append(anomalies, NewAnomaly(
					i.Name(),
					models.SeverityWarning,
					"Mutable Image Tag Detected",
					fmt.Sprintf("Container '%s' is using a mutable image tag: '%s'.", name, image),
					"Use a specific version tag (e.g., :v1.0.0) or digest (@sha256:...) to ensure immutability and reproducible deployments.",
				))
			}
		}
	}

	return anomalies
}

// PrivilegedContainerAnalyzer detects containers running with privileged: true
type PrivilegedContainerAnalyzer struct{}

func (p *PrivilegedContainerAnalyzer) Name() string { return "PrivilegedContainer" }

func (p *PrivilegedContainerAnalyzer) Analyze(obj *unstructured.Unstructured, client dynamic.Interface, clusterID string) []models.Anomaly {
	kind := obj.GetKind()
	supportedKinds := map[string]bool{
		"Deployment":  true,
		"StatefulSet": true,
		"DaemonSet":   true,
		"Pod":         true,
	}

	if !supportedKinds[kind] {
		return nil
	}

	// Helper to extract containers from different resource types
	var containers []interface{}
	var initContainers []interface{}
	var foundC, foundIC bool
	var errC, errIC error

	if kind == "Pod" {
		containers, foundC, errC = unstructured.NestedSlice(obj.Object, "spec", "containers")
		initContainers, foundIC, errIC = unstructured.NestedSlice(obj.Object, "spec", "initContainers")
	} else {
		// Workloads (Deployment, StatefulSet, DaemonSet) store containers in spec.template.spec.containers
		containers, foundC, errC = unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "containers")
		initContainers, foundIC, errIC = unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "initContainers")
	}

	if (errC != nil || !foundC) && (errIC != nil || !foundIC) {
		return nil
	}

	var anomalies []models.Anomaly
	allContainers := append(containers, initContainers...)

	for _, c := range allContainers {
		container, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		name, _, _ := unstructured.NestedString(container, "name")
		privileged, found, _ := unstructured.NestedBool(container, "securityContext", "privileged")

		if found && privileged {
			anomalies = append(anomalies, NewAnomaly(
				p.Name(),
				models.SeverityWarning,
				"Privileged Container Detected",
				fmt.Sprintf("Container '%s' is running in privileged mode.", name),
				"Avoid running containers as privileged unless absolutely necessary. Grant specific capabilities instead.",
			))
		}
	}

	return anomalies
}

// RootUserAnalyzer detects containers running as root (UID 0)
type RootUserAnalyzer struct{}

func (r *RootUserAnalyzer) Name() string { return "RootUser" }

func (r *RootUserAnalyzer) Analyze(obj *unstructured.Unstructured, client dynamic.Interface, clusterID string) []models.Anomaly {
	kind := obj.GetKind()
	supportedKinds := map[string]bool{
		"Deployment":  true,
		"StatefulSet": true,
		"DaemonSet":   true,
		"Pod":         true,
	}

	if !supportedKinds[kind] {
		return nil
	}

	// Helper to extract containers from different resource types
	var containers []interface{}
	var initContainers []interface{}
	var foundC, foundIC bool
	var errC, errIC error

	if kind == "Pod" {
		containers, foundC, errC = unstructured.NestedSlice(obj.Object, "spec", "containers")
		initContainers, foundIC, errIC = unstructured.NestedSlice(obj.Object, "spec", "initContainers")
	} else {
		// Workloads (Deployment, StatefulSet, DaemonSet) store containers in spec.template.spec.containers
		containers, foundC, errC = unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "containers")
		initContainers, foundIC, errIC = unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "initContainers")
	}

	if (errC != nil || !foundC) && (errIC != nil || !foundIC) {
		return nil
	}

	var anomalies []models.Anomaly
	allContainers := append(containers, initContainers...)

	// Check PodSecurityContext first for runAsNonRoot or runAsUser
	var podSecurityContext map[string]interface{}
	if kind == "Pod" {
		podSecurityContext, _, _ = unstructured.NestedMap(obj.Object, "spec", "securityContext")
	} else {
		podSecurityContext, _, _ = unstructured.NestedMap(obj.Object, "spec", "template", "spec", "securityContext")
	}

	podRunAsNonRoot := false
	podRunAsUser := int64(-1)
	if podSecurityContext != nil {
		if val, found, _ := unstructured.NestedBool(podSecurityContext, "runAsNonRoot"); found && val {
			podRunAsNonRoot = true
		}
		if val, found, _ := unstructured.NestedInt64(podSecurityContext, "runAsUser"); found {
			podRunAsUser = val
		}
	}

	for _, c := range allContainers {
		container, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		name, _, _ := unstructured.NestedString(container, "name")

		// Container Security Context overrides Pod Security Context
		securityContext, foundSC, _ := unstructured.NestedMap(container, "securityContext")

		runAsNonRoot := podRunAsNonRoot
		runAsUser := podRunAsUser

		if foundSC {
			if val, found, _ := unstructured.NestedBool(securityContext, "runAsNonRoot"); found {
				runAsNonRoot = val
			}
			if val, found, _ := unstructured.NestedInt64(securityContext, "runAsUser"); found {
				runAsUser = val
			}
		}

		// Analysis Logic:
		// 1. If runAsUser is 0 -> Root
		// 2. If runAsUser is NOT set, and runAsNonRoot is false (or unset) -> Potentially Root (default in many images)
		// We will only flag EXPLICIT root or MISSING non-root enforcement.
		// Actually, standard practice is to flag if it's NOT explicitly running as non-root.

		isRoot := false
		if runAsUser == 0 {
			isRoot = true
		} else if runAsUser == -1 && !runAsNonRoot {
			// If no user specified and not enforced non-root, it acts as root (UID 0) in standard Docker images.
			isRoot = true
		}

		if isRoot {
			msg := fmt.Sprintf("Container '%s' may be running as root.", name)
			if runAsUser == 0 {
				msg = fmt.Sprintf("Container '%s' is explicitly configured to run as root (UID 0).", name)
			}

			anomalies = append(anomalies, NewAnomaly(
				r.Name(),
				models.SeverityWarning,
				"Container Running as Root",
				msg,
				"Configure 'securityContext.runAsUser' to a non-zero ID or set 'securityContext.runAsNonRoot: true' to improve security isolation.",
			))
		}
	}

	return anomalies
}

func init() {
	GlobalAnalyzers = append(GlobalAnalyzers, &ImmutableTagAnalyzer{})
	GlobalAnalyzers = append(GlobalAnalyzers, &PrivilegedContainerAnalyzer{})
	GlobalAnalyzers = append(GlobalAnalyzers, &RootUserAnalyzer{})
	GlobalAnalyzers = append(GlobalAnalyzers, &HostPathAnalyzer{})
}

// HostPathAnalyzer detects usage of hostPath volumes
type HostPathAnalyzer struct{}

func (h *HostPathAnalyzer) Name() string { return "HostPathVolume" }

func (h *HostPathAnalyzer) Analyze(obj *unstructured.Unstructured, client dynamic.Interface, clusterID string) []models.Anomaly {
	kind := obj.GetKind()
	supportedKinds := map[string]bool{
		"Deployment":  true,
		"StatefulSet": true,
		"DaemonSet":   true,
		"Pod":         true,
	}

	if !supportedKinds[kind] {
		return nil
	}

	// Helper to extract volumes
	var volumes []interface{}
	var found bool
	var err error

	if kind == "Pod" {
		volumes, found, err = unstructured.NestedSlice(obj.Object, "spec", "volumes")
	} else {
		volumes, found, err = unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "volumes")
	}

	if err != nil || !found {
		return nil
	}

	var anomalies []models.Anomaly

	for _, v := range volumes {
		vol, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		name, _, _ := unstructured.NestedString(vol, "name")
		_, hasHostPath, _ := unstructured.NestedMap(vol, "hostPath")

		if hasHostPath {
			anomalies = append(anomalies, NewAnomaly(
				h.Name(),
				models.SeverityWarning,
				"HostPath Volume Detected",
				fmt.Sprintf("Volume '%s' is using hostPath.", name),
				"Avoid using hostPath volumes as they present security risks and limit pod portability. Use PVCs or emptyDir instead.",
			))
		}
	}

	return anomalies
}
