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

	// Auto-fix broken active reference
	if config.Active != "" {
		if _, exists := config.Providers[config.Active]; !exists {
			fmt.Fprintf(os.Stderr, "Warning: active provider '%s' not found, clearing\n", config.Active)
			config.Active = ""
		}
	}

	return &config, nil
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
func (c *Config) HasProvider(name string) bool {
	_, ok := c.Providers[name]
	return ok
}

// GetProvider returns a provider by name
func (c *Config) GetProvider(name string) (Provider, bool) {
	provider, ok := c.Providers[name]
	return provider, ok
}

// AddProvider adds or updates a provider in the config
func (c *Config) AddProvider(name string, provider Provider) error {
	if err := ValidateProviderName(name); err != nil {
		return err
	}
	if err := provider.Validate(); err != nil {
		return fmt.Errorf("invalid provider: %w", err)
	}
	c.Providers[name] = provider
	return nil
}

// RemoveProvider removes a provider from the config
func (c *Config) RemoveProvider(name string) error {
	if !c.HasProvider(name) {
		return fmt.Errorf("provider '%s' not found", name)
	}
	delete(c.Providers, name)

	// Clear active if this was the active provider
	if c.Active == name {
		c.Active = ""
	}

	return nil
}

// SetActive sets the active provider
func (c *Config) SetActive(name string) error {
	if !c.HasProvider(name) {
		return fmt.Errorf("provider '%s' not found", name)
	}
	c.Active = name
	return nil
}

// ClearActive clears the active provider
func (c *Config) ClearActive() {
	c.Active = ""
}
