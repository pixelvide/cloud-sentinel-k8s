package model

import (
	"sync"
	"time"

	"github.com/pixelvide/cloud-sentinel-k8s/pkg/common"
	"k8s.io/klog/v2"
)

type App struct {
	Model
	Name    string `gorm:"uniqueIndex;not null" json:"name"`
	Enabled bool   `gorm:"default:true" json:"enabled"`
}

type AppConfig struct {
	Model
	AppID uint   `gorm:"not null;uniqueIndex:idx_app_key" json:"app_id"`
	Key   string `gorm:"not null;uniqueIndex:idx_app_key" json:"key"`
	Value string `json:"value"`

	// Relationships
	App App `gorm:"foreignKey:AppID" json:"app,omitempty"`
}

type AppUser struct {
	Model
	AppID  uint `gorm:"uniqueIndex:idx_app_user;not null" json:"app_id"`
	UserID uint `gorm:"uniqueIndex:idx_app_user;not null" json:"user_id"`
	Access bool `gorm:"default:false" json:"access"` // "user have access to app or not"

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	App  App  `gorm:"foreignKey:AppID" json:"app,omitempty"`
}

const (
	DefaultUserAccessKey = "DEFAULT_USER_ACCESS"
	LocalLoginEnabledKey = "LOCAL_LOGIN_ENABLED"
	AIAllowUserKeys      = "AI_ALLOW_USER_KEYS"
	AIForceUserKeys      = "AI_FORCE_USER_KEYS"
)

var (
	configCache = make(map[string]string)
	cacheMutex  sync.RWMutex
)

// StartAppConfigRefresher starts a background goroutine to refresh the config cache every 5 minutes.
func StartAppConfigRefresher() {
	refreshCache()
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			refreshCache()
		}
	}()
}

func refreshCache() {
	if CurrentApp == nil {
		return
	}

	var configs []AppConfig
	if err := DB.Where("app_id = ?", CurrentApp.ID).Find(&configs).Error; err != nil {
		klog.Errorf("Failed to refresh app config cache: %v", err)
		return
	}

	newCache := make(map[string]string)
	for _, cfg := range configs {
		newCache[cfg.Key] = cfg.Value
	}

	cacheMutex.Lock()
	configCache = newCache
	cacheMutex.Unlock()
	klog.V(2).Info("App config cache refreshed")
}

func GetApp(name string) (*App, error) {
	var app App
	if err := DB.Where("name = ?", name).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func GetAppConfig(appID uint, key string) (*AppConfig, error) {
	cacheMutex.RLock()
	val, ok := configCache[key]
	cacheMutex.RUnlock()

	if ok {
		return &AppConfig{
			AppID: appID,
			Key:   key,
			Value: val,
		}, nil
	}

	// Fallback to DB if not in cache (e.g. before first refresh or if appID differs)
	var config AppConfig
	if err := DB.Where("app_id = ? AND key = ?", appID, key).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func SetAppConfig(appID uint, key, value string) error {
	var config AppConfig
	err := DB.Where("app_id = ? AND key = ?", appID, key).First(&config).Error
	if err == nil {
		config.Value = value
		err = DB.Save(&config).Error
	} else {
		// If not found, create new
		config = AppConfig{
			AppID: appID,
			Key:   key,
			Value: value,
		}
		err = DB.Create(&config).Error
	}

	if err == nil {
		// Update cache immediately
		cacheMutex.Lock()
		configCache[key] = value
		cacheMutex.Unlock()
	}
	return err
}

func GetAppConfigs(appID uint) ([]AppConfig, error) {
	var configs []AppConfig
	if err := DB.Where("app_id = ?", appID).Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

func IsLocalLoginEnabled() bool {
	var appID uint
	if CurrentApp != nil {
		appID = CurrentApp.ID
	} else {
		app, err := GetApp(common.AppName)
		if err != nil {
			return false
		}
		appID = app.ID
	}

	config, err := GetAppConfig(appID, LocalLoginEnabledKey)
	if err != nil {
		return false
	}
	return config.Value == "true"
}

func CheckOrInitializeUserAccess(userID uint) (bool, error) {
	var appID uint
	if CurrentApp != nil {
		appID = CurrentApp.ID
	} else {
		app, err := GetApp(common.AppName)
		if err != nil {
			return false, err
		}
		appID = app.ID
	}

	// Check if AppUser exists
	var appUser AppUser
	if err := DB.Where("app_id = ? AND user_id = ?", appID, userID).First(&appUser).Error; err == nil {
		return appUser.Access, nil
	}

	// Not exists: Get DefaultUserAccess setting
	config, err := GetAppConfig(appID, DefaultUserAccessKey)
	defaultAccess := err == nil && config.Value == "true"

	// Create AppUser with default access
	newAppUser := AppUser{
		AppID:  appID,
		UserID: userID,
		Access: defaultAccess,
	}
	if err := DB.Create(&newAppUser).Error; err != nil {
		return false, err
	}

	return defaultAccess, nil
}
