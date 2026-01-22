package tools

import (
	"context"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

type MockTool struct{}

func (t *MockTool) Name() string { return "mock_tool" }
func (t *MockTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name: "mock_tool",
		},
	}
}
func (t *MockTool) Execute(ctx context.Context, args string) (string, error) {
	return "executed " + args, nil
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()
	mock := &MockTool{}
	r.Register(mock)

	defs := r.GetDefinitions()
	assert.Len(t, defs, 1)
	assert.Equal(t, "mock_tool", defs[0].Function.Name)

	res, err := r.Execute(context.Background(), "mock_tool", "test")
	assert.NoError(t, err)
	assert.Equal(t, "executed test", res)

	_, err = r.Execute(context.Background(), "unknown", "test")
	assert.Error(t, err)
}
