package models

import (
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"k8s.io/klog/v2"
)

var (
	DB *gorm.DB

	once sync.Once
)

type Model struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
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
		dbType := os.Getenv("DB_TYPE")
		if dbType == "" {
			dbType = "postgres"
		}

		dsn := os.Getenv("DB_DSN")
		if dsn == "" {
			panic("DB_DSN environment variable is required")
		}

		cfg := &gorm.Config{
			Logger: newLogger,
		}

		// Retry logic for database connection
		for i := 0; i < 10; i++ {
			switch dbType {
			case "sqlite":
				DB, err = gorm.Open(sqlite.Open(dsn), cfg)
			case "mysql":
				mysqlDSN := strings.TrimPrefix(dsn, "mysql://")
				if !strings.Contains(mysqlDSN, "parseTime=") {
					separator := "?"
					if strings.Contains(mysqlDSN, "?") {
						separator = "&"
					}
					mysqlDSN = mysqlDSN + separator + "parseTime=true"
				}
				DB, err = gorm.Open(mysql.Open(mysqlDSN), cfg)
			case "postgres":
				DB, err = gorm.Open(postgres.Open(dsn), cfg)
			default:
				panic("unsupported DB_TYPE: " + dbType + " (supported: postgres, mysql, sqlite)")
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
		if dbType == "sqlite" {
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

	// Ensure cloud-sentinel-k8s app exists
	var app App
	result := DB.Where("name = ?", "cloud-sentinel-k8s").First(&app)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			app = App{
				Name:              "cloud-sentinel-k8s",
				Enabled:           true,
				DefaultUserAccess: true,
			}
			if err := DB.Create(&app).Error; err != nil {
				log.Printf("Failed to create cloud-sentinel-k8s app: %v", err)
			} else {
				log.Println("Created cloud-sentinel-k8s app")
			}
		} else {
			log.Printf("Error checking for cloud-sentinel-k8s app: %v", result.Error)
		}
	}
}
