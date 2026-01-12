package api

import (
	"bufio"
	"io"
	"log"
	"net/http"

	"cloud-sentinel-k8s/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	v1 "k8s.io/api/core/v1"
)

// HandleLogs streams logs from a pod via WebSocket
func HandleLogs(c *gin.Context) {
	ns := c.Query("namespace")
	pod := c.Query("pod")
	container := c.Query("container")
	ctxName := c.Query("context")
	timestampsStr := c.Query("timestamps")

	if ns == "" || pod == "" {
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

	showTimestamps := true
	if timestampsStr == "false" {
		showTimestamps = false
	}

	opts := &v1.PodLogOptions{
		Container:  container,
		Follow:     true,
		Previous:   false,
		Timestamps: showTimestamps,
		TailLines:  func() *int64 { i := int64(100); return &i }(),
	}

	req := clientset.CoreV1().Pods(ns).GetLogs(pod, opts)
	stream, err := req.Stream(c.Request.Context())
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("Error opening stream: "+err.Error()))
		return
	}
	defer stream.Close()

	reader := bufio.NewReader(stream)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				ws.WriteMessage(websocket.TextMessage, []byte("Stream error: "+err.Error()))
			}
			break
		}
		if len(line) > 0 {
			err = ws.WriteMessage(websocket.TextMessage, line)
			if err != nil {
				// Client disconnected?
				break
			}
		}
	}
}
