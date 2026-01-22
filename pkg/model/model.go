package model

import (
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/common"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"k8s.io/klog/v2"
)

var (
	DB *gorm.DB

	CurrentApp *App

	once sync.Once
)

type Model struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func InitDB() {
	dsn := common.DBDSN
	level := logger.Silent
	if klog.V(2).Enabled() {
		level = logger.Info
	}
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      level,
			Colorful:      false,
		},
	)

	var err error
	once.Do(func() {
		cfg := &gorm.Config{
			Logger: newLogger,
		}
		if common.DBType == "sqlite" {
			DB, err = gorm.Open(sqlite.Open(dsn), cfg)
			if err != nil {
				panic("failed to connect database: " + err.Error())
			}
		}

		if common.DBType == "mysql" {
			mysqlDSN := strings.TrimPrefix(dsn, "mysql://")
			if !strings.Contains(mysqlDSN, "parseTime=") {
				separator := "?"
				if strings.Contains(mysqlDSN, "?") {
					separator = "&"
				}
				mysqlDSN = mysqlDSN + separator + "parseTime=true"
			}
			DB, err = gorm.Open(mysql.Open(mysqlDSN), cfg)
			if err != nil {
				panic("failed to connect database: " + err.Error())
			}
		}

		if common.DBType == "postgres" {
			DB, err = gorm.Open(postgres.Open(dsn), cfg)
			if err != nil {
				panic("failed to connect database: " + err.Error())
			}
		}
	})

	if DB == nil {
		panic("database connection is nil, check your DB_TYPE and DB_DSN settings")
	}

	// For SQLite we must enable foreign key enforcement explicitly.
	// SQLite has foreign key constraints defined in the schema but they are
	// not enforced unless PRAGMA foreign_keys = ON is set on the connection.
	if common.DBType == "sqlite" {
		if err := DB.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
			panic("failed to enable sqlite foreign keys: " + err.Error())
		}
	}

	models := []interface{}{
		App{},
		AppConfig{},

		GitlabHosts{},

		User{},
		PersonalAccessToken{},
		UserConfig{},
		UserIdentity{},
		AppUser{},
		UserGitlabConfig{},
		UserAWSConfig{},

		Cluster{},
		OAuthProvider{},
		Role{},
		RoleAssignment{},
		ResourceTemplate{},

		AuditLog{},
		AISettings{},
		ChatSession{},
		ChatMessage{},
	}
	for _, model := range models {
		err = DB.AutoMigrate(model)
		if err != nil {
			panic("failed to migrate database: " + err.Error())
		}
	}

	seedGitlabHosts()

	// Try to get the default app
	app, err := GetApp(common.AppName)
	if err == nil {
		CurrentApp = app
	} else {
		// If not found (or other error), try to create it
		newApp := App{
			Name:    common.AppName,
			Enabled: true,
		}
		if err := DB.Create(&newApp).Error; err != nil {
			panic("failed to create default app: " + err.Error())
		}
		CurrentApp = &newApp
		klog.Infof("Created default app: %s", common.AppName)
	}

	// Initialize default configs if missing
	defaultConfigs := map[string]string{
		DefaultUserAccessKey: "true",
		LocalLoginEnabledKey: "true",
	}
	for key, value := range defaultConfigs {
		if _, err := GetAppConfig(CurrentApp.ID, key); err != nil {
			if err := SetAppConfig(CurrentApp.ID, key, value); err != nil {
				klog.Errorf("Failed to set default config %s: %v", key, err)
			} else {
				klog.Infof("Initialized default config: %s=%s", key, value)
			}
		}
	}

	sqldb, err := DB.DB()
	if err == nil {
		sqldb.SetMaxOpenConns(common.DBMaxOpenConns)
		sqldb.SetMaxIdleConns(common.DBMaxIdleConns)
		sqldb.SetConnMaxLifetime(common.DBMaxIdleTime)
	}
}

// seedGitlabHosts seeds the gitlab_hosts table from the GITLAB_HOSTS environment variable
func seedGitlabHosts() {
	envHosts := os.Getenv("GITLAB_HOSTS")
	if envHosts == "" {
		return
	}

	hosts := strings.Split(envHosts, ",")
	for _, hostStr := range hosts {
		hostStr = strings.TrimSpace(hostStr)
		if hostStr == "" {
			continue
		}

		// Check scheme to determine HTTPS (defaulting to true if no scheme or https)
		isHTTPS := true
		cleanHost := hostStr
		if strings.HasPrefix(hostStr, "http://") {
			isHTTPS = false
			cleanHost = strings.TrimPrefix(hostStr, "http://")
		} else if strings.HasPrefix(hostStr, "https://") {
			isHTTPS = true
			cleanHost = strings.TrimPrefix(hostStr, "https://")
		}

		// Remove any path or query components, keeping just the host
		if idx := strings.Index(cleanHost, "/"); idx != -1 {
			cleanHost = cleanHost[:idx]
		}

		// Only proceed if we have a valid host string
		if cleanHost == "" {
			continue
		}

		hostEntry := GitlabHosts{
			Host:    cleanHost,
			IsHTTPS: &isHTTPS,
		}

		// FirstOrCreate finds by the unique index (Host) or creates a new record
		if err := DB.Model(&GitlabHosts{}).Where("host = ?", cleanHost).FirstOrCreate(&hostEntry).Error; err != nil {
			klog.Errorf("Failed to seed gitlab host %s: %v", cleanHost, err)
		} else {
			klog.Infof("Seeded gitlab host: %s (https=%v)", cleanHost, isHTTPS)
		}
	}
}
