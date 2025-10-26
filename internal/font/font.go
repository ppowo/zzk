package font

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// GetUserFontDir returns the user-level font directory for the current OS.
// Does not require admin/sudo privileges.
//
// Returns:
//   - macOS: ~/Library/Fonts
//   - Linux: ~/.local/share/fonts
//   - Windows: %LOCALAPPDATA%\Microsoft\Windows\Fonts
func GetUserFontDir() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Library/Fonts
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Fonts"), nil
	case "linux":
		// Linux: ~/.local/share/fonts
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "share", "fonts"), nil
	case "windows":
		// Windows: %LOCALAPPDATA%\Microsoft\Windows\Fonts
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return "", fmt.Errorf("LOCALAPPDATA environment variable not set")
		}
		return filepath.Join(localAppData, "Microsoft", "Windows", "Fonts"), nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// RefreshFontCache attempts to refresh the system font cache.
// This is a best-effort operation and will not return an error if it fails.
// Platform-specific behavior:
//   - Linux: Runs fc-cache if available
//   - macOS: Runs fc-cache if available (usually not needed)
//   - Windows: No action needed (automatic)
func RefreshFontCache() {
	switch runtime.GOOS {
	case "linux":
		// Try fc-cache on Linux
		cmd := exec.Command("fc-cache", "-f", "-v")
		_ = cmd.Run() // Ignore errors
	case "darwin":
		// macOS automatically updates font cache, but we can try fc-cache if available
		cmd := exec.Command("fc-cache", "-f", "-v")
		_ = cmd.Run() // Ignore errors
	case "windows":
		// Windows automatically updates font cache when files are added
		// No action needed
	}
}
