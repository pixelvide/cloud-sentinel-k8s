package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

	// 1. Authorization and Resolution Logic
	userConfig, err := model.GetUserConfig(user.ID)
	if err != nil || !userConfig.IsAIChatEnabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "AI Chat is disabled for your account."})
		return
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

	var resolvedConfig *ai.AIConfig

	// Attempt to find user settings
	var userSettings model.AISettings
	// Priority: 1. Default, 2. Active, 3. Any
	err = model.DB.Where("user_id = ? AND is_default = ?", user.ID, true).First(&userSettings).Error
	if err != nil {
		err = model.DB.Where("user_id = ? AND is_active = ?", user.ID, true).First(&userSettings).Error
		if err != nil {
			err = model.DB.Where("user_id = ?", user.ID).First(&userSettings).Error
		}
	}
	hasUserSettings := err == nil

	// Fallback Logic
	if aiForceUserKeys == "true" {
		if !hasUserSettings || userSettings.APIKey == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Administrator requires you to provide your own AI API key in settings."})
			return
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "AI is not configured by the administrator."})
		return
	}

	// 1.5 Override model if requested specifically in chat
	if req.Model != "" {
		// If we use system profile, we should check its allowed models
		// If we use user profile, we already checked its allowed models for userSettings.ModelOverride,
		// but we still need to check it for req.Model.

		// We need to fetch the profile again or keep track of it to check AllowedModels.
		// Since we want to be efficient, let's just fetch it if it has AllowedModels.
		var profile model.AIProviderProfile
		// Find which profile we are using. If we have profileID in userSettings, use that, else use system.
		profileID := uint(0)
		if hasUserSettings && userSettings.ProfileID != 0 {
			profileID = userSettings.ProfileID
		}

		if profileID != 0 {
			model.DB.First(&profile, profileID)
		} else {
			model.DB.Where("is_system = ?", true).First(&profile)
		}

		if len(profile.AllowedModels) > 0 {
			found := false
			for _, m := range profile.AllowedModels {
				if m == req.Model {
					found = true
					break
				}
			}
			if found {
				resolvedConfig.Model = req.Model
			} else {
				klog.Warningf("Chat: requested model %s is not in allowed list for profile %d", req.Model, profile.ID)
			}
		} else {
			resolvedConfig.Model = req.Model
		}
	}

	// 2. Get ClientSet (for tools)
	val, _ := c.Get("cluster")
	var clientSet *cluster.ClientSet
	if val != nil {
		clientSet = val.(*cluster.ClientSet)
	}

	// 3. Load/Create Session
	var session model.AIChatSession
	if req.SessionID != "" {
		if err := model.DB.Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at asc")
		}).Where("id = ? AND user_id = ?", req.SessionID, user.ID).First(&session).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}
	} else {
		session = model.AIChatSession{
			ID:        uuid.NewString(),
			UserID:    user.ID,
			Title:     "New Chat",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := model.DB.Create(&session).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
			return
		}
	}

	// 4. Prepare Context & Tools
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

	toolDefs := registry.GetDefinitions()

	// 5. Construct Message History
	var openAIMessages []openai.ChatCompletionMessage

	// System Prompt
	openAIMessages = append(openAIMessages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "You are a helpful Kubernetes assistant inside the Cloud Sentinel K8s dashboard. You have access to the cluster via tools. If the user asks to perform an action, use the appropriate tool. If you need confirmation for a destructive action (like scaling), the tool will enforce it. Be concise. If the tool returns an error about missing cluster context, ask the user to select a cluster in the dashboard.",
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
		openAIMessages = append(openAIMessages, msg)
	}

	// Add current user message
	userMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.Message,
	}
	openAIMessages = append(openAIMessages, userMsg)

	// Save user message to DB
	model.DB.Create(&model.AIChatMessage{
		SessionID: session.ID,
		Role:      openai.ChatMessageRoleUser,
		Content:   req.Message,
		CreatedAt: time.Now(),
	})

	// 6. Loop for Tool Execution
	maxIterations := 5
	var finalResponse string

	// Context for tools
	toolCtx := context.Background()
	if clientSet != nil {
		klog.Infof("AI Chat: Injecting cluster %s into tool context", clientSet.Name)
		toolCtx = context.WithValue(toolCtx, "cluster_client", clientSet)
	} else {
		klog.Warningf("AI Chat: No cluster context found in Gin context")
	}

	for i := 0; i < maxIterations; i++ {
		resp, err := aiClient.ChatCompletion(c.Request.Context(), openAIMessages, toolDefs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "AI Provider error: " + err.Error()})
			return
		}

		if len(resp.Choices) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Empty response from AI"})
			return
		}

		choice := resp.Choices[0]
		msg := choice.Message

		openAIMessages = append(openAIMessages, msg)

		// Save assistant message
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

		if len(msg.ToolCalls) > 0 {
			// Execute Tools
			for _, tc := range msg.ToolCalls {
				klog.Infof("AI executing tool: %s args: %s", tc.Function.Name, tc.Function.Arguments)

				var result string
				if clientSet == nil {
					result = "Error: No active cluster context. Please select a cluster in the dashboard."
				} else {
					klog.Infof("AI executing tool: %s", tc.Function.Name)
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
				openAIMessages = append(openAIMessages, toolMsg)

				model.DB.Create(&model.AIChatMessage{
					SessionID: session.ID,
					Role:      openai.ChatMessageRoleTool,
					Content:   result,
					ToolID:    tc.ID,
					CreatedAt: time.Now(),
				})
			}
			// Loop again
			continue
		} else {
			// Final response
			finalResponse = msg.Content
			break
		}
	}

	// Update session timestamp
	model.DB.Model(&session).Update("updated_at", time.Now())

	c.JSON(http.StatusOK, ChatResponse{
		SessionID: session.ID,
		Message:   finalResponse,
	})
}
