package ai

import (
	"context"
	"fmt"

	"github.com/pixelvide/cloud-sentinel-k8s/pkg/model"
	openai "github.com/sashabaranov/go-openai"
)

type Client struct {
	client *openai.Client
	model  string
}

func NewClient(settings *model.AISettings) (*Client, error) {
	if settings.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	config := openai.DefaultConfig(settings.APIKey)
	if settings.BaseURL != "" {
		config.BaseURL = settings.BaseURL
	}

	// Default to gpt-4o-mini if not specified, but usually the UI will send it
	modelName := settings.Model
	if modelName == "" {
		modelName = openai.GPT3Dot5Turbo
	}

	return &Client{
		client: openai.NewClientWithConfig(config),
		model:  modelName,
	}, nil
}

func (c *Client) ChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessage, tools []openai.Tool) (openai.ChatCompletionResponse, error) {
	return c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    tools,
	})
}
