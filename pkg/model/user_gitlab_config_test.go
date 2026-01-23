package model

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func cleanupUserGitlabConfig(db *gorm.DB) {
	db.Exec("DELETE FROM user_gitlab_configs")
	db.Exec("DELETE FROM gitlab_hosts")
	db.Exec("DELETE FROM users")
}

func TestUserGitlabConfig(t *testing.T) {
	// Setup in-memory SQLite DB for testing
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Enable foreign keys
	db.Exec("PRAGMA foreign_keys = ON")

	err = db.AutoMigrate(&User{}, &GitlabHosts{}, &UserGitlabConfig{})
	assert.NoError(t, err)

	// Set the global DB to our test DB (be careful if tests run in parallel,
	// but here we are just testing the model logic in isolation effectively if we pass db explicitly,
	// however model methods often use the global DB.
	// Since we are testing constraints, we can use the local db instance mostly,
	// but let's see if we need to mock the global DB.
	// For this test, we can just use the db instance directly.)

	t.Run("Create UserGitlabConfig", func(t *testing.T) {
		cleanupUserGitlabConfig(db)
		defer cleanupUserGitlabConfig(db)

		user := User{Username: "testuser_gitlab", Password: "password"}
		err := db.Create(&user).Error
		assert.NoError(t, err)

		host := GitlabHosts{Host: "gitlab.com"}
		err = db.Create(&host).Error
		assert.NoError(t, err)

		config := UserGitlabConfig{
			UserID:       user.ID,
			GitlabHostID: host.ID,
			Token:        "supersecretToken",
		}
		err = db.Create(&config).Error
		assert.NoError(t, err)
		assert.NotZero(t, config.ID)
	})

	t.Run("Unique Constraint User/Host", func(t *testing.T) {
		cleanupUserGitlabConfig(db)

		user := User{Username: "testuser_gitlab", Password: "password"}
		db.Create(&user)
		host := GitlabHosts{Host: "gitlab.com"}
		db.Create(&host)

		config1 := UserGitlabConfig{UserID: user.ID, GitlabHostID: host.ID, Token: "token1"}
		db.Create(&config1)

		config2 := UserGitlabConfig{UserID: user.ID, GitlabHostID: host.ID, Token: "token2"}
		err := db.Create(&config2).Error
		assert.Error(t, err) // Should fail due to unique index
	})

	t.Run("Foreign Key Cascade - User", func(t *testing.T) {
		cleanupUserGitlabConfig(db)

		user := User{Username: "testuser_gitlab", Password: "password"}
		db.Create(&user)
		host := GitlabHosts{Host: "gitlab.com"}
		db.Create(&host)

		config := UserGitlabConfig{UserID: user.ID, GitlabHostID: host.ID, Token: "token1"}
		db.Create(&config)

		// Delete user
		err := db.Delete(&user, user.ID).Error
		assert.NoError(t, err)

		// Config should be gone
		var count int64
		db.Model(&UserGitlabConfig{}).Where("id = ?", config.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("Foreign Key Cascade - Host", func(t *testing.T) {
		cleanupUserGitlabConfig(db)

		user := User{Username: "testuser_gitlab", Password: "password"}
		db.Create(&user)
		host := GitlabHosts{Host: "gitlab.com"}
		db.Create(&host)

		config := UserGitlabConfig{UserID: user.ID, GitlabHostID: host.ID, Token: "token1"}
		db.Create(&config)

		// Delete host
		err := db.Delete(&host, host.ID).Error
		assert.NoError(t, err)

		// Config should be gone
		var count int64
		db.Model(&UserGitlabConfig{}).Where("id = ?", config.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})
}
