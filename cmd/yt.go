package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var aria2cArgs = []string{
	"--no-netrc=true",
	"--log-level=error",
	"--summary-interval=0",
	"--auto-save-interval=0",
	"--file-allocation=falloc",
	"--console-log-level=error",
	"--split=16",
	"--min-split-size=1M",
	"--http-no-cache=true",
	"--max-connection-per-server=16",
	"--max-overall-download-limit=6M",
}

var baseYtDlpArgs = []string{
	"--geo-bypass",
	"--no-cache-dir",
	"--restrict-filenames",
	"--external-downloader", "aria2c",
}

func GetBaseYtDlpArgs() []string {
	aria2cArgStr := "aria2c:" + strings.Join(aria2cArgs, " ")
	args := make([]string, len(baseYtDlpArgs))
	copy(args, baseYtDlpArgs)
	args = append(args, "--external-downloader-args", aria2cArgStr)
	return args
}

var audioArgs = []string{
	"-o", "%(title)s.%(ext)s",
	"-f", "bestaudio/best",
	"--no-playlist",
}

var albumArgs = []string{
	"-o", "%(uploader,artist|Unknown Artist)s-%(playlist_title)s/%(autonumber)s-%(title)s.%(ext)s",
	"-f", "bestaudio/best",
	"--yes-playlist",
}

var videoArgs = []string{
	"--sub-langs", "en.*",
	"--write-subs",
	"--no-playlist",
	"-o", "%(upload_date)s_%(title)s-[%(id)s].%(ext)s",
}

func GetAudioArgs() []string {
	args := GetBaseYtDlpArgs()
	return append(args, audioArgs...)
}

func GetAlbumArgs() []string {
	args := GetBaseYtDlpArgs()
	return append(args, albumArgs...)
}

func GetScreenHeight() (int, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("system_profiler", "SPDisplaysDataType")
	case "linux":
		cmd = exec.Command("xrandr")
	case "windows":
		cmd = exec.Command("wmic", "path", "Win32_VideoController", "get", "CurrentVerticalResolution")
	default:
		return 0, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get screen resolution: %w", err)
	}

	maxHeight := 0
	outputStr := string(output)

	switch runtime.GOOS {
	case "darwin":
		// Look for "Resolution:" lines in macOS system_profiler output
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Resolution:") {
				// Format: "Resolution: 1920 x 1080"
				parts := strings.Fields(line)
				for i, part := range parts {
					if part == "x" && i+1 < len(parts) {
						if height, err := strconv.Atoi(parts[i+1]); err == nil {
							if height > maxHeight {
								maxHeight = height
							}
						}
					}
				}
			}
		}
	case "linux":
		// Parse xrandr output - look for lines like "1920x1080"
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "x") && strings.Contains(line, "+") {
				parts := strings.Fields(line)
				if len(parts) > 0 {
					resParts := strings.Split(parts[0], "x")
					if len(resParts) == 2 {
						if height, err := strconv.Atoi(resParts[1]); err == nil {
							if height > maxHeight {
								maxHeight = height
							}
						}
					}
				}
			}
		}
	case "windows":
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && line != "CurrentVerticalResolution" {
				if height, err := strconv.Atoi(line); err == nil {
					if height > maxHeight {
						maxHeight = height
					}
				}
			}
		}
	}
	if maxHeight == 0 {
		return 0, fmt.Errorf("could not detect screen resolution")
	}
	return maxHeight, nil
}

func GetVideoArgs() ([]string, error) {
	args := GetBaseYtDlpArgs()
	maxHeight, err := GetScreenHeight()
	if err != nil {
		return nil, err
	}
	qualityStr := fmt.Sprintf("bestvideo[height<=%d]+bestaudio/best[height<=%d]/best", maxHeight, maxHeight)
	args = append(args, videoArgs...)
	args = append(args, "-f", qualityStr)
	return args, nil
}

var ytCmd = &cobra.Command{
	Use:   "yt",
	Short: "YouTube download operations using yt-dlp",
	Long:  `Parent command for YouTube download operations. Use subcommands to perform actions.`,
}

func init() {
	rootCmd.AddCommand(ytCmd)
}

func CheckAria2c() error {
	_, err := exec.LookPath("aria2c")
	if err != nil {
		return fmt.Errorf("aria2c is not installed. Please install aria2c first.\n" +
			"  macOS: brew install aria2\n" +
			"  Linux (Debian/Ubuntu): sudo apt install aria2\n" +
			"  Linux (Fedora): sudo dnf install aria2\n" +
			"  Windows: scoop install aria2 or choco install aria2")
	}
	return nil
}

func EnsureYtDlp() error {
	binPath := GetYtDlpPath()

	// Check if yt-dlp exists
	if _, err := os.Stat(binPath); err == nil {
		// Check if it needs updating
		if needsUpdate, err := checkYtDlpUpdate(binPath); err == nil && needsUpdate {
			fmt.Println("yt-dlp update available, downloading...")
			return downloadYtDlp(binPath)
		}
		return nil
	}

	// yt-dlp not found, download it
	fmt.Println("yt-dlp not found, downloading...")
	return downloadYtDlp(binPath)
}

func checkYtDlpUpdate(binPath string) (bool, error) {
	// Run yt-dlp --version to get current version
	cmd := exec.Command(binPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		// If we can't get version, assume it needs update
		return true, nil
	}

	currentVersion := strings.TrimSpace(string(output))

	// Check latest version from GitHub API
	resp, err := http.Get("https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest")
	if err != nil {
		// If we can't check, don't update
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to check latest version: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	// Simple JSON parsing - look for "tag_name"
	bodyStr := string(body)
	tagStart := strings.Index(bodyStr, `"tag_name":"`)
	if tagStart == -1 {
		return false, fmt.Errorf("could not parse latest version")
	}
	tagStart += len(`"tag_name":"`)
	tagEnd := strings.Index(bodyStr[tagStart:], `"`)
	if tagEnd == -1 {
		return false, fmt.Errorf("could not parse latest version")
	}
	latestVersion := bodyStr[tagStart : tagStart+tagEnd]

	// yt-dlp --version outputs just the date (e.g., "2025.09.26")
	// GitHub tag_name is the same format
	currentVersion = strings.TrimSpace(currentVersion)
	latestVersion = strings.TrimSpace(latestVersion)

	// Compare versions
	return currentVersion != latestVersion, nil
}

func downloadYtDlp(binPath string) error {
	url := getYtDlpURL()
	if url == "" {
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	zzkDir := filepath.Dir(binPath)
	if err := os.MkdirAll(zzkDir, 0755); err != nil {
		return fmt.Errorf("failed to create zzk directory: %w", err)
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	out, err := os.Create(binPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(binPath, 0755); err != nil {
			return fmt.Errorf("failed to make executable: %w", err)
		}
	}

	fmt.Printf("✓ Downloaded yt-dlp to: %s\n", binPath)
	return nil
}

func GetYtDlpPath() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	zzkDir := filepath.Join(cacheDir, "zzk")
	if runtime.GOOS == "windows" {
		return filepath.Join(zzkDir, "yt-dlp.exe")
	}
	return filepath.Join(zzkDir, "yt-dlp")
}

func getYtDlpURL() string {
	baseURL := "https://github.com/yt-dlp/yt-dlp/releases/latest/download/"
	switch runtime.GOOS {
	case "linux":
		return baseURL + "yt-dlp_linux"
	case "darwin":
		return baseURL + "yt-dlp_macos"
	case "windows":
		return baseURL + "yt-dlp.exe"
	default:
		return ""
	}
}
