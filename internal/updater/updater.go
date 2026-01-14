package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/janyksteenbeek/uppi-server-agent/internal/config"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// CheckForUpdates checks GitHub for new releases and updates if available
func CheckForUpdates() {
	log.Println("Checking for updates...")

	currentVersion := config.Version

	latestRelease, err := getLatestRelease()
	if err != nil {
		log.Printf("Failed to check for updates: %v", err)
		return
	}

	if isNewerVersion(latestRelease.TagName, currentVersion) {
		log.Printf("New version available: %s (current: %s)", latestRelease.TagName, currentVersion)

		if err := performUpdate(latestRelease); err != nil {
			log.Printf("Failed to update: %v", err)
			return
		}

		log.Println("Update completed successfully. Restarting...")
		restartAgent()
	} else {
		log.Println("Already running the latest version")
	}
}

func getLatestRelease() (*GitHubRelease, error) {
	resp, err := http.Get("https://api.github.com/repos/janyksteenbeek/uppi-server-agent/releases/latest")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func isNewerVersion(latest, current string) bool {
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")
	return latest != current
}

func performUpdate(release *GitHubRelease) error {
	goos := runtime.GOOS
	arch := runtime.GOARCH
	assetName := fmt.Sprintf("uppi-agent-%s-%s", goos, arch)

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no asset found for architecture %s-%s", goos, arch)
	}

	tempFile := "/tmp/uppi-agent-new"
	if err := downloadFile(downloadURL, tempFile); err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}

	if err := os.Chmod(tempFile, 0755); err != nil {
		return fmt.Errorf("failed to make new binary executable: %w", err)
	}

	currentPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	backupPath := currentPath + ".backup"
	if err := copyFile(currentPath, backupPath); err != nil {
		log.Printf("Warning: failed to create backup: %v", err)
	}

	if err := copyFile(tempFile, currentPath); err != nil {
		if backupErr := copyFile(backupPath, currentPath); backupErr != nil {
			log.Printf("Critical: failed to restore backup after failed update: %v", backupErr)
		}
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	os.Remove(tempFile)
	os.Remove(backupPath)

	return nil
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

func restartAgent() {
	if isSystemdService() {
		cmd := exec.Command("systemctl", "restart", "uppi-agent")
		cmd.Run()
		return
	}

	args := os.Args
	if err := exec.Command(args[0], args[1:]...).Start(); err != nil {
		log.Printf("Failed to restart: %v", err)
	}

	go func() {
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()
}

func isSystemdService() bool {
	return os.Getenv("SYSTEMD_EXEC_PID") != "" || os.Getppid() == 1
}
