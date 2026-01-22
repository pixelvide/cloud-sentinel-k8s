package model

import (
	"time"

	"gorm.io/gorm"
)

type AISettings struct {
	UserID    uint      `json:"userID" gorm:"primaryKey"`
	Provider  string    `json:"provider"` // "openai", "azure", "custom"
	Model     string    `json:"model"`
	APIKey    string    `json:"apiKey"`
	BaseURL   string    `json:"baseUrl"` // Optional, for local LLMs or Azure
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ChatSession struct {
	ID        string         `json:"id" gorm:"primaryKey"` // UUID
	UserID    uint           `json:"userID" gorm:"index"`
	Title     string         `json:"title"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	Messages  []ChatMessage  `json:"messages" gorm:"foreignKey:SessionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type ChatMessage struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	SessionID string         `json:"sessionID" gorm:"index"`
	Role      string         `json:"role"`                                 // "system", "user", "assistant", "tool"
	Content   string         `json:"content"`                              // Text content
	ToolCalls string         `json:"toolCalls,omitempty" gorm:"type:text"` // JSON encoded tool calls
	ToolID    string         `json:"toolID,omitempty"`                     // For tool messages
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"`
}
