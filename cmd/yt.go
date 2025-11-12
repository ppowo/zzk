package cmd

import (
	"fmt"
	"os/exec"
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

func CheckYtDlp() error {
	_, err := exec.LookPath("yt-dlp")
	if err != nil {
		return fmt.Errorf("yt-dlp is not installed. Please install yt-dlp first.\n" +
			"  macOS: brew install yt-dlp\n" +
			"  Linux (Debian/Ubuntu): sudo apt install yt-dlp\n" +
			"  Linux (Fedora): sudo dnf install yt-dlp\n" +
			"  Windows: scoop install yt-dlp or choco install yt-dlp")
	}
	return nil
}

