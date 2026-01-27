package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"strings"

	"github.com/pixelvide/cloud-sentinel-k8s/pkg/cluster"
	openai "github.com/sashabaranov/go-openai"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DebugAppConnectionTool struct{}

func (t *DebugAppConnectionTool) Name() string { return "debug_app_connection" }

func (t *DebugAppConnectionTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "debug_app_connection",
			Description: "Diagnose connection issues for a URL or Service. Traces the full path from Ingress -> Service -> Pods and automatically checks logs/events for errors.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"query": {
						"type": "string",
						"description": "The URL (e.g., 'https://api.example.com/v1') or Service Name (e.g., 'frontend') to debug."
					},
					"namespace": {
						"type": "string",
						"description": "Optional namespace context. Use if query is just a name."
					}
				},
				"required": ["query"]
			}`),
		},
	}
}

func (t *DebugAppConnectionTool) Execute(ctx context.Context, args string) (string, error) {
	var params struct {
		Query     string `json:"query"`
		Namespace string `json:"namespace"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}

	cs, err := GetClientSet(ctx)
	if err != nil {
		return "", err
	}

	report := strings.Builder{}
	fmt.Fprintf(&report, "## 🔍 Debug Report for '%s'\n\n", params.Query)

	// 1. Determine Entry Point
	targetService, targetNamespace, targetPort, err := t.identifyEntryPoint(ctx, cs, params.Query, params.Namespace, &report)
	if err != nil {
		return report.String(), nil //nolint:nilerr
	}

	// 2. Check Service
	svcObj, err := t.checkService(ctx, cs, targetNamespace, targetService, targetPort, &report)
	if err != nil {
		return report.String(), nil //nolint:nilerr
	}

	// 3. Check Endpoints
	t.checkEndpoints(ctx, cs, targetNamespace, targetService, &report)

	// 4. Trace to Pods
	t.checkWorkload(ctx, cs, targetNamespace, svcObj, &report)

	return report.String(), nil
}

func (t *DebugAppConnectionTool) identifyEntryPoint(ctx context.Context, cs *cluster.ClientSet, query, namespace string, report *strings.Builder) (string, string, int32, error) {
	// Heuristic: If missing scheme but has dots, assume web URL
	if !strings.Contains(query, "://") && strings.Contains(query, ".") {
		query = "https://" + query
	}

	u, err := url.Parse(query)
	isURL := err == nil && u.Scheme != "" && u.Host != ""

	if isURL {
		report.WriteString("### 1️⃣ Route Resolution (Ingress)\n")
		ing, svc, port, err := t.resolveIngress(ctx, cs, u)
		if err != nil {
			fmt.Fprintf(report, "❌ **Failed to resolve route**: %v\n", err)
			return "", "", 0, err
		}
		if ing == nil {
			report.WriteString("❌ No Ingress found matching this host/path.\n")
			return "", "", 0, fmt.Errorf("no ingress found")
		}

		fmt.Fprintf(report, "✅ **Matched Ingress**: `%s/%s`\n", ing.Namespace, ing.Name)
		fmt.Fprintf(report, "   - Host: `%s`\n", u.Host)
		fmt.Fprintf(report, "   - Path: `%s`\n", u.Path)
		fmt.Fprintf(report, "   - Backend: Service `%s`:%d\n\n", svc, port)

		return svc, ing.Namespace, port, nil
	}

	// Assume Service Name
	targetService := query
	targetNamespace := namespace
	if targetNamespace == "" {
		targetNamespace = "default"
	}
	fmt.Fprintf(report, "### 1️⃣ Entry Point: Direct Service `%s/%s`\n", targetNamespace, targetService)
	return targetService, targetNamespace, 0, nil
}

func (t *DebugAppConnectionTool) checkService(ctx context.Context, cs *cluster.ClientSet, ns, svcName string, targetPort int32, report *strings.Builder) (*corev1.Service, error) {
	fmt.Fprintf(report, "### 2️⃣ Service Check: `%s`\n", svcName)
	svcObj, err := cs.K8sClient.ClientSet.CoreV1().Services(ns).Get(ctx, svcName, metav1.GetOptions{})
	if err != nil {
		fmt.Fprintf(report, "❌ **Service Not Found**: %v\n", err)
		return nil, err
	}
	fmt.Fprintf(report, "✅ Service exists (Type: %s, ClusterIP: %s)\n", svcObj.Spec.Type, svcObj.Spec.ClusterIP)

	// Verify Port
	portFound := false
	for _, p := range svcObj.Spec.Ports {
		if targetPort != 0 {
			if p.Port == targetPort || (p.Name != "" && p.Name == fmt.Sprintf("%d", targetPort)) {
				portFound = true
				break
			}
		} else {
			portFound = true
			targetPort = p.Port
			break
		}
	}

	if !portFound && targetPort != 0 {
		fmt.Fprintf(report, "⚠️ **Warning**: Ingress targets port %d, but Service exposes: ", targetPort)
		for _, p := range svcObj.Spec.Ports {
			fmt.Fprintf(report, "%d ", p.Port)
		}
		report.WriteString("\n")
	}
	return svcObj, nil
}

