package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"
)

func restoreBackup(target BackupTarget, code string) error {
	timestamp := time.Now().Format("2006-01-02 15:04")
	fmt.Printf("%s - Starting %s restore from code: %s\n", timestamp, target.Name, code)

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	targetPath := filepath.Join(home, target.Path)
	url := fmt.Sprintf("%s/%s.tar.xz", backupServiceURL, code)

	// Download to /tmp first for validation
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("%s-restore-*.tar.xz", target.Name))
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpArchive := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpArchive)

	fmt.Printf("%s - Downloading...\n", time.Now().Format("2006-01-02 15:04"))

	curlCmd := exec.Command("curl", "-sL", "-A", "zzk-backup/1.0", "-o", tmpArchive, url)
	if err := curlCmd.Run(); err != nil {
		return fmt.Errorf("failed to download archive: %w", err)
	}

	// Verify it's a valid tar.xz (not HTML error page)
	fmt.Printf("%s - Verifying downloaded archive...\n", time.Now().Format("2006-01-02 15:04"))
	if err := verifyTarXz(tmpArchive); err != nil {
		return fmt.Errorf("downloaded file is not a valid tar.xz archive: %w\nYou may have entered the wrong code or the file may have expired", err)
	}

	// Get archive size
	stat, err := os.Stat(tmpArchive)
	if err != nil {
		return fmt.Errorf("failed to stat archive: %w", err)
	}
	sizeMB := float64(stat.Size()) / (1024 * 1024)
	fmt.Printf("%s - Archive verified (size: %.2f MB)\n", time.Now().Format("2006-01-02 15:04"), sizeMB)

	// Test extraction to /tmp to ensure archive is not corrupted
	testDir, err := os.MkdirTemp("", fmt.Sprintf("%s-test-*", target.Name))
	if err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}
	defer os.RemoveAll(testDir)

	fmt.Printf("%s - Testing archive extraction...\n", time.Now().Format("2006-01-02 15:04"))
	testCmd := exec.Command("tar", "-xJf", tmpArchive, "-C", testDir)
	if output, err := testCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("archive extraction test failed: %w\n%s", err, output)
	}

	// Verify target directory exists in extracted content
	testTargetPath := filepath.Join(testDir, target.Path)
	if _, err := os.Stat(testTargetPath); os.IsNotExist(err) {
		return fmt.Errorf("archive does not contain a %s directory", target.Path)
	}

	fmt.Printf("%s - Archive test successful\n", time.Now().Format("2006-01-02 15:04"))

	// Backup existing target if it exists
	var existingBackup string
	if _, err := os.Stat(targetPath); err == nil {
		fmt.Printf("%s - Existing %s directory found, creating backup...\n", time.Now().Format("2006-01-02 15:04"), target.Name)

		timestamp := time.Now().Format("20060102-150405")
		existingBackup = filepath.Join(home, fmt.Sprintf("%s%s", target.BackupPrefix, timestamp))

		if err := os.Rename(targetPath, existingBackup); err != nil {
			return fmt.Errorf("failed to backup existing %s: %w", target.Name, err)
		}
		fmt.Printf("%s - Backup created at %s\n", time.Now().Format("2006-01-02 15:04"), existingBackup)

		// Clean up old backups, keep only last N
		if err := cleanupOldBackups(home, target.BackupPrefix, target.KeepBackups); err != nil {
			fmt.Printf("%s - Warning: failed to cleanup old backups: %v\n", time.Now().Format("2006-01-02 15:04"), err)
		}
	}

	// Extract to home directory
	fmt.Printf("%s - Extracting archive to %s...\n", time.Now().Format("2006-01-02 15:04"), home)
	extractCmd := exec.Command("tar", "-xJf", tmpArchive, "-C", home)
	if output, err := extractCmd.CombinedOutput(); err != nil {
		// If extraction failed and we made a backup, try to restore it
		if existingBackup != "" {
			fmt.Printf("%s - Extraction failed, restoring backup...\n", time.Now().Format("2006-01-02 15:04"))
			os.Rename(existingBackup, targetPath)
		}
		return fmt.Errorf("failed to extract archive: %w\n%s", err, output)
	}

	fmt.Printf("%s - %s restored successfully!\n", time.Now().Format("2006-01-02 15:04"), target.Name)
	if existingBackup != "" {
		fmt.Printf("%s - Previous %s backed up to: %s\n", time.Now().Format("2006-01-02 15:04"), target.Name, existingBackup)
	}
	fmt.Printf("%s - Temporary archive removed.\n", time.Now().Format("2006-01-02 15:04"))

	return nil
}

// cleanupOldBackups removes old backup directories, keeping only the most recent N
func cleanupOldBackups(homeDir string, backupPrefix string, keepCount int) error {
	// Find all backup directories with the given prefix
	pattern := filepath.Join(homeDir, backupPrefix+"*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to find backup directories: %w", err)
	}

	// If we have fewer backups than keepCount, nothing to do
	if len(matches) <= keepCount {
		return nil
	}

	// Sort by modification time (newest first)
	type backupInfo struct {
		path    string
		modTime time.Time
	}

	backups := make([]backupInfo, 0, len(matches))
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue // Skip if we can't stat it
		}
		backups = append(backups, backupInfo{path: match, modTime: info.ModTime()})
	}

	// Sort by modification time, newest first
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].modTime.After(backups[j].modTime)
	})

	// Remove backups beyond keepCount
	for i := keepCount; i < len(backups); i++ {
		if err := os.RemoveAll(backups[i].path); err != nil {
			fmt.Printf("%s - Warning: failed to remove old backup %s: %v\n",
				time.Now().Format("2006-01-02 15:04"), backups[i].path, err)
		} else {
			fmt.Printf("%s - Removed old backup: %s\n",
				time.Now().Format("2006-01-02 15:04"), filepath.Base(backups[i].path))
		}
	}

	return nil
}
