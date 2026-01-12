package api

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// HandleLogs streams logs from a pod via WebSocket
func HandleLogs(c *gin.Context) {
	ns := c.Query("namespace")
	podName := c.Query("pod")
	containerName := c.Query("container")
	ctxName := c.Query("context")
	timestampsStr := c.Query("timestamps")
	selectorStr := c.Query("selector")

	if ns == "" || podName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace and pod required"})
		return
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WS Upgrade failed: %v", err)
		return
	}
	defer ws.Close()

	userVal, exists := c.Get("user")
	var storageNamespace string
	if exists {
		user := userVal.(*models.User)
		storageNamespace = user.StorageNamespace
	}

	clientset, _, err := GetClientInfo(storageNamespace, ctxName)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("Config Error: "+err.Error()))
		return
	}

	prefixStr := c.Query("prefix")
	showPrefix := prefixStr == "true"

	showTimestamps := timestampsStr != "false"

	// Handle __all__ for pods - stream from all pods matching selector
	if podName == "__all__" {
		if selectorStr == "" {
			ws.WriteMessage(websocket.TextMessage, []byte("Error: selector required when pod=__all__"))
			return
		}

		// List all pods matching selector
		listOpts := metav1.ListOptions{LabelSelector: selectorStr}
		podList, err := clientset.CoreV1().Pods(ns).List(c.Request.Context(), listOpts)
		if err != nil {
			ws.WriteMessage(websocket.TextMessage, []byte("Error listing pods: "+err.Error()))
			return
		}

		if len(podList.Items) == 0 {
			ws.WriteMessage(websocket.TextMessage, []byte("No pods found matching selector"))
			return
		}

		// Always show prefix for multi-pod logs
		showPrefix = true
		streamMultiplePods(c, ws, clientset, ns, podList.Items, containerName, showTimestamps, showPrefix)
		return
	}

	// Original single-pod logic
	// Parse requested containers
	var targetContainers []string
	if containerName == "__all__" {
		pod, err := clientset.CoreV1().Pods(ns).Get(c.Request.Context(), podName, metav1.GetOptions{})
		if err != nil {
			ws.WriteMessage(websocket.TextMessage, []byte("Error getting pod: "+err.Error()))
			return
		}
		for _, c := range pod.Spec.InitContainers {
			targetContainers = append(targetContainers, c.Name)
		}
		for _, c := range pod.Spec.Containers {
			targetContainers = append(targetContainers, c.Name)
		}
	} else if strings.Contains(containerName, ",") {
		targetContainers = strings.Split(containerName, ",")
	} else {
		// Single container requested
		targetContainers = []string{containerName}
	}

	// Unified streaming logic for both single and multi-container requests
	// Default showPrefix to true if we are streaming multiple sources
	if prefixStr == "" && (len(targetContainers) > 1 || containerName == "__all__") {
		showPrefix = true
	}
	streamContainers(c, ws, clientset, ns, podName, targetContainers, showTimestamps, showPrefix)
}

const (
	pingPeriod = 15 * time.Second
	writeWait  = 10 * time.Second
)

func streamContainers(c *gin.Context, ws *websocket.Conn, clientset *kubernetes.Clientset, ns, podName string, containers []string, showTimestamps bool, showPrefix bool) {
	var wg sync.WaitGroup
	logChan := make(chan string)

	// Start a goroutine for each container
	for _, cName := range containers {
		cName = strings.TrimSpace(cName)
		if cName == "" {
			continue
		}
		wg.Add(1)
		go func(container string) {
			defer wg.Done()
			streamLogsInternal(c, clientset, ns, podName, container, showTimestamps, showPrefix, logChan)
		}(cName)
	}

	// Closer routine
	go func() {
		wg.Wait()
		close(logChan)
	}()

	writeLoop(ws, logChan)
}

// streamMultiplePods streams logs from multiple pods
func streamMultiplePods(c *gin.Context, ws *websocket.Conn, clientset *kubernetes.Clientset, ns string, pods []v1.Pod, containerFilter string, showTimestamps bool, showPrefix bool) {
	var wg sync.WaitGroup
	logChan := make(chan string)

	for _, pod := range pods {
		podName := pod.Name

		// Build container list for this pod
		var containers []string
		if containerFilter == "__all__" {
			for _, c := range pod.Spec.InitContainers {
				containers = append(containers, c.Name)
			}
			for _, c := range pod.Spec.Containers {
				containers = append(containers, c.Name)
			}
		} else if containerFilter != "" {
			// Specific container requested
			containers = []string{containerFilter}
		} else {
			// Default to first container
			if len(pod.Spec.Containers) > 0 {
				containers = []string{pod.Spec.Containers[0].Name}
			}
		}

		// Stream from each container in this pod
		for _, cName := range containers {
			wg.Add(1)
			go func(pName, container string) {
				defer wg.Done()
				streamLogsWithPodPrefix(c, clientset, ns, pName, container, showTimestamps, showPrefix, logChan)
			}(podName, cName)
		}
	}

	// Closer routine
	go func() {
		wg.Wait()
		close(logChan)
	}()

	writeLoop(ws, logChan)
}

// streamLogsWithPodPrefix streams logs with [pod/container] prefix
func streamLogsWithPodPrefix(c *gin.Context, clientset *kubernetes.Clientset, ns, podName, containerName string, showTimestamps bool, showPrefix bool, outChan chan<- string) {
	prefix := ""
	if showPrefix {
		prefix = "[" + podName + "/" + containerName + "] "
	}

	opts := &v1.PodLogOptions{
		Container:  containerName,
		Follow:     true,
		Timestamps: showTimestamps,
		TailLines:  func() *int64 { i := int64(100); return &i }(),
	}

	req := clientset.CoreV1().Pods(ns).GetLogs(podName, opts)
	stream, err := req.Stream(c.Request.Context())
	if err != nil {
		outChan <- prefix + "Error opening stream: " + err.Error() + "\n"
		return
	}
	defer stream.Close()

	reader := bufio.NewReader(stream)
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			outChan <- prefix + string(line)
		}
		if err != nil {
			if err != io.EOF {
				outChan <- prefix + "Stream ended with error: " + err.Error() + "\n"
			}
			break
		}
	}
}

func writeLoop(ws *websocket.Conn, messages <-chan string) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-messages:
			if !ok {
				return
			}
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				return
			}
		case <-ticker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// streamLogsInternal is unique for "all" mode as it writes to a channel instead of WS directly
func streamLogsInternal(c *gin.Context, clientset *kubernetes.Clientset, ns, podName, containerName string, showTimestamps bool, showPrefix bool, outChan chan<- string) {
	prefix := ""
	if showPrefix {
		prefix = "[" + containerName + "] "
	}

	opts := &v1.PodLogOptions{
		Container:  containerName,
		Follow:     true,
		Timestamps: showTimestamps,
		TailLines:  func() *int64 { i := int64(100); return &i }(),
	}

	req := clientset.CoreV1().Pods(ns).GetLogs(podName, opts)
	stream, err := req.Stream(c.Request.Context())
	if err != nil {
		outChan <- prefix + "Error opening stream: " + err.Error() + "\n"
		return
	}
	defer stream.Close()

	reader := bufio.NewReader(stream)
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			outChan <- prefix + string(line)
		}
		if err != nil {
			if err != io.EOF {
				outChan <- prefix + "Stream ended with error: " + err.Error() + "\n"
			}
			break
		}
	}
}
