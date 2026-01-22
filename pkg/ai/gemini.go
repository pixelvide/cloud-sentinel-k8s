package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/model"
	openai "github.com/sashabaranov/go-openai"
	"google.golang.org/api/option"
)

type GeminiAdapter struct {
	client *genai.Client
	model  string
}

func NewGeminiAdapter(settings *model.AISettings) (*GeminiAdapter, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(settings.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	modelName := settings.Model
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}

	return &GeminiAdapter{
		client: client,
		model:  modelName,
	}, nil
}

func (g *GeminiAdapter) ChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessage, tools []openai.Tool) (openai.ChatCompletionResponse, error) {
	gm := g.client.GenerativeModel(g.model)

	// Map Tools
	if len(tools) > 0 {
		var genaiTools []*genai.Tool
		functionDecls := []*genai.FunctionDeclaration{}

		for _, t := range tools {
			if t.Type != openai.ToolTypeFunction || t.Function == nil {
				continue
			}

			// Gemini Schema is distinct but similar to OpenAPI schema
			// Since openai.FunctionDefinition.Parameters is json.RawMessage, we have to parse it
			// to convert to genai.Schema. This is complex.
			// Simplified approach: We assume the tool definition is simple enough or we construct it.
			// Ideally, we should parse the JSON schema.

			// For now, let's attempt to reconstruct a basic Schema from the RawMessage if possible,
			// or use a generic "Any" schema if Gemini supports it (it doesn't easily).
			// A better approach for this adapter is to decode the openai JSON schema.

			rawParams, _ := json.Marshal(t.Function.Parameters)
			schema, err := convertJSONSchemaToGenAISchema(rawParams)
			if err != nil {
				return openai.ChatCompletionResponse{}, fmt.Errorf("failed to convert tool schema for %s: %w", t.Function.Name, err)
			}

			functionDecls = append(functionDecls, &genai.FunctionDeclaration{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  schema,
			})
		}
		if len(functionDecls) > 0 {
			genaiTools = append(genaiTools, &genai.Tool{
				FunctionDeclarations: functionDecls,
			})
			gm.Tools = genaiTools
		}
	}

	// Map Messages to History + Last Message
	// Gemini uses a ChatSession object which maintains history, or we can send contents manually.
	// Since we are stateless (loading history from DB each time), we will construct the Content list.

	// Separate System Prompt if present
	var systemInstruction *genai.Content
	var history []*genai.Content

	for _, m := range messages {
		role := ""
		switch m.Role {
		case openai.ChatMessageRoleSystem:
			systemInstruction = &genai.Content{
				Parts: []genai.Part{genai.Text(m.Content)},
			}
			continue // System prompt is handled separately in Gemini
		case openai.ChatMessageRoleUser:
			role = "user"
		case openai.ChatMessageRoleAssistant:
			role = "model"
		case openai.ChatMessageRoleTool:
			role = "function" // Gemini uses 'function' role for tool outputs
		default:
			role = "user"
		}

		parts := []genai.Part{}

		// Text Content
		if m.Content != "" {
			parts = append(parts, genai.Text(m.Content))
		}

		// Tool Calls (Assistant -> User/Tool)
		if len(m.ToolCalls) > 0 {
			for _, tc := range m.ToolCalls {
				args := map[string]interface{}{}
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
					// Fallback if args aren't valid JSON map
					args = map[string]interface{}{"args": tc.Function.Arguments}
				}
				parts = append(parts, genai.FunctionCall{
					Name: tc.Function.Name,
					Args: args,
				})
			}
		}

		// Tool Responses (Tool -> Assistant)
		if m.Role == openai.ChatMessageRoleTool {
			// In OpenAI, ToolCallID links the response.
			// In Gemini, we need to wrap it in FunctionResponse.
			// We need to assume the previous message was the call, OR just send the response.
			// Gemini expects FunctionResponse parts.

			// We don't have the function name in the openai.ChatMessageRoleTool message struct directly
			// (it's usually in the ToolCallID mapping), but for simplicity here we might need to look it up
			// or rely on implicit ordering.
			// However, OpenAI's ChatMessage struct usually has ToolCallID.

			// Problem: OpenAI message doesn't store the function name in the Tool role message, only the content and ToolCallID.
			// Gemini NEEDS the function name in FunctionResponse.
			// We have to scan backwards to find the tool call name?
			// This is expensive.

			// Alternative: We store tool name in our DB? No we don't.
			// Let's try to infer or carry it.
			// Actually, in the Chat Loop (handlers/ai_chat.go), we know the tool name when we create the tool response message.
			// But here we are reconstructing from DB history.

			// Hack: In `pkg/model/ai.go`, we have `ToolID`.
			// We need the Name.
			// We will look back in the messages slice to find the assistant message with the matching ToolCallID.
			name := findToolName(messages, m.ToolCallID)

			var response map[string]interface{}
			// Try to parse content as JSON, otherwise wrap string
			if err := json.Unmarshal([]byte(m.Content), &response); err != nil {
				response = map[string]interface{}{"result": m.Content}
			}

			parts = append(parts, genai.FunctionResponse{
				Name:     name,
				Response: response,
			})
		}

		if len(parts) > 0 {
			history = append(history, &genai.Content{
				Role:  role,
				Parts: parts,
			})
		}
	}

	if systemInstruction != nil {
		gm.SystemInstruction = systemInstruction
	}

	// Generate Content
	// Since we built the full history, we don't use StartChat with history,
	// because `GenerateContent` is for a single turn unless we manage session.
	// But `cs := gm.StartChat()` allows setting History.
	cs := gm.StartChat()
	cs.History = history[:len(history)-1] // All but last
	lastMsg := history[len(history)-1]

	// Send last message
	resp, err := cs.SendMessage(ctx, lastMsg.Parts...)
	if err != nil {
		return openai.ChatCompletionResponse{}, fmt.Errorf("gemini request failed: %w", err)
	}

	return convertGeminiResponseToOpenAI(resp), nil
}

