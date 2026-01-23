package model

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserConfig struct {
	Model
	UserID           uint   `json:"user_id" gorm:"uniqueIndex:idx_user_config_user_id;not null"`
	StorageNamespace string `json:"storage_namespace" gorm:"uniqueIndex;not null"`

	// Settings
	IsAIChatEnabled bool `json:"is_ai_chat_enabled" gorm:"default:true"`

	// Relationships
	User User `json:"-" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (UserConfig) TableName() string {
	return "user_configs"
}

func (u *UserConfig) BeforeCreate(tx *gorm.DB) (err error) {
	if u.StorageNamespace == "" {
		u.StorageNamespace = uuid.New().String()
	}
	return
}

// GetUserConfig retrieves the user config for a given user ID.
// If it does not exist, it creates one with a default StorageNamespace (UUIDv4).
func GetUserConfig(userID uint) (*UserConfig, error) {
	var config UserConfig
	err := DB.Where("user_id = ?", userID).First(&config).Error
	if err == nil {
		return &config, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new config if not found
	config = UserConfig{
		UserID: userID,
	}

	if err := DB.Create(&config).Error; err != nil {
		// Handle potential race condition where it was created in between
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "unique constraint") {
			return GetUserConfig(userID)
		}
		return nil, err
	}

	return &config, nil
}
