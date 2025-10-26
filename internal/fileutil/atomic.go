package fileutil

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/natefinch/atomic"
)

// AtomicWrite writes data to a file atomically by writing to a temp file first,
// then renaming it to the destination. This ensures the file is never partially written.
func AtomicWrite(dst string, data []byte, perm os.FileMode) error {
	reader := bytes.NewReader(data)
	if err := atomic.WriteFile(dst, reader); err != nil {
		return fmt.Errorf("failed to write file atomically: %w", err)
	}

	// Set correct permissions (atomic.WriteFile doesn't take perm parameter)
	if err := os.Chmod(dst, perm); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	return nil
}

// CopyFile copies a file from src to dst, preserving permissions
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination: %w", err)
	}

	return nil
}
