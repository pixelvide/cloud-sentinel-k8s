package mcp

import (
	"context"
	"net/http"

	"github.com/mark3labs/mcp-go/server"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/cluster"
	"k8s.io/klog/v2"
)

type MCPServer struct {
	server *server.MCPServer
	cm     *cluster.ClusterManager
}

func NewMCPServer(cm *cluster.ClusterManager) *MCPServer {
	s := server.NewMCPServer(
		"cloud-sentinel-mcp",
		"1.0.0",
		server.WithLogging(),
	)

	m := &MCPServer{
		server: s,
		cm:     cm,
	}

	m.registerTools()
	return m
}

func (m *MCPServer) ServeStdio() error {
	klog.Info("Starting MCP server on stdio")
	stdio := server.NewStdioServer(m.server)
	return stdio.Listen(context.Background(), nil, nil)
}

func (m *MCPServer) SSEHandler(baseURL string) http.Handler {
	sse := server.NewSSEServer(m.server,
		server.WithBaseURL(baseURL),
		server.WithStaticBasePath("/api/v1/mcp"),
	)
	return sse
}
