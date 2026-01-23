package model_test

import (
	"testing"

	"github.com/pixelvide/cloud-sentinel-k8s/pkg/common"
	"github.com/pixelvide/cloud-sentinel-k8s/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestAppModel(t *testing.T) {
	// Setup in-memory SQLite DB for testing
	common.DBType = "sqlite"
	common.DBDSN = "file::memory:?cache=shared"
	model.InitDB()

	// Create an App
	app := model.App{Name: "TestApp", Enabled: true}
	err := model.DB.Create(&app).Error
	assert.NoError(t, err)

	// Create an AppConfig
	config := model.AppConfig{AppID: app.ID, Key: "some_key", Value: "some_value"}
	err = model.DB.Create(&config).Error
	assert.NoError(t, err)

	// Query back
	var retrievedConfig model.AppConfig
	err = model.DB.Preload("App").First(&retrievedConfig, config.ID).Error
	assert.NoError(t, err)
	assert.NotNil(t, retrievedConfig.App)
	assert.Equal(t, "TestApp", retrievedConfig.App.Name)
	assert.Equal(t, "some_key", retrievedConfig.Key)
	assert.Equal(t, "some_value", retrievedConfig.Value)
}

func TestInitDB_CreatesDefaultApp(t *testing.T) {
	// Reset DB settings
	common.DBType = "sqlite"
	common.DBDSN = "file::memory:?cache=shared"

	// Note: InitDB uses sync.Once, so we can't easily re-run it in the same process
	// if it was already called. However, provided we are running this test or the previous one,
	// InitDB *should* have been called.
	// If we assume a fresh test run, InitDB will run.
	// Since TestAppModel calls InitDB, subsequent calls are no-ops.
	// But InitDB ensures the default app exists. So we can just check if it's there.

	// If InitDB hasn't run yet in this context (impossible if TestAppModel ran first within same pkg test but possible if run individually), call it.
	model.InitDB()

	var app model.App
	err := model.DB.Where("name = ?", common.AppName).First(&app).Error
	assert.NoError(t, err)
	assert.Equal(t, common.AppName, app.Name)
	assert.True(t, app.Enabled)

	// Verify global variable
	assert.NotNil(t, model.CurrentApp)
	assert.Equal(t, common.AppName, model.CurrentApp.Name)

	// Verify default configs
	config, err := model.GetAppConfig(model.CurrentApp.ID, model.DefaultUserAccessKey)
	assert.NoError(t, err)
	assert.Equal(t, "true", config.Value)

	config, err = model.GetAppConfig(model.CurrentApp.ID, model.LocalLoginEnabledKey)
	assert.NoError(t, err)
	assert.Equal(t, "true", config.Value)
}

func TestInitDB_DoesNotOverwriteExistingConfigs(t *testing.T) {
	// Setup - user already has custom config
	common.DBType = "sqlite"
	common.DBDSN = "file::memory:?cache=shared"
	model.InitDB() // First run creates defaults

	// Change to "false"
	app := model.CurrentApp
	err := model.SetAppConfig(app.ID, model.DefaultUserAccessKey, "false")
	assert.NoError(t, err)

	// Simulate restart
	model.InitDB()

	// Should still be "false"
	config, err := model.GetAppConfig(app.ID, model.DefaultUserAccessKey)
	assert.NoError(t, err)
	assert.Equal(t, "false", config.Value)
}

func TestAppHelpers(t *testing.T) {
	// Setup DB
	common.DBType = "sqlite"
	common.DBDSN = "file::memory:?cache=shared"
	model.InitDB()

	app, err := model.GetApp(common.AppName)
	assert.NoError(t, err)
	assert.Equal(t, common.AppName, app.Name)

	// Set Config
	err = model.SetAppConfig(app.ID, "test_key", "test_value")
	assert.NoError(t, err)

	// Get Config
	config, err := model.GetAppConfig(app.ID, "test_key")
	assert.NoError(t, err)
	assert.Equal(t, "test_value", config.Value)

	// Update Config
	err = model.SetAppConfig(app.ID, "test_key", "updated_value")
	assert.NoError(t, err)

	config, err = model.GetAppConfig(app.ID, "test_key")
	assert.NoError(t, err)
	assert.Equal(t, "updated_value", config.Value)

	// List Configs
	configs, err := model.GetAppConfigs(app.ID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(configs), 1)
}

func TestAppUniqueness(t *testing.T) {
	// Setup DB
	common.DBType = "sqlite"
	common.DBDSN = "file::memory:?cache=shared"
	model.InitDB()

	app1 := model.App{Name: "UniqueApp", Enabled: true}
	err := model.DB.Create(&app1).Error
	assert.NoError(t, err)

	app2 := model.App{Name: "UniqueApp", Enabled: false}
	err = model.DB.Create(&app2).Error
	assert.Error(t, err) // Should fail due to unique constraint
}

