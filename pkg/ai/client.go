package ai

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIAdapter struct {
	client *openai.Client
	model  string
}

// NewClient returns an AIClient based on the provider in config
func NewClient(config *AIConfig) (AIClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	if config.Provider == "google" || config.Provider == "gemini" {
		return NewGeminiAdapter(config)
	}

	// Default to OpenAI / Custom OpenAI-compatible
	oaConfig := openai.DefaultConfig(config.APIKey)
	if config.BaseURL != "" {
		oaConfig.BaseURL = config.BaseURL
	}

	modelName := config.Model
	if modelName == "" {
		modelName = config.DefaultModel
	}
	if modelName == "" {
		modelName = openai.GPT3Dot5Turbo
	}

	return &OpenAIAdapter{
		client: openai.NewClientWithConfig(oaConfig),
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
func (c *OpenAIAdapter) ChatCompletionStream(ctx context.Context, messages []openai.ChatCompletionMessage, tools []openai.Tool) (chan openai.ChatCompletionStreamResponse, error) {
	req := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    tools,
		Stream:   true,
	}

	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}

	streamChan := make(chan openai.ChatCompletionStreamResponse, 100)

	go func() {
		defer close(streamChan)
		defer func() {
			_ = stream.Close()
		}()

		for {
			resp, err := stream.Recv()
			if err != nil {
				return
			}
			streamChan <- resp
		}
	}()

	return streamChan, nil
}