func (t *DebugAppConnectionTool) checkEndpoints(ctx context.Context, cs *cluster.ClientSet, ns, svcName string, report *strings.Builder) {
	report.WriteString("\n### 3️⃣ Wiring Check (Endpoints)\n")
	epObj, err := cs.K8sClient.ClientSet.CoreV1().Endpoints(ns).Get(ctx, svcName, metav1.GetOptions{})
	if err != nil {
		fmt.Fprintf(report, "❌ **Failed to get Endpoints**: %v\n", err)
		return
	}

	subsetCount := 0
	readyAddrCount := 0
	for _, s := range epObj.Subsets {
		subsetCount++
		readyAddrCount += len(s.Addresses)
	}
	if readyAddrCount > 0 {
		fmt.Fprintf(report, "✅ Found **%d ready endpoints** backing this service.\n", readyAddrCount)
	} else {
		report.WriteString("❌ **No Ready Endpoints found**. The Service is valid, but no pods are ready to accept traffic.\n")
	}
}

func (t *DebugAppConnectionTool) checkWorkload(ctx context.Context, cs *cluster.ClientSet, ns string, svcObj *corev1.Service, report *strings.Builder) {
	report.WriteString("\n### 4️⃣ Workload Check (Pods)\n")
	if len(svcObj.Spec.Selector) == 0 {
		report.WriteString("⚠️ Service has no selectors (ExternalName or manual endpoints?)\n")
		return
	}

	selectorStr := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: svcObj.Spec.Selector})
	pods, err := cs.K8sClient.ClientSet.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{LabelSelector: selectorStr})
	if err != nil {
		fmt.Fprintf(report, "❌ Failed to list pods: %v\n", err)
		return
	}

	if len(pods.Items) == 0 {
		fmt.Fprintf(report, "❌ **No Pods found** matching selector `%s`.\n", selectorStr)
		return
	}

	fmt.Fprintf(report, "Found %d Pods:\n", len(pods.Items))
	for _, pod := range pods.Items {
		fmt.Fprintf(report, "- **%s** (Status: `%s`, Restarts: %d)\n", pod.Name, pod.Status.Phase, containerRestarts(pod))

		if pod.Status.Phase != corev1.PodRunning || containerRestarts(pod) > 0 {
			t.diagnosePod(ctx, cs, ns, pod, report)
		}
	}
}

func (t *DebugAppConnectionTool) diagnosePod(ctx context.Context, cs *cluster.ClientSet, ns string, pod corev1.Pod, report *strings.Builder) {
	// Get Events
	events, _ := cs.K8sClient.ClientSet.CoreV1().Events(ns).List(ctx, metav1.ListOptions{FieldSelector: fmt.Sprintf("involvedObject.name=%s", pod.Name)})
	if len(events.Items) > 0 {
		report.WriteString("  - 📢 **Events**:\n")
		for _, e := range events.Items {
			if e.Type == "Warning" {
				fmt.Fprintf(report, "    - [%s] %s: %s\n", e.LastTimestamp.Format("15:04"), e.Reason, e.Message)
			}
		}
	}

	// Get Logs
	if len(pod.Spec.Containers) > 0 {
		containerName := pod.Spec.Containers[0].Name
		fmt.Fprintf(report, "  - 📜 **Logs** (Tail 15, %s):\n", containerName)
		logs, err := t.getPodLogs(ctx, cs, ns, pod.Name, containerName)
		if err == nil && logs != "" {
			report.WriteString("    ```\n")
			report.WriteString(logs)
			report.WriteString("\n    ```\n")
		}
	}
}

// --- Helpers ---

// resolveIngress implements "best match" logic for Ingress
func (t *DebugAppConnectionTool) resolveIngress(ctx context.Context, cs *cluster.ClientSet, u *url.URL) (*networkingv1.Ingress, string, int32, error) {
	ingresses, err := cs.K8sClient.ClientSet.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, "", 0, err
	}

	var bestMatch *networkingv1.Ingress
	var bestPathLen int
	var targetService string
	var targetPort int32

	// Normalize Path
	requestPath := u.Path
	if requestPath == "" {
		requestPath = "/"
	}

	for _, ing := range ingresses.Items {
		for _, rule := range ing.Spec.Rules {
			// 1. Host Match
			if rule.Host != "" && !strings.EqualFold(rule.Host, u.Host) {
				// Handle wildcards? simple exact match for now + ignoring wildcard complexity for MVP
				continue
			}

			// 2. Path Match
			for _, path := range rule.HTTP.Paths {
				// Simple prefix matching
				if strings.HasPrefix(requestPath, path.Path) {
					if len(path.Path) > bestPathLen {
						bestPathLen = len(path.Path)
						bestMatch = &ing
						targetService = path.Backend.Service.Name
						if path.Backend.Service.Port.Number != 0 {
							targetPort = path.Backend.Service.Port.Number
						} else if path.Backend.Service.Port.Name != "" {
							targetPort = 0
						}
					}
				}
			}
		}
	}

	if bestMatch != nil {
		// Return copy to avoid loop variable issues (though Go 1.22 fixed this)
		match := *bestMatch
		return &match, targetService, targetPort, nil
	}

	return nil, "", 0, nil
}

func containerRestarts(pod corev1.Pod) int32 {
	var restarts int32
	for _, cs := range pod.Status.ContainerStatuses {
		restarts += cs.RestartCount
	}
	return restarts
}

func (t *DebugAppConnectionTool) getPodLogs(ctx context.Context, cs *cluster.ClientSet, ns, name, container string) (string, error) {
	tail := int64(15)
	opts := &corev1.PodLogOptions{
		Container: container,
		TailLines: &tail,
	}
	req := cs.K8sClient.ClientSet.CoreV1().Pods(ns).GetLogs(name, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = stream.Close()
	}()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, stream)
	return strings.TrimSpace(buf.String()), err
}
