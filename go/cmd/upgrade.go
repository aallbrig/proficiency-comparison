package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade to the latest version",
	Long:  `Check for and install the latest version from GitHub releases.`,
	RunE:  runUpgrade,
}

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	fmt.Println("Checking for updates...")

	// Get latest release
	resp, err := http.Get("https://api.github.com/repos/aallbrig/proficiency-comparison/releases/latest")
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to check for updates: HTTP %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release info: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(Version, "v")

	fmt.Printf("Current version: %s\n", currentVersion)
	fmt.Printf("Latest version:  %s\n", latestVersion)

	if latestVersion == currentVersion {
		fmt.Println("✓ You are already on the latest version!")
		return nil
	}

	// Find appropriate binary for current platform
	osArch := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	var downloadURL string
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, osArch) {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no binary found for platform %s", osArch)
	}

	fmt.Printf("\n📥 Downloading %s...\n", downloadURL)

	// Download new binary
	resp, err = http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	// Get current executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Write to temporary file
	tmpPath := exePath + ".new"
	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write binary: %w", err)
	}

	// Replace old binary
	if err := os.Rename(tmpPath, exePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to replace binary: %w (try running with sudo)", err)
	}

	fmt.Println("✅ Successfully upgraded to version", latestVersion)
	fmt.Println("Run 'edu-stats-cli version' to verify.")

	return nil
}
