package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func uploadBackup(target BackupTarget) error {
	timestamp := time.Now().Format("2006-01-02 15:04")
	fmt.Printf("%s - Starting %s backup\n", timestamp, target.Name)
	fmt.Printf("This will archive your ~/%s and upload it for backup/sharing\n", target.Path)
	fmt.Println()

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	targetPath := filepath.Join(home, target.Path)
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("%s directory not found at %s", target.Name, targetPath)
	}

	fmt.Printf("%s - Found %s directory at %s\n", time.Now().Format("2006-01-02 15:04"), target.Name, targetPath)

	// Create temporary archive
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("%s-backup-*.tar.xz", target.Name))
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpArchive := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpArchive)

	// Build tar command
	tarArgs := []string{"-cJf", tmpArchive}
	for _, pattern := range globalExcludeGlobs {
		tarArgs = append(tarArgs, "--exclude", pattern)
	}
	tarArgs = append(tarArgs, target.Path)

	fmt.Printf("%s - Creating compressed archive...\n", time.Now().Format("2006-01-02 15:04"))

	cmd := exec.Command("tar", tarArgs...)
	cmd.Dir = home
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create archive: %w\n%s", err, output)
	}

	// Get archive size
	stat, err := os.Stat(tmpArchive)
	if err != nil {
		return fmt.Errorf("failed to stat archive: %w", err)
	}
	sizeMB := float64(stat.Size()) / (1024 * 1024)
	fmt.Printf("%s - Archive created successfully (size: %.2f MB)\n", time.Now().Format("2006-01-02 15:04"), sizeMB)

	// Upload
	fmt.Printf("%s - Uploading...\n", time.Now().Format("2006-01-02 15:04"))

	curlCmd := exec.Command("curl", "-s", "-A", "zzk-backup/1.0", "-F", fmt.Sprintf("file=@%s", tmpArchive), backupServiceURL)
	output, err := curlCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}

	url := cleanURL(string(output))
	if url == "" {
		return fmt.Errorf("upload failed: empty response")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("upload failed: invalid URL response: %s", url)
	}

	// Verify upload by downloading to /tmp and checking it's a valid tar.xz
	fmt.Printf("%s - Verifying upload...\n", time.Now().Format("2006-01-02 15:04"))

	verifyFile, err := os.CreateTemp("", fmt.Sprintf("%s-verify-*.tar.xz", target.Name))
	if err != nil {
		return fmt.Errorf("failed to create verification temp file: %w", err)
	}
	verifyPath := verifyFile.Name()
	verifyFile.Close()
	defer os.Remove(verifyPath)

	// Download the uploaded file
	curlDownload := exec.Command("curl", "-sL", "-A", "zzk-backup/1.0", "-o", verifyPath, url)
	if err := curlDownload.Run(); err != nil {
		return fmt.Errorf("failed to download for verification: %w", err)
	}

	// Check if it's a valid tar.xz file (not HTML)
	if err := verifyTarXz(verifyPath); err != nil {
		return fmt.Errorf("upload verification failed: %w\nReceived file may be an error page instead of archive", err)
	}

	fmt.Printf("%s - Upload verified successfully!\n", time.Now().Format("2006-01-02 15:04"))
	fmt.Printf("%s - Your %s backup is available at:\n", time.Now().Format("2006-01-02 15:04"), target.Name)
	fmt.Println(url)

	// Extract code from URL
	code := strings.TrimSuffix(filepath.Base(url), ".tar.xz")
	fmt.Printf("%s - Restore with: zzk backup %s %s\n", time.Now().Format("2006-01-02 15:04"), target.Name, code)
	fmt.Printf("%s - Temporary archive removed.\n", time.Now().Format("2006-01-02 15:04"))

	return nil
}

// cleanURL removes control characters from a URL string returned by the upload service.
// This handles cases where the response includes trailing newlines, carriage returns,
// or other control characters (0x00-0x1F) that are invalid in URLs.
func cleanURL(s string) string {
	var result strings.Builder
	result.Grow(len(s)) // Pre-allocate for efficiency

	for _, r := range s {
		// Keep printable ASCII (0x20 space through 0x7E tilde)
		// This includes spaces, letters, numbers, and URL-safe punctuation
		if r >= 32 && r <= 126 {
			result.WriteRune(r)
		}
	}

	// Trim any resulting spaces from edges (in case spaces were at boundaries)
	return strings.TrimSpace(result.String())
}

// verifyTarXz checks if a file is a valid tar.xz archive
func verifyTarXz(path string) error {
	// Check file magic bytes for XZ format
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	magic := make([]byte, 6)
	n, err := f.Read(magic)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	if n < 6 {
		return fmt.Errorf("file too small to be a valid tar.xz")
	}

	// XZ files start with 0xFD 0x37 0x7A 0x58 0x5A 0x00
	if magic[0] != 0xFD || magic[1] != 0x37 || magic[2] != 0x7A ||
		magic[3] != 0x58 || magic[4] != 0x5A || magic[5] != 0x00 {
		return fmt.Errorf("file is not a valid XZ archive (wrong magic bytes)")
	}

	// Try to list contents with tar
	cmd := exec.Command("tar", "-tJf", path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to verify tar archive: %w", err)
	}

	return nil
}
