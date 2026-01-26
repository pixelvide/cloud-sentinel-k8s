package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	openai "github.com/sashabaranov/go-openai"
	"google.golang.org/api/option"
	"k8s.io/klog/v2"
)

type GeminiAdapter struct {
	client *genai.Client
	model  string
}

func NewGeminiAdapter(config *AIConfig) (*GeminiAdapter, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	modelName := config.Model
	if modelName == "" {
		modelName = config.DefaultModel
	}
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}

	return &GeminiAdapter{
		client: client,
		model:  modelName,
	}, nil
}

func (g *GeminiAdapter) ChatCompletionStream(ctx context.Context, messages []openai.ChatCompletionMessage, tools []openai.Tool) (chan openai.ChatCompletionStreamResponse, error) {
	gm := g.client.GenerativeModel(g.model)
	klog.Infof("Gemini Stream: ChatCompletion items: %d, tools: %d", len(messages), len(tools))

	// Map Tools
	if len(tools) > 0 {
		var genaiTools []*genai.Tool
		functionDecls := []*genai.FunctionDeclaration{}

		for _, t := range tools {
			if t.Type != openai.ToolTypeFunction || t.Function == nil {
				continue
			}

			rawParams, err := json.Marshal(t.Function.Parameters)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal tool parameters for %s: %w", t.Function.Name, err)
			}
			schema, err := convertJSONSchemaToGenAISchema(rawParams)
			if err != nil {
				return nil, fmt.Errorf("failed to convert tool schema for %s: %w", t.Function.Name, err)
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

	history, systemInstruction := buildGeminiHistory(messages)

	if systemInstruction != nil {
		gm.SystemInstruction = systemInstruction
	}

	if len(history) == 0 {
		return nil, fmt.Errorf("gemini stream: no messages to send")
	}

	cs := gm.StartChat()
	cs.History = history[:len(history)-1]
	lastMsg := history[len(history)-1]

	streamChan := make(chan openai.ChatCompletionStreamResponse, 100)

	go func() {
		defer close(streamChan)
		iter := cs.SendMessageStream(ctx, lastMsg.Parts...)
		for {
			resp, err := iter.Next()
			if err != nil {
				if errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "iterator done") || strings.Contains(err.Error(), "EOF") {
					break
				}
				klog.Errorf("Gemini Stream error: %v", err)
				return
			}

			// Convert and send
			sResp := convertGeminiStreamResponseToOpenAI(resp)
			streamChan <- sResp
		}
	}()

	return streamChan, nil
}

func convertGeminiStreamResponseToOpenAI(resp *genai.GenerateContentResponse) openai.ChatCompletionStreamResponse {
	choices := []openai.ChatCompletionStreamChoice{}

	for i, cand := range resp.Candidates {
		delta := openai.ChatCompletionStreamChoiceDelta{
			Role: openai.ChatMessageRoleAssistant,
		}

		var contentBuilder strings.Builder
		for _, part := range cand.Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				contentBuilder.WriteString(string(txt))
			} else if fnCall, ok := part.(genai.FunctionCall); ok {
				argsBytes, err := json.Marshal(fnCall.Args)
				if err != nil {
					klog.Errorf("Gemini: failed to marshal function args for %s: %v", fnCall.Name, err)
					argsBytes = []byte("{}")
				}
				delta.ToolCalls = append(delta.ToolCalls, openai.ToolCall{
					ID:   "call_" + fnCall.Name,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      fnCall.Name,
						Arguments: string(argsBytes),
					},
				})
			}
		}
		delta.Content = contentBuilder.String()

		choices = append(choices, openai.ChatCompletionStreamChoice{
			Index: i,
			Delta: delta,
		})
	}

	return openai.ChatCompletionStreamResponse{
		ID:      "gemini-stream-resp",
		Object:  "chat.completion.chunk",
		Created: 0,
		Model:   "gemini",
		Choices: choices,
	}
}

