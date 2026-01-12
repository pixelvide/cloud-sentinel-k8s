package models

import (
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	Email            string         `gorm:"uniqueIndex;not null" json:"email"`
	Name             string         `json:"name"`
	Sub              string         `gorm:"uniqueIndex;not null" json:"sub"` // OIDC Subject
	StorageNamespace string         `gorm:"uniqueIndex;not null" json:"storage_namespace"`
	GitlabConfigs    []GitlabConfig `json:"gitlab_configs,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.StorageNamespace = uuid.New().String()

	// Create user data directory
	dataDir := "/data"
	userDir := filepath.Join(dataDir, u.StorageNamespace)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return err
	}

	return
}
