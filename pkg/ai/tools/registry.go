package tools

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

type Tool interface {
	Definition() openai.Tool
	Execute(ctx context.Context, args string) (string, error)
	Name() string
}

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}

func (r *Registry) GetDefinitions() []openai.Tool {
	defs := []openai.Tool{}
	for _, t := range r.tools {
		defs = append(defs, t.Definition())
	}
	return defs
}

func (r *Registry) Execute(ctx context.Context, name string, args string) (string, error) {
	tool, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("tool %s not found", name)
	}
	return tool.Execute(ctx, args)
}