func TestIsLocalLoginEnabled(t *testing.T) {
	// Setup DB
	common.DBType = "sqlite"
	common.DBDSN = "file::memory:?cache=shared"
	model.InitDB()

	// Default is true
	assert.True(t, model.IsLocalLoginEnabled())

	// Set to false
	err := model.SetAppConfig(model.CurrentApp.ID, model.LocalLoginEnabledKey, "false")
	assert.NoError(t, err)

	assert.False(t, model.IsLocalLoginEnabled())

	// Set back to true
	err = model.SetAppConfig(model.CurrentApp.ID, model.LocalLoginEnabledKey, "true")
	assert.NoError(t, err)

	assert.True(t, model.IsLocalLoginEnabled())
}

func TestAppUserModel(t *testing.T) {
	// Setup DB
	common.DBType = "sqlite"
	common.DBDSN = "file::memory:?cache=shared"
	model.InitDB()

	// Create App
	app := model.App{Name: "UserApp", Enabled: true}
	err := model.DB.Create(&app).Error
	assert.NoError(t, err)

	// Create User
	user := model.User{
		Username: "testuser_app",
		Password: "hashedpassword",
		Provider: "password",
	}
	err = model.DB.Create(&user).Error
	assert.NoError(t, err)

	// Create AppUser
	appUser := model.AppUser{
		AppID:  app.ID,
		UserID: user.ID,
		Access: true,
	}
	err = model.DB.Create(&appUser).Error
	assert.NoError(t, err)

	// Verify Associations
	var retrievedAppUser model.AppUser
	err = model.DB.Preload("App").Preload("User").First(&retrievedAppUser, appUser.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, app.Name, retrievedAppUser.App.Name)
	assert.Equal(t, user.Username, retrievedAppUser.User.Username)
	assert.True(t, retrievedAppUser.Access)

	// Test Unique Constraint
	duplicateAppUser := model.AppUser{
		AppID:  app.ID,
		UserID: user.ID,
		Access: false,
	}
	err = model.DB.Create(&duplicateAppUser).Error
	assert.Error(t, err)
}

func TestCheckOrInitializeUserAccess(t *testing.T) {
	// Setup DB
	common.DBType = "sqlite"
	common.DBDSN = "file::memory:?cache=shared"
	model.InitDB()

	// Force reset default access to true because other tests might have changed it
	// and InitDB (singelton) reuses the same DB instance.
	_ = model.SetAppConfig(model.CurrentApp.ID, model.DefaultUserAccessKey, "true")

	// Clean up any existing test data from shared DB
	model.DB.Exec("DELETE FROM app_users WHERE user_id IN (SELECT id FROM users WHERE username LIKE 'test_access_user%')")
	model.DB.Exec("DELETE FROM users WHERE username LIKE 'test_access_user%'")

	// Helper to create a user with unique name
	createUser := func(suffix string) uint {
		user := model.User{Username: "test_access_user_" + suffix, Provider: "password"}
		model.DB.Create(&user)
		return user.ID
	}

	// Case 1: Default Access = True (initialized by InitDB)
	user1ID := createUser("user1")
	access, err := model.CheckOrInitializeUserAccess(user1ID)
	assert.NoError(t, err)
	assert.True(t, access)

	// Verify DB entry
	var appUser1 model.AppUser
	err = model.DB.Where("user_id = ?", user1ID).First(&appUser1).Error
	assert.NoError(t, err)
	assert.True(t, appUser1.Access)

	// Case 2: Default Access = False
	err = model.SetAppConfig(model.CurrentApp.ID, model.DefaultUserAccessKey, "false")
	assert.NoError(t, err)

	user2ID := createUser("user2")
	access, err = model.CheckOrInitializeUserAccess(user2ID)
	assert.NoError(t, err)
	assert.False(t, access)

	// Verify DB entry
	var appUser2 model.AppUser
	err = model.DB.Where("user_id = ?", user2ID).First(&appUser2).Error
	assert.NoError(t, err)
	assert.False(t, appUser2.Access)

	// Case 3: Existing User (Access=True), Default=False
	// user1 already has access=true, changing default shouldn't affect them
	access, err = model.CheckOrInitializeUserAccess(user1ID)
	assert.NoError(t, err)
	assert.True(t, access)

	// Case 4: Existing User (Access=False), Default=True
	err = model.SetAppConfig(model.CurrentApp.ID, model.DefaultUserAccessKey, "true")
	assert.NoError(t, err)

	// user2 has access=false, changing default shouldn't affect them
	access, err = model.CheckOrInitializeUserAccess(user2ID)
	assert.NoError(t, err)
	assert.False(t, access)
}
