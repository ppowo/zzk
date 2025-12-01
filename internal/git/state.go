package git

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// State represents the persistent state of git identity management
type State struct {
	Version    string                     `json:"version"`
	LastSync   time.Time                  `json:"lastSync"`
	Identities map[string]*IdentityState `json:"identities"`
}

// IdentityState tracks the state of a single identity
type IdentityState struct {
	LastSync          time.Time `json:"lastSync"`
	SSHKeyFingerprint string    `json:"sshKeyFingerprint,omitempty"`
}

// LoadState loads the state file or creates a new one if it doesn't exist
func LoadState() (*State, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".config", "zzk")
	statePath := filepath.Join(configDir, "git-state.json")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	// Try to read existing state
	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, create new state
			return &State{
				Version:    "1.0",
				Identities: make(map[string]*IdentityState),
			}, nil
		}
		return nil, err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	// Initialize identities map if nil
	if state.Identities == nil {
		state.Identities = make(map[string]*IdentityState)
	}

	return &state, nil
}

// Save saves the state to disk
func (s *State) Save() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".config", "zzk")
	statePath := filepath.Join(configDir, "git-state.json")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	// Write to file atomically
	return os.WriteFile(statePath, data, 0644)
}
