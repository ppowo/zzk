package git

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// BackupFiles creates a tar.gz archive of the given files
func BackupFiles(files []string, reason string) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("no files to backup")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	backupDir := filepath.Join(homeDir, ".config", "zzk", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", err
	}

	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("git-orphans-%s.tar.gz", timestamp))

	// Create the tar.gz file
	outFile, err := os.Create(backupPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(outFile)
	defer gzipWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Add files to archive
	for _, file := range files {
		if err := addFileToTar(tarWriter, file); err != nil {
			return "", err
		}
	}

	return backupPath, nil
}

// addFileToTar adds a file to the tar archive
func addFileToTar(tarWriter *tar.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	// Use base name for the header
	header.Name = filepath.Base(filename)

	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tarWriter, file)
	return err
}

// RotateBackups keeps only the most recent N backups and deletes older ones
func RotateBackups(dir string, keep int) error {
	pattern := filepath.Join(dir, "git-orphans-*.tar.gz")
	backups, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(backups) <= keep {
		return nil
	}

	// Sort by modification time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		infoI, errI := os.Stat(backups[i])
		infoJ, errJ := os.Stat(backups[j])
		if errI != nil || errJ != nil {
			return false
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})

	// Delete old backups
	for i := keep; i < len(backups); i++ {
		if err := os.Remove(backups[i]); err != nil {
			return err
		}
	}

	return nil
}
