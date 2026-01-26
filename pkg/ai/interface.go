package ai

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

// AIClient defines the interface that all AI providers must implement
type AIClient interface {
	ChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessage, tools []openai.Tool) (openai.ChatCompletionResponse, error)
	ChatCompletionStream(ctx context.Context, messages []openai.ChatCompletionMessage, tools []openai.Tool) (chan openai.ChatCompletionStreamResponse, error)
}
