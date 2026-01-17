package api

import (
	"io"
	"log"
	"net/http"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// HandleExec manages WebSocket connection for terminal
func HandleExec(c *gin.Context) {
	ns := c.Query("namespace")
	pod := c.Query("pod")
	container := c.Query("container")
	ctxName := GetKubeContext(c)

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

	// Get user from context or session?
	// Note: auth.AuthMiddleware usually sets it but check if it's available in WS request
	// (Browsers send cookies automatically, so AuthMiddleware should work if registered)
	userVal, exists := c.Get("user")
	var storageNamespace string
	if exists {
		user := userVal.(*models.User)
		storageNamespace = user.StorageNamespace

		// Record Audit Log for exec
		RecordAuditLog(c, "POD_EXEC", gin.H{
			"namespace": ns,
			"pod":       pod,
			"container": container,
			"context":   ctxName,
		})
	}

	clientset, restConfig, err := GetClientInfo(storageNamespace, ctxName)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("Config Error: "+err.Error()))
		return
	}

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(ns).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Container: container,
		Command:   []string{"/bin/sh"},
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(restConfig, "POST", req.URL())
	if err != nil {
		log.Printf("Executor init failed: %v", err)
		ws.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		return
	}

	handler := &WsStreamHandler{Ws: ws, ResizeChan: make(chan remotecommand.TerminalSize)}

	// Handle incoming WS messages (Input + Resize) in a goroutine
	go func() {
		for {
			_, p, err := ws.ReadMessage()
			if err != nil {
				return
			}
			// Simple protocol: If starts with '{', assume JSON resize?
			// Or use xterm.js binary/text.
			// Let's assume raw text is Input, and specific JSON is resize.
			// For MVP, just Input.
			// To support resize, we need a structure.
			// Let's decode JSON: e.g. {type: "input", data: "..."} or {type: "resize", rows: ...}
			// BUT xterm.js addon-attach sends raw string? No, we customize it.

			// Let's stick to simple: Text Message = Input.
			// If we want resize later, we can add it.
			handler.InputChan <- p
		}
	}()

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:             handler,
		Stdout:            handler,
		Stderr:            handler,
		Tty:               true,
		TerminalSizeQueue: handler,
	})

	if err != nil {
		log.Printf("Stream error: %v", err)
	}
}

type WsStreamHandler struct {
	Ws         *websocket.Conn
	ResizeChan chan remotecommand.TerminalSize
	InputChan  chan []byte
}

func (w *WsStreamHandler) Read(p []byte) (size int, err error) {
	if w.InputChan == nil {
		w.InputChan = make(chan []byte, 10)
	}

	data, ok := <-w.InputChan
	if !ok {
		return 0, io.EOF
	}
	return copy(p, data), nil
}

func (w *WsStreamHandler) Write(p []byte) (size int, err error) {
	err = w.Ws.WriteMessage(websocket.TextMessage, p) // xterm expects text or binary
	return len(p), err
}

func (w *WsStreamHandler) Next() *remotecommand.TerminalSize {
	// ret := <-w.ResizeChan
	// return &ret
	return nil // Disable resize for MVP simplicity
}
