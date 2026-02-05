package kube

import (
	"context"

	"k8s.io/client-go/kubernetes"
)

// Drainer is a helper for draining a node.
type Drainer struct {
	Ctx    context.Context
	Client *kubernetes.Clientset
}

// NewDrainer creates a new Drainer.
func NewDrainer(client *kubernetes.Clientset, ctx context.Context) *Drainer {
	return &Drainer{
		Client: client,
		Ctx:    ctx,
	}
}
