package utils

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/rand"
)

func IsSecureRegistry(host string) error {
	hostName := host
	// Remove port if present
	if strings.Contains(host, ":") {
		h, _, err := net.SplitHostPort(host)
		if err == nil {
			hostName = h
		}
	}

	// Handle IPv6 literal
	if strings.HasPrefix(hostName, "[") && strings.HasSuffix(hostName, "]") {
		hostName = hostName[1 : len(hostName)-1]
	}

	if strings.ToLower(hostName) == "localhost" {
		return errors.New("access to localhost is restricted")
	}

	ip := net.ParseIP(hostName)
	if ip != nil {
		if ip.IsLoopback() || ip.IsUnspecified() {
			return errors.New("access to loopback/unspecified addresses is restricted")
		}
	}

	return nil
}

func InjectKubeSentinelBase(htmlContent string, base string) string {
	baseScript := fmt.Sprintf(`<script>window.__dynamic_base__='%s';</script>`, base)
	re := regexp.MustCompile(`<head>`)
	return re.ReplaceAllString(htmlContent, "<head>\n    "+baseScript)
}

func RandomString(length int) string {
	return rand.String(length)
}

func ToEnvName(input string) string {
	s := input
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ToUpper(s)
	return s
}

func GetImageRegistryAndRepo(image string) (string, string) {
	image = strings.SplitN(image, ":", 2)[0]
	parts := strings.Split(image, "/")
	if len(parts) == 1 {
		return "", "library/" + parts[0]
	}
	if len(parts) > 1 {
		if strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":") {
			return parts[0], strings.Join(parts[1:], "/")
		}
		return "", strings.Join(parts, "/")
	}
	return "", image
}

var DataDir = "/data"

// GetUserGlabConfigDir returns the directory path for a user's glab configuration.
// It ensures the directory exists and has 0700 permissions.
func GetUserGlabConfigDir(storageNamespace string) (string, error) {
	path := filepath.Join(DataDir, storageNamespace, ".config", "glab-cli")
	if err := os.MkdirAll(path, 0700); err != nil {
		return "", fmt.Errorf("failed to create glab config directory: %w", err)
	}
	// Explicitly chmod to ensure permissions are correct regardless of umask
	if err := os.Chmod(path, 0700); err != nil {
		return "", fmt.Errorf("failed to chmod glab config directory: %w", err)
	}
	return path, nil
}

func GlabAuthLogin(host, token, configDir string) error {
	loginCmd := exec.Command("glab", "auth", "login", "--hostname", host, "--token", token)
	loginCmd.Env = append(os.Environ(), "GLAB_CONFIG_DIR="+configDir)
	if output, err := loginCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("login failed for %s: %s", host, string(output))
	}
	return nil
}

// GetUserAWSCredentialsPath returns the path to the user's AWS credentials file.
func GetUserAWSCredentialsPath(storageNamespace string) string {
	return filepath.Join(DataDir, storageNamespace, ".config", "aws", "credentials")
}

// WriteUserAWSCredentials writes the AWS credentials content to the user's specific path.
func WriteUserAWSCredentials(storageNamespace string, content string) error {
	path := GetUserAWSCredentialsPath(storageNamespace)
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create aws config directory: %w", err)
	}
	// Explicitly chmod to ensure permissions are correct regardless of umask
	if err := os.Chmod(dir, 0700); err != nil {
		return fmt.Errorf("failed to chmod aws config directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write aws credentials file: %w", err)
	}

	if err := os.Chmod(path, 0600); err != nil {
		return fmt.Errorf("failed to chmod aws credentials file: %w", err)
	}

	return nil
}

func ContainsString(slice []string, val string) bool {
	for _, s := range slice {
		if strings.ToLower(s) == val {
			return true
		}
	}
	return false
}
