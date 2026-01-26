package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/ai"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/ai/tools"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/cluster"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/model"
	openai "github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

type ChatRequest struct {
	SessionID string `json:"sessionID"` // Optional, if empty create new
	Message   string `json:"message"`
	Model     string `json:"model"` // Optional model override
}

type ChatResponse struct {
	SessionID string `json:"sessionID"`
	Message   string `json:"message"` // The assistant's reply
}

// --- Helpers ---

func resolveAIConfig(user *model.User) (*ai.AIConfig, error) {
	// 1. Authorization and Resolution Logic
	userConfig, err := model.GetUserConfig(user.ID)
	if err != nil || !userConfig.IsAIChatEnabled {
		return nil, fmt.Errorf("AI Chat is disabled for your account")
	}

	// Load AppConfigs for AI governance
	aiAllowUserKeysCfg, _ := model.GetAppConfig(model.CurrentApp.ID, model.AIAllowUserKeys)
	aiForceUserKeysCfg, _ := model.GetAppConfig(model.CurrentApp.ID, model.AIForceUserKeys)

	aiAllowUserKeys := "true"
	if aiAllowUserKeysCfg != nil {
		aiAllowUserKeys = aiAllowUserKeysCfg.Value
	}
	aiForceUserKeys := "false"
	if aiForceUserKeysCfg != nil {
		aiForceUserKeys = aiForceUserKeysCfg.Value
	}

	overrideEnabled := model.IsAIAllowUserOverrideEnabled()

	var resolvedConfig *ai.AIConfig

	// Attempt to find user settings
	var userSettings model.AISettings
	hasUserSettings := false

	if overrideEnabled {
		// Priority: 1. Default, 2. Active, 3. Any
		err = model.DB.Where("user_id = ? AND is_default = ?", user.ID, true).First(&userSettings).Error
		if err != nil {
			err = model.DB.Where("user_id = ? AND is_active = ?", user.ID, true).First(&userSettings).Error
			if err != nil {
				err = model.DB.Where("user_id = ?", user.ID).First(&userSettings).Error
			}
		}
		hasUserSettings = err == nil
	}

	// Fallback Logic
	if aiForceUserKeys == "true" {
		if !hasUserSettings || userSettings.APIKey == "" {
			return nil, fmt.Errorf("administrator requires you to provide your own AI API key in settings")
		}
	}

	if hasUserSettings && (aiAllowUserKeys == "true" || aiForceUserKeys == "true") && userSettings.APIKey != "" {
		// Use user settings
		var profile model.AIProviderProfile
		if err := model.DB.Where("is_enabled = ?", true).First(&profile, userSettings.ProfileID).Error; err == nil {
			modelOverride := userSettings.ModelOverride

			// Validate model override against allowed models
			if len(profile.AllowedModels) > 0 && modelOverride != "" {
				found := false
				for _, m := range profile.AllowedModels {
					if m == modelOverride {
						found = true
						break
					}
				}
				if !found {
					// Fallback to default if override is not allowed
					modelOverride = ""
				}
			}

			resolvedConfig = &ai.AIConfig{
				Provider:     profile.Provider,
				APIKey:       userSettings.APIKey,
				BaseURL:      profile.BaseURL,
				Model:        modelOverride,
				DefaultModel: profile.DefaultModel,
			}
		}
	}

	// Falling back to global system settings (active profile) if not resolved
	if resolvedConfig == nil && aiForceUserKeys != "true" {
		var profile model.AIProviderProfile
		if err := model.DB.Where("is_system = ? AND is_enabled = ?", true, true).First(&profile).Error; err == nil {
			resolvedConfig = &ai.AIConfig{
				Provider:     profile.Provider,
				APIKey:       profile.APIKey,
				BaseURL:      profile.BaseURL,
				Model:        profile.DefaultModel,
				DefaultModel: profile.DefaultModel,
			}
		}
	}

	if resolvedConfig == nil {
		return nil, fmt.Errorf("AI is not configured by the administrator")
	}

	return resolvedConfig, nil
}

