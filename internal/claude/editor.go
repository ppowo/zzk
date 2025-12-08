package claude

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/google/shlex"
)

// containsShellMetacharacters checks if a string contains dangerous shell metacharacters
func containsShellMetacharacters(cmd string) bool {
	// Check for shell metacharacters that could enable command injection
	dangerous := ";|&`$()<>\n\r"
	return strings.ContainsAny(cmd, dangerous)
}

// GetEditor returns the editor to use, checking $VISUAL, $EDITOR, then fallbacks
func GetEditor() (string, error) {
	// Priority: $VISUAL > $EDITOR > fallbacks
	if editor := os.Getenv("VISUAL"); editor != "" {
		// Validate that it doesn't contain obvious command injection attempts
		if containsShellMetacharacters(editor) {
			return "", fmt.Errorf("$VISUAL contains shell metacharacters (;|&`$()<>), refusing to execute for security")
		}
		return editor, nil
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		// Validate that it doesn't contain obvious command injection attempts
		if containsShellMetacharacters(editor) {
			return "", fmt.Errorf("$EDITOR contains shell metacharacters (;|&`$()<>), refusing to execute for security")
		}
		return editor, nil
	}

	// Try common editors in order
	for _, editor := range []string{"vim", "vi", "nano", "emacs"} {
		if path, err := exec.LookPath(editor); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no editor found. Please set $EDITOR or $VISUAL")
}

// EditFile opens a file in the user's editor and returns true if modified
func EditFile(path string) (bool, error) {
	editor, err := GetEditor()
	if err != nil {
		return false, err
	}

	// Get file hash before editing
	beforeHash, err := fileHash(path)
	if err != nil {
		return false, fmt.Errorf("failed to hash file before editing: %w", err)
	}

	// Parse editor command safely using shlex (POSIX shell quoting)
	parts, err := shlex.Split(editor)
	if err != nil {
		return false, fmt.Errorf("invalid editor command: %w", err)
	}
	if len(parts) == 0 {
		return false, fmt.Errorf("empty editor command")
	}

	// Build command with editor args + file path
	args := append(parts[1:], path)
	cmd := exec.Command(parts[0], args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("editor failed: %w", err)
	}

	// Check if file was actually modified by comparing hashes
	afterHash, err := fileHash(path)
	if err != nil {
		return false, fmt.Errorf("failed to hash file after editing: %w", err)
	}

	return beforeHash != afterHash, nil
}

// fileHash computes SHA256 hash of a file
func fileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// ProviderConfig represents the TOML structure for provider configuration
type ProviderConfig struct {
	Required struct {
		BaseURL  string `toml:"base_url"`
		APIToken string `toml:"api_token"`
	} `toml:"required"`
	Optional struct {
		OpusModel        string `toml:"opus_model"`
		SonnetModel      string `toml:"sonnet_model"`
		HaikuModel       string `toml:"haiku_model"`
		SubagentModel    string `toml:"subagent_model"`
		DisableTelemetry bool   `toml:"disable_telemetry"`
	} `toml:"optional"`
}

// ParseProviderFile parses a provider configuration file in TOML format
func ParseProviderFile(path string) (*Provider, error) {
	// Limit file size to 1MB to prevent DoS
	const maxFileSize = 1024 * 1024

	stat, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	if stat.Size() > maxFileSize {
		return nil, fmt.Errorf("config file too large (max 1MB)")
	}

	var cfg ProviderConfig
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	return &Provider{
		BaseURL:          cfg.Required.BaseURL,
		APIToken:         cfg.Required.APIToken,
		OpusModel:        cfg.Optional.OpusModel,
		SonnetModel:      cfg.Optional.SonnetModel,
		HaikuModel:       cfg.Optional.HaikuModel,
		SubagentModel:    cfg.Optional.SubagentModel,
		DisableTelemetry: cfg.Optional.DisableTelemetry,
	}, nil
}

// EditProvider opens an editor for creating or editing a provider
func EditProvider(existingProvider *Provider) (*Provider, error) {
	// Create temporary file in a secure location
	tmpFile, err := os.CreateTemp("", "zzk-claude-*.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Set restrictive permissions immediately before writing sensitive data
	if err := os.Chmod(tmpPath, 0600); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return nil, fmt.Errorf("failed to set secure permissions: %w", err)
	}

	// Ensure cleanup even on panic or error
	defer func() {
		os.Remove(tmpPath)
	}()

	// Write template or existing config to temp file
	var content string
	if existingProvider != nil {
		content = existingProvider.ToTemplate()
	} else {
		content = Template()
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// Open editor
	modified, err := EditFile(tmpPath)
	if err != nil {
		return nil, err
	}

	if !modified {
		return nil, nil // Return nil to indicate user cancelled
	}

	// Parse the edited file
	provider, err := ParseProviderFile(tmpPath)
	if err != nil {
		return nil, err
	}

	// Validate
	if err := provider.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return provider, nil
}
