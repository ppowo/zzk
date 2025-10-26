package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DetectIdentity detects which identity applies to the given directory
func DetectIdentity(config *Config, dir string) (*Identity, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	absDir = filepath.Clean(absDir)

	for _, identity := range config.Identities {
		for _, folder := range identity.Folders {
			expandedFolder := expandPath(folder)
			absFolder, err := filepath.Abs(expandedFolder)
			if err != nil {
				continue
			}

			absFolder = filepath.Clean(absFolder)

			if strings.HasPrefix(absDir, absFolder) {
				identityCopy := identity
				return &identityCopy, nil
			}
		}
	}

	return nil, fmt.Errorf("no identity found for directory: %s", dir)
}

// GetCurrentIdentity gets the identity for the current working directory
func GetCurrentIdentity(config *Config) (*Identity, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	return DetectIdentity(config, cwd)
}

// MatchingFolder returns which folder pattern matched for the given directory
func MatchingFolder(identity Identity, dir string) string {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	absDir = filepath.Clean(absDir)

	for _, folder := range identity.Folders {
		expandedFolder := expandPath(folder)
		absFolder, err := filepath.Abs(expandedFolder)
		if err != nil {
			continue
		}
		absFolder = filepath.Clean(absFolder)

		if strings.HasPrefix(absDir, absFolder) {
			return folder
		}
	}

	return ""
}
