package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	appVersion    = "dev"
	latestVersion string
	lastChecked   time.Time
	checkMutex    sync.Mutex
	checkInterval = 1 * time.Hour
	githubRepo    = "pixelvide/cloud-sentinel-k8s" // Correct repo
)

type VersionResponse struct {
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version"`
	UpdateAvailable bool   `json:"update_available"`
}

func SetAppVersion(v string) {
	appVersion = v
}

func GetAppVersion(c *gin.Context) {
	checkMutex.Lock()
	defer checkMutex.Unlock()

	autoCheck := os.Getenv("AUTO_CHECK_FOR_UPDATES")
	if autoCheck != "false" { // Default to true if not set to "false"
		if time.Since(lastChecked) > checkInterval || latestVersion == "" {
			fetchLatestVersion()
		}
	} else {
		latestVersion = "" // Ensure latest version is empty if disabled
	}

	c.JSON(http.StatusOK, VersionResponse{
		CurrentVersion:  appVersion,
		LatestVersion:   latestVersion,
		UpdateAvailable: isUpdateAvailable(appVersion, latestVersion),
	})
}

func fetchLatestVersion() {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo))
	if err != nil {
		fmt.Printf("Error fetching latest version: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("GitHub API returned status: %s\n", resp.Status)
		return
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fmt.Printf("Error decoding GitHub response: %v\n", err)
		return
	}

	latestVersion = release.TagName
	lastChecked = time.Now()
}

func isUpdateAvailable(current, latest string) bool {
	if current == "dev" || latest == "" {
		return false
	}
	// Simple string comparison for now, can be improved with semver logic
	// Remove 'v' prefix if present
	if len(current) > 0 && current[0] == 'v' {
		current = current[1:]
	}
	if len(latest) > 0 && latest[0] == 'v' {
		latest = latest[1:]
	}
	return current != latest && latest != ""
}
