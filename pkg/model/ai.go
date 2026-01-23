package model

import (
	"time"

	"gorm.io/gorm"
)

type AIProviderProfile struct {
	Model
	Name              string      `json:"name"`
	Provider          string      `json:"provider"` // "gemini", "openai", "azure", "custom"
	BaseURL           string      `json:"baseUrl"`
	DefaultModel      string      `json:"defaultModel"`
	APIKey            string      `json:"apiKey" gorm:"type:text"` // Global key for this profile
	IsSystem          bool        `json:"isSystem"`
	IsEnabled         bool        `json:"isEnabled" gorm:"default:true"` // If false, profile is hidden from users
	AllowUserOverride bool        `json:"allowUserOverride"`             // If true, users can provide their own key
	AllowedModels     SliceString `json:"allowedModels"`                 // Comma-separated in DB, array in JSON
}

func (AIProviderProfile) TableName() string {
	return "ai_provider_profiles"
}

type AISettings struct {
	Model
	UserID        uint   `json:"userID" gorm:"uniqueIndex:idx_user_profile"`
	ProfileID     uint   `json:"profileID" gorm:"uniqueIndex:idx_user_profile"`
	APIKey        string `json:"apiKey"`
	ModelOverride string `json:"modelOverride"`
	IsActive      bool   `json:"isActive"`
	IsDefault     bool   `json:"isDefault"` // If true, this is the user's default AI profile
}

func (AISettings) TableName() string {
	return "k8s_ai_settings"
}

type AIChatSession struct {
	ID        string          `json:"id" gorm:"primaryKey"` // UUID
	UserID    uint            `json:"userID" gorm:"index"`
	Title     string          `json:"title"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
	DeletedAt gorm.DeletedAt  `json:"deletedAt" gorm:"index"`
	Messages  []AIChatMessage `json:"messages" gorm:"foreignKey:SessionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (AIChatSession) TableName() string {
	return "k8s_ai_chat_sessions"
}

type AIChatMessage struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	SessionID string         `json:"sessionID" gorm:"index"`
	Role      string         `json:"role"`                                 // "system", "user", "assistant", "tool"
	Content   string         `json:"content"`                              // Text content
	ToolCalls string         `json:"toolCalls,omitempty" gorm:"type:text"` // JSON encoded tool calls
	ToolID    string         `json:"toolID,omitempty"`                     // For tool messages
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"`
}

func (AIChatMessage) TableName() string {
	return "k8s_ai_chat_messages"
}
