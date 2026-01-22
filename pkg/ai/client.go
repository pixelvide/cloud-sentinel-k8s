package ai

import (
	"context"
	"fmt"

	"github.com/pixelvide/cloud-sentinel-k8s/pkg/model"
	openai "github.com/sashabaranov/go-openai"
)

type OpenAIAdapter struct {
	client *openai.Client
	model  string
}

// NewClient returns an AIClient based on the provider in settings
func NewClient(settings *model.AISettings) (AIClient, error) {
	if settings.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	if settings.Provider == "google" {
		return NewGeminiAdapter(settings)
	}

	// Default to OpenAI / Custom OpenAI-compatible
	config := openai.DefaultConfig(settings.APIKey)
	if settings.BaseURL != "" {
		config.BaseURL = settings.BaseURL
	}

	modelName := settings.Model
	if modelName == "" {
		modelName = openai.GPT3Dot5Turbo
	}

	return &OpenAIAdapter{
		client: openai.NewClientWithConfig(config),
		model:  modelName,
	}, nil
}

func (c *OpenAIAdapter) ChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessage, tools []openai.Tool) (openai.ChatCompletionResponse, error) {
	req := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    tools,
	}
	return c.client.CreateChatCompletion(ctx, req)
}