// Helpers

func findToolName(messages []openai.ChatCompletionMessage, toolID string) string {
	for _, m := range messages {
		if m.Role == openai.ChatMessageRoleAssistant {
			for _, tc := range m.ToolCalls {
				if tc.ID == toolID {
					return tc.Function.Name
				}
			}
		}
	}
	return "unknown_tool"
}

func convertJSONSchemaToGenAISchema(raw json.RawMessage) (*genai.Schema, error) {
	// Basic implementation: Parse JSON and recursively build genai.Schema
	// This is a simplified parser covering common types used in our tools.
	var def map[string]interface{}
	if err := json.Unmarshal(raw, &def); err != nil {
		return nil, err
	}
	return buildSchema(def), nil
}

func buildSchema(def map[string]interface{}) *genai.Schema {
	t, _ := def["type"].(string)
	s := &genai.Schema{}

	switch t {
	case "object":
		s.Type = genai.TypeObject
		props, ok := def["properties"].(map[string]interface{})
		if ok {
			s.Properties = make(map[string]*genai.Schema)
			for k, v := range props {
				if vMap, ok := v.(map[string]interface{}); ok {
					s.Properties[k] = buildSchema(vMap)
				}
			}
		}
		required, ok := def["required"].([]interface{})
		if ok {
			for _, r := range required {
				if rStr, ok := r.(string); ok {
					s.Required = append(s.Required, rStr)
				}
			}
		}
	case "string":
		s.Type = genai.TypeString
		if enum, ok := def["enum"].([]interface{}); ok {
			for _, e := range enum {
				if eStr, ok := e.(string); ok {
					s.Enum = append(s.Enum, eStr)
				}
			}
		}
	case "integer":
		s.Type = genai.TypeInteger
	case "number":
		s.Type = genai.TypeNumber
	case "boolean":
		s.Type = genai.TypeBoolean
	case "array":
		s.Type = genai.TypeArray
		if items, ok := def["items"].(map[string]interface{}); ok {
			s.Items = buildSchema(items)
		}
	}

	if desc, ok := def["description"].(string); ok {
		s.Description = desc
	}

	return s
}

func convertGeminiResponseToOpenAI(resp *genai.GenerateContentResponse) openai.ChatCompletionResponse {
	// Convert Gemini Candidates to OpenAI Choices
	choices := []openai.ChatCompletionChoice{}

	for i, cand := range resp.Candidates {
		msg := openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleAssistant,
		}

		var contentBuilder strings.Builder

		for _, part := range cand.Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				contentBuilder.WriteString(string(txt))
			} else if fnCall, ok := part.(genai.FunctionCall); ok {
				// Convert FunctionCall to ToolCall
				argsBytes, _ := json.Marshal(fnCall.Args)
				msg.ToolCalls = append(msg.ToolCalls, openai.ToolCall{
					ID:   "call_" + fnCall.Name, // Gemini doesn't give IDs, generate one? Or use Name if unique per turn.
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      fnCall.Name,
						Arguments: string(argsBytes),
					},
				})
			}
		}
		msg.Content = contentBuilder.String()

		choices = append(choices, openai.ChatCompletionChoice{
			Index:   i,
			Message: msg,
			FinishReason: openai.FinishReasonStop, // Approximate
		})
	}

	return openai.ChatCompletionResponse{
		ID:      "gemini-resp",
		Object:  "chat.completion",
		Created: 0,
		Model:   "gemini",
		Choices: choices,
	}
}