func (g *GeminiAdapter) ChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessage, tools []openai.Tool) (openai.ChatCompletionResponse, error) {
	gm := g.client.GenerativeModel(g.model)
	klog.Infof("Gemini: ChatCompletion items: %d, tools: %d", len(messages), len(tools))

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

			rawParams, err := json.Marshal(t.Function.Parameters)
			if err != nil {
				return openai.ChatCompletionResponse{}, fmt.Errorf("failed to marshal tool parameters for %s: %w", t.Function.Name, err)
			}
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

	history, systemInstruction := buildGeminiHistory(messages)

	if systemInstruction != nil {
		gm.SystemInstruction = systemInstruction
	}

	// Generate Content
	if len(history) == 0 {
		return openai.ChatCompletionResponse{}, fmt.Errorf("gemini: no messages to send")
	}

	// Since we built the full history, we don't use StartChat with history,
	// because `GenerateContent` is for a single turn unless we manage session.
	// But `cs := gm.StartChat()` allows setting History.
	cs := gm.StartChat()
	cs.History = history[:len(history)-1] // All but last
	lastMsg := history[len(history)-1]

	// Send last message
	klog.V(2).Infof("Gemini: Sending message with %d history items", len(cs.History))
	resp, err := cs.SendMessage(ctx, lastMsg.Parts...)
	if err != nil {
		klog.Errorf("Gemini: SendMessage failed: %v", err)
		return openai.ChatCompletionResponse{}, fmt.Errorf("gemini request failed: %w", err)
	}

	converted := convertGeminiResponseToOpenAI(resp)
	if len(converted.Choices) > 0 && len(converted.Choices[0].Message.ToolCalls) > 0 {
		klog.Infof("Gemini: AI returned %d tool calls", len(converted.Choices[0].Message.ToolCalls))
	}
	return converted, nil
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
				argsBytes, err := json.Marshal(fnCall.Args)
				if err != nil {
					klog.Errorf("Gemini: failed to marshal function args for %s: %v", fnCall.Name, err)
					argsBytes = []byte("{}")
				}
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
			Index:        i,
			Message:      msg,
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

func buildGeminiHistory(messages []openai.ChatCompletionMessage) ([]*genai.Content, *genai.Content) {
	var systemInstruction *genai.Content
	var history []*genai.Content

	for _, m := range messages {
		if m.Role == openai.ChatMessageRoleSystem {
			systemInstruction = &genai.Content{
				Parts: []genai.Part{genai.Text(m.Content)},
			}
			continue
		}

		role := "user"
		switch m.Role {
		case openai.ChatMessageRoleAssistant:
			role = "model"
		case openai.ChatMessageRoleTool:
			role = "user"
		}

		parts := []genai.Part{}

		if m.Content != "" && m.Role != openai.ChatMessageRoleTool {
			parts = append(parts, genai.Text(m.Content))
		}

		if len(m.ToolCalls) > 0 {
			for _, tc := range m.ToolCalls {
				args := map[string]interface{}{}
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
					args = map[string]interface{}{"args": tc.Function.Arguments}
				}
				parts = append(parts, genai.FunctionCall{
					Name: tc.Function.Name,
					Args: args,
				})
			}
		}

		if m.Role == openai.ChatMessageRoleTool {
			name := findToolName(messages, m.ToolCallID)
			var response map[string]interface{}
			if err := json.Unmarshal([]byte(m.Content), &response); err != nil {
				response = map[string]interface{}{"result": m.Content}
			}
			parts = append(parts, genai.FunctionResponse{
				Name:     name,
				Response: response,
			})
		}

		if len(parts) > 0 {
			if len(history) > 0 && history[len(history)-1].Role == role {
				history[len(history)-1].Parts = append(history[len(history)-1].Parts, parts...)
			} else {
				history = append(history, &genai.Content{
					Role:  role,
					Parts: parts,
				})
			}
		}
	}
	return history, systemInstruction
}
