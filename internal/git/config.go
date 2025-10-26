package git

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the ~/.git-identities.json configuration file
type Config struct {
	Identities map[string]Identity `json:"identities"`
}

// ConfigPath returns the path to the config file
func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "~/.git-identities.json"
	}
	return filepath.Join(home, ".git-identities.json")
}

// LoadConfig loads the configuration from ~/.git-identities.json
func LoadConfig() (*Config, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if config.Identities == nil {
		return nil, fmt.Errorf("no identities defined in config")
	}

	for name, identity := range config.Identities {
		identity.Name = name
		if err := identity.Validate(); err != nil {
			return nil, fmt.Errorf("invalid identity %s: %w", name, err)
		}
		config.Identities[name] = identity
	}

	return &config, nil
}

// SaveConfig saves the configuration to ~/.git-identities.json
func SaveConfig(config *Config) error {
	path := ConfigPath()
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// CreateExampleConfig creates an example configuration file
func CreateExampleConfig() error {
	path := ConfigPath()

	exampleConfig := `{
  "identities": {
    "github-work": {
      "user": "Your GitHub Work Username",
      "email": "work@company.com",
      "domain": "github.com",
      "folders": [
        "~/Work/Github"
      ]
    },
    "github-personal": {
      "user": "Your GitHub Personal Username",
      "email": "personal@example.com",
      "domain": "github.com",
      "folders": [
        "~/Personal/Github"
      ]
    },
    "gitlab": {
      "user": "Your GitLab Username",
      "email": "user@gitlab.com",
      "domain": "gitlab.com",
      "folders": [
        "~/Gitlab"
      ]
    },
    "codeberg": {
      "user": "Your Codeberg Username",
      "email": "username@noreply.codeberg.org",
      "domain": "codeberg.org",
      "folders": [
        "~/Codeberg"
      ]
    }
  }
}
`

	if err := os.WriteFile(path, []byte(exampleConfig), 0644); err != nil {
		return fmt.Errorf("failed to create example config: %w", err)
	}

	return nil
}

// HasIdentity checks if an identity exists in the config
func (c *Config) HasIdentity(name string) bool {
	_, ok := c.Identities[name]
	return ok
}

// GetIdentity returns an identity by name
func (c *Config) GetIdentity(name string) (Identity, bool) {
	identity, ok := c.Identities[name]
	return identity, ok
}
