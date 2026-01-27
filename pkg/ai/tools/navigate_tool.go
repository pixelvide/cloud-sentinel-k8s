package tools

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// --- Navigate To Tool ---

type NavigateToTool struct{}

func (t *NavigateToTool) Name() string { return "navigate_to" }

func (t *NavigateToTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "navigate_to",
			Description: "NAVIGATE the user's UI to a specific page. WARNING: Only use this if the user EXPLICITLY asks to 'go to', 'see', or 'visit' a page/resource. Do NOT use this for your own data gathering.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"page": {
						"type": "string",
						"description": "The page to navigate to. Enum: ['pods', 'deployments', 'services', 'ingresses', 'nodes', 'namespaces', 'settings', 'security', 'helm', 'events']."
					},
					"path": {
						"type": "string",
						"description": "The specific path for a resource. format: '/c/:cluster/:kind/:namespace/:name'. Example: '/c/local/pods/default/nginx-1'."
					}
				},
				"required": ["page"]
			}`),
		},
	}
}

func (t *NavigateToTool) Execute(ctx context.Context, args string) (string, error) {
	var params struct {
		Page string `json:"page"`
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}

	// This tool is special; it doesn't return data to the AI to process.
	// Instead, it triggers a client-side action.
	// The return value here is just for the AI's history.

	if params.Path != "" {
		return fmt.Sprintf("Navigating to path: %s", params.Path), nil
	}
	return fmt.Sprintf("Navigating to page: %s", params.Page), nil
}
