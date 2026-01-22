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
}

type ChatResponse struct {
	SessionID string `json:"sessionID"`
	Message   string `json:"message"` // The assistant's reply
}

func Chat(c *gin.Context) {
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

	// 1. Get Settings
	var settings model.AISettings
	if err := model.DB.First(&settings, "user_id = ?", user.ID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "AI not configured. Please go to settings."})
		return
	}

	// 2. Get ClientSet (for tools)
	val, _ := c.Get("cluster")
	var clientSet *cluster.ClientSet
	if val != nil {
		clientSet = val.(*cluster.ClientSet)
	}

	// 3. Load/Create Session
	var session model.ChatSession
	if req.SessionID != "" {
		if err := model.DB.Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at asc")
		}).Where("id = ? AND user_id = ?", req.SessionID, user.ID).First(&session).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}
	} else {
		session = model.ChatSession{
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
	aiClient, err := ai.NewClient(&settings)
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
	model.DB.Create(&model.ChatMessage{
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
		toolCtx = context.WithValue(toolCtx, tools.ClientSetKey{}, clientSet)
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
		dbMsg := model.ChatMessage{
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
					res, err := registry.Execute(toolCtx, tc.Function.Name, tc.Function.Arguments)
					if err != nil {
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

				model.DB.Create(&model.ChatMessage{
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