func validateAndOverrideModel(resolvedConfig *ai.AIConfig, requestedModel string, user *model.User) {
	if requestedModel == "" {
		return
	}

	// Fetch profile to check allowed models
	var profile model.AIProviderProfile
	var userSettings model.AISettings
	err := model.DB.Where("user_id = ? AND (is_default = ? OR is_active = ?)", user.ID, true, true).First(&userSettings).Error

	profileID := uint(0)
	if err == nil && userSettings.ProfileID != 0 {
		profileID = userSettings.ProfileID
	}

	if profileID != 0 {
		model.DB.First(&profile, profileID)
	} else {
		model.DB.Where("is_system = ?", true).First(&profile)
	}

	if !model.IsAIAllowUserOverrideEnabled() {
		// User override is disabled, force use of default model for the profile
		resolvedConfig.Model = profile.DefaultModel
		return
	}

	if len(profile.AllowedModels) > 0 {
		found := false
		for _, m := range profile.AllowedModels {
			if m == requestedModel {
				found = true
				break
			}
		}
		if found {
			resolvedConfig.Model = requestedModel
		} else {
			klog.Warningf("Chat: requested model %s is not in allowed list for profile %d", requestedModel, profile.ID)
		}
	} else {
		resolvedConfig.Model = requestedModel
	}
}

func getOrCreateSession(sessionID string, userID uint) (*model.AIChatSession, error) {
	var session model.AIChatSession
	if sessionID != "" {
		if err := model.DB.Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at asc")
		}).Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
			return nil, err
		}
	} else {
		session = model.AIChatSession{
			ID:        uuid.NewString(),
			UserID:    userID,
			Title:     "New Chat",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := model.DB.Create(&session).Error; err != nil {
			return nil, err
		}
	}
	return &session, nil
}

func buildMessageHistory(session model.AIChatSession, userMessage string) []openai.ChatCompletionMessage {
	var messages []openai.ChatCompletionMessage

	// System Prompt
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "You are a helpful Kubernetes assistant inside the Cloud Sentinel K8s dashboard. You have access to the cluster via tools. You are specifically designed to assist with Kubernetes, DevOps, and cluster management tasks. If a user asks a question that is entirely unrelated to these topics (e.g., general knowledge, weather, personal advice), politely inform them that you are only able to help with cluster management and DevOps related queries within the Cloud Sentinel context.\n\n**CRITICAL INSTRUCTION FOR MISSING PARAMETERS:**\nIf a user request is missing a key parameter (like a node name, a namespace, or a specific pod), do NOT just ask for the missing information. Instead, use your tools (like 'list_resources' with the appropriate 'kind') to fetch a list of available options, present them to the user in a clear list or table, and ask them to select one. For example, if asked about pods on 'a node', first use 'list_resources' to show all nodes and then ask the user which one they mean.\n\n**MULTI-STEP PLANNING & COMPLEX QUERIES:**\nFor complex queries that cannot be solved with a single tool call (e.g., 'list helm releases on a specific node'), you MUST:\n1. Use your internal reasoning (<thought> tag) to plan the steps (e.g., Step 1: List all pods on the node to find their names/namespaces. Step 2: Use labels or other info from pods to map them to Helm releases). \n2. Execute tools sequentially to gather necessary information.\n3. Clearly explain your plan and progress in the thought block.\n\nIf the user asks for resources but doesn't provide a full name, use the 'list_resources' tool with the 'name_filter' parameter to find what they're looking for. If you need confirmation for a destructive action (like scaling), the tool will enforce it.\n\n**REASONING PRIVACY:**\nYou MUST provide your internal reasoning or 'thinking' process enclosed in <thought> tags for EVERY turn. This is essential for transparency. After the thought block, provide the final response for the user. Use markdown for your final response, including bold text for emphasis and tables for structured data. Be concise. If the tool returns an error about missing cluster context, ask the user to select a cluster in the dashboard.",
	})

	for _, m := range session.Messages {
		msg := openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
		if m.ToolCalls != "" {
			var tcs []openai.ToolCall
			if err := json.Unmarshal([]byte(m.ToolCalls), &tcs); err == nil {
				msg.ToolCalls = tcs
			}
		}
		if m.ToolID != "" {
			msg.ToolCallID = m.ToolID
		}
		messages = append(messages, msg)
	}

	// Add current user message
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userMessage,
	})

	return messages
}

