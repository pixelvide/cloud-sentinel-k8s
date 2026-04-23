package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSecureFilePermissions(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "kube-sentinel-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Override DataDir
	originalDataDir := DataDir
	DataDir = tempDir
	defer func() { DataDir = originalDataDir }()

	// Clean up temp dir
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	t.Run("GetUserGlabConfigDir", func(t *testing.T) {
		path, err := GetUserGlabConfigDir("test-namespace")
		if err != nil {
			t.Fatalf("GetUserGlabConfigDir failed: %v", err)
		}

		// Check directory permissions
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Failed to stat path: %v", err)
		}
		mode := info.Mode().Perm()
		if mode != 0700 {
			t.Errorf("Expected directory permissions 0700, got %04o", mode)
		}
	})

	t.Run("WriteUserAWSCredentials", func(t *testing.T) {
		err := WriteUserAWSCredentials("test-namespace", "aws-creds")
		if err != nil {
			t.Fatalf("WriteUserAWSCredentials failed: %v", err)
		}

		path := GetUserAWSCredentialsPath("test-namespace")

		// Check file permissions
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Failed to stat file: %v", err)
		}
		mode := info.Mode().Perm()
		if mode != 0600 {
			t.Errorf("Expected file permissions 0600, got %04o", mode)
		}

		// Check parent directory permissions
		dir := filepath.Dir(path)
		info, err = os.Stat(dir)
		if err != nil {
			t.Fatalf("Failed to stat directory: %v", err)
		}
		mode = info.Mode().Perm()
		if mode != 0700 {
			t.Errorf("Expected parent directory permissions 0700, got %04o", mode)
		}
	})
}
