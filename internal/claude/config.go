package claude

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ppowo/zzk/internal/fileutil"
)

// Config represents the ~/.claude-providers.json configuration file
type Config struct {
	Providers map[string]Provider `json:"providers"`
	Active    string              `json:"active,omitempty"`
}

// ConfigPath returns the path to the config file
func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home unavailable
		cwd, _ := os.Getwd()
		if cwd != "" {
			return filepath.Join(cwd, ".claude-providers.json")
		}
		return ".claude-providers.json"
	}
	return filepath.Join(home, ".claude-providers.json")
}

// EnsureConfigDir ensures the config directory exists
func EnsureConfigDir() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	dir := filepath.Join(home, ".config", "zzk")
	return os.MkdirAll(dir, 0755)
}

// LoadConfig loads the configuration from ~/.claude-providers.json
func LoadConfig() (*Config, error) {
	path := ConfigPath()

	// If file doesn't exist, return empty config
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return &Config{
			Providers: make(map[string]Provider),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat config: %w", err)
	}

	// Limit config file size to 10MB to prevent DoS
	const maxConfigSize = 10 * 1024 * 1024
	if stat.Size() > maxConfigSize {
		return nil, fmt.Errorf("config file too large (max 10MB)")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("invalid JSON in config file: %w", err)
	}

	// Initialize map if nil
	if config.Providers == nil {
		config.Providers = make(map[string]Provider)
	}

	// Check for old config format (has base_url or api_token fields)
	if err := config.detectOldFormat(data); err != nil {
		return nil, err
	}

	// Validate all provider keys are valid template IDs
	for name := range config.Providers {
		if !IsValidTemplate(name) {
			return nil, fmt.Errorf("unknown provider '%s' in config - valid providers: %v", name, TemplateIDs())
		}
	}

	// Auto-fix broken active reference
	if config.Active != "" {
		if _, exists := config.Providers[config.Active]; !exists {
			fmt.Fprintf(os.Stderr, "Warning: active provider '%s' not found, clearing\n", config.Active)
			config.Active = ""
		}
	}

	return &config, nil
}

// detectOldFormat checks if the config uses the old format (with base_url field)
func (c *Config) detectOldFormat(data []byte) error {
	// Parse as raw JSON to check for old fields
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil // Let the main parser handle JSON errors
	}

	providersRaw, ok := raw["providers"]
	if !ok {
		return nil // No providers, nothing to check
	}

	var providers map[string]json.RawMessage
	if err := json.Unmarshal(providersRaw, &providers); err != nil {
		return nil
	}

	for name, providerRaw := range providers {
		var fields map[string]interface{}
		if err := json.Unmarshal(providerRaw, &fields); err != nil {
			continue
		}

		// Check for old format fields
		if _, hasBaseURL := fields["base_url"]; hasBaseURL {
			return fmt.Errorf(`config file uses old format with 'base_url' field

The provider configuration format has changed. Provider URLs are now hardcoded.

To migrate, delete %s and reconfigure your providers:
  rm %s
  zzk claude add synthetic    # for Synthetic
  zzk claude add openrouter   # for OpenRouter
  zzk claude add zai          # for Z.AI

Your old provider '%s' had a custom URL which is no longer supported.`, ConfigPath(), ConfigPath(), name)
		}

		if _, hasAPIToken := fields["api_token"]; hasAPIToken {
			return fmt.Errorf(`config file uses old format with 'api_token' field

The provider configuration format has changed. The field is now 'api_key'.

To migrate, delete %s and reconfigure your providers:
  rm %s
  zzk claude add synthetic    # for Synthetic
  zzk claude add openrouter   # for OpenRouter
  zzk claude add zai          # for Z.AI`, ConfigPath(), ConfigPath())
		}
	}

	return nil
}

// SaveConfig saves the configuration to ~/.claude-providers.json
func SaveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	path := ConfigPath()

	// Create backup if original exists
	if _, err := os.Stat(path); err == nil {
		backup := path + ".backup"
		if err := fileutil.CopyFile(path, backup); err != nil {
			// Non-fatal: warn but continue
			fmt.Fprintf(os.Stderr, "Warning: failed to create backup: %v\n", err)
		}
	}

	return fileutil.AtomicWrite(path, data, 0600)
}

// HasProvider checks if a provider exists in the config
func (c *Config) HasProvider(templateID string) bool {
	_, ok := c.Providers[templateID]
	return ok
}

// GetProvider returns a provider by template ID
func (c *Config) GetProvider(templateID string) (Provider, bool) {
	provider, ok := c.Providers[templateID]
	return provider, ok
}

// AddProvider adds or updates a provider in the config.
// The templateID must be a valid template from the registry.
func (c *Config) AddProvider(templateID string, provider Provider) error {
	if !IsValidTemplate(templateID) {
		return fmt.Errorf("unknown provider template: %s (valid: %v)", templateID, TemplateIDs())
	}
	if err := provider.Validate(templateID); err != nil {
		return fmt.Errorf("invalid provider: %w", err)
	}
	c.Providers[templateID] = provider
	return nil
}

// RemoveProvider removes a provider from the config
func (c *Config) RemoveProvider(templateID string) error {
	if !c.HasProvider(templateID) {
		return fmt.Errorf("provider '%s' not configured", templateID)
	}
	delete(c.Providers, templateID)

	// Clear active if this was the active provider
	if c.Active == templateID {
		c.Active = ""
	}

	return nil
}

// SetActive sets the active provider
func (c *Config) SetActive(templateID string) error {
	if !c.HasProvider(templateID) {
		return fmt.Errorf("provider '%s' not configured", templateID)
	}
	c.Active = templateID
	return nil
}

// ClearActive clears the active provider
func (c *Config) ClearActive() {
	c.Active = ""
}