func generateChatTitle(ctx context.Context, aiClient ai.AIClient, userMessage string) string {
	prompt := fmt.Sprintf("Summarize the following user message into a short, descriptive chat title (max 4 words). Output ONLY the title text, no quotes or punctuation: %s", userMessage)
	msgs := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	resp, err := aiClient.ChatCompletion(ctx, msgs, nil)
	if err != nil {
		klog.Errorf("Failed to generate chat title: %v", err)
		return ""
	}

	if len(resp.Choices) > 0 {
		return strings.TrimSpace(resp.Choices[0].Message.Content)
	}
	return ""
}

func executeAIChatStreamLoop(ctx context.Context, aiClient ai.AIClient, session *model.AIChatSession, messages []openai.ChatCompletionMessage, toolDefs []openai.Tool, registry *tools.Registry, toolCtx context.Context, c *gin.Context) (string, error) {
	maxIterations := 5
	var finalContent strings.Builder

	for i := 0; i < maxIterations; i++ {
		stream, err := aiClient.ChatCompletionStream(ctx, messages, toolDefs)
		if err != nil {
			return "", fmt.Errorf("AI Provider error: %w", err)
		}

		var currentAssistantMessage strings.Builder
		var currentToolCalls []openai.ToolCall

		for resp := range stream {
			if len(resp.Choices) == 0 {
				continue
			}
			choice := resp.Choices[0]
			delta := choice.Delta

			if delta.Content != "" {
				currentAssistantMessage.WriteString(delta.Content)
				finalContent.WriteString(delta.Content)
				// Send chunk via SSE
				c.SSEvent("message", gin.H{"content": delta.Content})
				c.Writer.Flush()
			}

			if len(delta.ToolCalls) > 0 {
				// Handle tool call chunks
				for _, tc := range delta.ToolCalls {
					if tc.Index != nil {
						idx := *tc.Index
						for len(currentToolCalls) <= idx {
							currentToolCalls = append(currentToolCalls, openai.ToolCall{})
						}
						if tc.ID != "" {
							currentToolCalls[idx].ID = tc.ID
						}
						if tc.Function.Name != "" {
							currentToolCalls[idx].Function.Name += tc.Function.Name
						}
						if tc.Function.Arguments != "" {
							currentToolCalls[idx].Function.Arguments += tc.Function.Arguments
						}
					} else {
						currentToolCalls = append(currentToolCalls, tc)
					}
				}
			}
		}

		// Construct assistant message
		msg := openai.ChatCompletionMessage{
			Role:      openai.ChatMessageRoleAssistant,
			Content:   currentAssistantMessage.String(),
			ToolCalls: currentToolCalls,
		}
		messages = append(messages, msg)

		// Save assistant message to DB (excluding reasoning)
		dbMsg := model.AIChatMessage{
			SessionID: session.ID,
			Role:      msg.Role,
			Content:   msg.Content,
			CreatedAt: time.Now(),
		}
		if len(msg.ToolCalls) > 0 {
			tcBytes, err := json.Marshal(msg.ToolCalls)
			if err == nil {
				dbMsg.ToolCalls = string(tcBytes)
			}
		}
		model.DB.Create(&dbMsg)

		if len(currentToolCalls) > 0 {
			// Notify user about tool execution
			c.SSEvent("status", gin.H{"status": "Executing tools..."})
			c.Writer.Flush()

			for _, tc := range currentToolCalls {
				klog.Infof("AI executing tool: %s args: %s", tc.Function.Name, tc.Function.Arguments)

				var result string
				if val := toolCtx.Value(tools.ClientSetKey{}); val == nil {
					result = "Error: No active cluster context. Please select a cluster in the dashboard."
				} else {
					res, err := registry.Execute(toolCtx, tc.Function.Name, tc.Function.Arguments)
					if err != nil {
						klog.Errorf("AI tool %s failed: %v", tc.Function.Name, err)
						result = fmt.Sprintf("Error executing tool: %v", err)
					} else {
						result = res
					}
				}

				// Append tool result
				toolMsg := openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    result,
					ToolCallID: tc.ID,
				}
				messages = append(messages, toolMsg)

				model.DB.Create(&model.AIChatMessage{
					SessionID: session.ID,
					Role:      openai.ChatMessageRoleTool,
					Content:   result,
					ToolID:    tc.ID,
					CreatedAt: time.Now(),
				})
			}
			continue
		} else {
			break
		}
	}
	return finalContent.String(), nil
}

func AIChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := getUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 1. Resolve Config
	resolvedConfig, err := resolveAIConfig(user)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// 1.5 Override model if requested specifically in chat
	validateAndOverrideModel(resolvedConfig, req.Model, user)

	// 2. Get ClientSet (for tool context)
	var clientSet *cluster.ClientSet
	if val, ok := c.Get("cluster"); ok && val != nil {
		clientSet = val.(*cluster.ClientSet)
	}

	// 3. Load/Create Session
	session, err := getOrCreateSession(req.SessionID, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// 4. Prepare Client & Registry
	aiClient, err := ai.NewClient(resolvedConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create AI client: " + err.Error()})
		return
	}

	registry := tools.NewRegistry()
	registry.Register(&tools.ListPodsTool{})
	registry.Register(&tools.GetPodLogsTool{})
	registry.Register(&tools.DescribeResourceTool{})
	registry.Register(&tools.ScaleDeploymentTool{})
	registry.Register(&tools.AnalyzeSecurityTool{})
	registry.Register(&tools.CheckImageSecurityTool{})
	registry.Register(&tools.ListResourcesTool{})
	registry.Register(&tools.GetClusterInfoTool{})

	toolDefs := registry.GetDefinitions()

	// 5. Build Message History
	openAIMessages := buildMessageHistory(*session, req.Message)

	// Save user message to DB
	model.DB.Create(&model.AIChatMessage{
		SessionID: session.ID,
		Role:      openai.ChatMessageRoleUser,
		Content:   req.Message,
		CreatedAt: time.Now(),
	})

	// Generate dynamic title if it's a new chat
	if session.Title == "New Chat" {
		go func(sID string, msg string, client ai.AIClient) {
			newTitle := generateChatTitle(context.Background(), client, msg)
			if newTitle != "" {
				model.DB.Model(&model.AIChatSession{}).Where("id = ?", sID).Update("title", newTitle)
			}
		}(session.ID, req.Message, aiClient)
	}

	// 6. Execute Chat Loop
	toolCtx := context.Background()
	if clientSet != nil {
		klog.Infof("AI Chat: Injecting cluster %s into tool context", clientSet.Name)
		toolCtx = context.WithValue(toolCtx, tools.ClientSetKey{}, clientSet)
	}

	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// Send initial session ID
	c.SSEvent("session", gin.H{"sessionID": session.ID})
	c.Writer.Flush()

	_, err = executeAIChatStreamLoop(c.Request.Context(), aiClient, session, openAIMessages, toolDefs, registry, toolCtx, c)
	if err != nil {
		c.SSEvent("error", gin.H{"error": err.Error()})
		c.Writer.Flush()
		return
	}

	// Update session timestamp
	model.DB.Model(&session).Update("updated_at", time.Now())

	c.SSEvent("done", gin.H{})
	c.Writer.Flush()
}
