package models

import (
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"cloud-sentinel-k8s/pkg/common"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"k8s.io/klog/v2"
)

var (
	DB *gorm.DB

	once sync.Once

	// GlobalApp holds the single instance of the application configuration
	GlobalApp *App
)

type Model struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func InitDB() {
	// Configure GORM logger based on klog verbosity
	level := logger.Silent
	if klog.V(2).Enabled() {
		level = logger.Info
	}
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      level,
			Colorful:      false,
		},
	)

	var err error
	once.Do(func() {
		if common.DBDSN == "" {
			panic("DB_DSN environment variable is required")
		}

		cfg := &gorm.Config{
			Logger: newLogger,
		}

		// Retry logic for database connection
		for i := 0; i < 10; i++ {
			switch common.DBType {
			case "sqlite":
				DB, err = gorm.Open(sqlite.Open(common.DBDSN), cfg)
			case "mysql":
				mysqlDSN := strings.TrimPrefix(common.DBDSN, "mysql://")
				if !strings.Contains(mysqlDSN, "parseTime=") {
					separator := "?"
					if strings.Contains(mysqlDSN, "?") {
						separator = "&"
					}
					mysqlDSN = mysqlDSN + separator + "parseTime=true"
				}
				DB, err = gorm.Open(mysql.Open(mysqlDSN), cfg)
			case "postgres":
				DB, err = gorm.Open(postgres.Open(common.DBDSN), cfg)
			default:
				panic("unsupported DB_TYPE: " + common.DBType + " (supported: postgres, mysql, sqlite)")
			}

			if err == nil {
				break
			}
			log.Printf("Failed to connect to database: %v. Retrying in 2 seconds...", err)
			time.Sleep(2 * time.Second)
		}

		if err != nil {
			panic("Failed to connect to database after retries: " + err.Error())
		}

		// For SQLite we must enable foreign key enforcement explicitly
		if common.DBType == "sqlite" {
			if err := DB.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
				panic("failed to enable sqlite foreign keys: " + err.Error())
			}
		}
	})

	if DB == nil {
		panic("database connection is nil")
	}

	log.Println("Database connection established")

	// Auto-migrate all models
	models := []interface{}{
		&App{},
		&AppConfig{},
		&AppUser{},
		&OAuthProvider{},
		&User{},
		&UserIdentity{},
		&AuditLog{},
		&GitlabConfig{},
		&GitlabK8sAgentConfig{},
		&K8sClusterContextMapping{},
		&KubeConfig{},
		&AWSConfig{},
		&EKSCluster{},
	}

	for _, model := range models {
		err = DB.AutoMigrate(model)
		if err != nil {
			panic("failed to migrate database: " + err.Error())
		}
	}
	log.Println("Database migration completed")

	// Ensure default app exists
	var app App
	result := DB.Where("name = ?", common.AppName).First(&app)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			app = App{
				Name:              common.AppName,
				Enabled:           true,
				DefaultUserAccess: true,
			}
			if err := DB.Create(&app).Error; err != nil {
				log.Printf("Failed to create %s app: %v", common.AppName, err)
			} else {
				log.Printf("Created %s app", common.AppName)
			}
		} else {
			log.Printf("Error checking for %s app: %v", common.AppName, result.Error)
		}
	}

	GlobalApp = &app

	// Ensure default app configs exist
	if app.ID != 0 {
		// LOCAL_LOGIN_ENABLED
		if _, err := GetAppConfig("LOCAL_LOGIN_ENABLED"); err != nil {
			if err == gorm.ErrRecordNotFound {
				config := AppConfig{
					// AppID is set inside CreateAppConfig using GlobalApp
					Key:   "LOCAL_LOGIN_ENABLED",
					Value: "true",
				}
				if err := CreateAppConfig(&config); err != nil {
					log.Printf("Failed to create LOCAL_LOGIN_ENABLED config: %v", err)
				} else {
					log.Printf("Seeded LOCAL_LOGIN_ENABLED=true for app %s", app.Name)
				}
			}
		}
	}

	initOIDCProvider()
}

func initOIDCProvider() {
	issuer := os.Getenv("OIDC_ISSUER")
	if issuer == "" {
		return
	}

	clientID := os.Getenv("OIDC_CLIENT_ID")
	clientSecret := os.Getenv("OIDC_CLIENT_SECRET")

	var count int64
	DB.Model(&OAuthProvider{}).Where("name = ?", "oidc").Count(&count)
	if count == 0 {
		log.Println("Persisting OIDC configuration from environment to database...")
		newProvider := OAuthProvider{
			Name:         "oidc",
			Issuer:       issuer,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Enabled:      true,
		}
		if err := CreateOAuthProvider(&newProvider); err != nil {
			log.Printf("Failed to persist OIDC provider to DB: %v", err)
		}
	}
}
