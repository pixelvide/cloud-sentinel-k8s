package handlers

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/model"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"k8s.io/klog/v2"
)

func setupBenchmarkDB() {
	// Use a unique DB name or ensure we drop tables
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	model.DB = db

	// Drop tables to ensure clean state
	if err := model.DB.Migrator().DropTable(&model.User{}, &model.UserAWSConfig{}, &model.UserConfig{}); err != nil {
		panic("failed to drop tables: " + err.Error())
	}

	err = model.DB.AutoMigrate(&model.User{}, &model.UserAWSConfig{}, &model.UserConfig{})
	if err != nil {
		panic("failed to migrate database: " + err.Error())
	}
}

func BenchmarkRestoreAWSConfigs(b *testing.B) {
	// Silence klog
	klog.SetOutput(io.Discard)

	// Setup temporary directory for data
	tmpDir, err := os.MkdirTemp("", "benchmark-data")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Override DataDir
	utils.DataDir = tmpDir

	setupBenchmarkDB()

	// Seed data
	userCount := 500
	for i := 0; i < userCount; i++ {
		user := model.User{
			Username: fmt.Sprintf("user%d", i),
		}
		if err := model.DB.Create(&user).Error; err != nil {
			b.Fatalf("Failed to create user: %v", err)
		}

		// Create UserConfig (expected to exist)
		userConfig := model.UserConfig{
			UserID:           user.ID,
			StorageNamespace: fmt.Sprintf("ns-%d", i),
		}
		if err := model.DB.Create(&userConfig).Error; err != nil {
			b.Fatalf("Failed to create user config: %v", err)
		}

		// Create UserAWSConfig
		awsConfig := model.UserAWSConfig{
			UserID:             user.ID,
			CredentialsContent: model.SecretString("some-credentials"),
		}
		if err := model.DB.Create(&awsConfig).Error; err != nil {
			b.Fatalf("Failed to create aws config: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RestoreAWSConfigs()
	}
}
