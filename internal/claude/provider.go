package claude

import (
	"fmt"
	"strings"
)

// Provider represents a user's configuration for a Claude API provider.
// The provider template (ID, base URL) is determined by the template registry.
type Provider struct {
	APIKey        string `json:"api_key"`
	OpusModel     string `json:"opus_model,omitempty"`
	SonnetModel   string `json:"sonnet_model,omitempty"`
	HaikuModel    string `json:"haiku_model,omitempty"`
	SubagentModel string `json:"subagent_model,omitempty"`
}

// Validate validates a provider configuration.
// If templateID is provided, it also validates model overrides against template rules.
func (p *Provider) Validate(templateID string) error {
	if p.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	// Check for newlines (would break env file format)
	if strings.ContainsAny(p.APIKey, "\n\r") {
		return fmt.Errorf("API key must not contain newlines")
	}
	// Check actual key length
	if len(p.APIKey) < 8 {
		return fmt.Errorf("API key must be at least 8 characters")
	}
	// Ensure no leading/trailing whitespace
	if strings.TrimSpace(p.APIKey) != p.APIKey {
		return fmt.Errorf("API key must not have leading or trailing whitespace")
	}

	// Check if template allows model overrides
	if templateID != "" {
		tmpl, ok := GetTemplate(templateID)
		if !ok {
			return fmt.Errorf("unknown provider template: %s", templateID)
		}
		if !tmpl.AllowModels && p.HasModelOverrides() {
			return fmt.Errorf("provider %q does not support model overrides", tmpl.Name)
		}
	}

	// Validate model names if provided
	if err := validateModelName("opus_model", p.OpusModel); err != nil {
		return err
	}
	if err := validateModelName("sonnet_model", p.SonnetModel); err != nil {
		return err
	}
	if err := validateModelName("haiku_model", p.HaikuModel); err != nil {
		return err
	}
	if err := validateModelName("subagent_model", p.SubagentModel); err != nil {
		return err
	}

	return nil
}

// HasModelOverrides returns true if any model override is set.
func (p *Provider) HasModelOverrides() bool {
	return p.OpusModel != "" || p.SonnetModel != "" || p.HaikuModel != "" || p.SubagentModel != ""
}

// validateModelName validates a model name doesn't contain problematic characters
func validateModelName(fieldName, modelName string) error {
	if modelName == "" {
		return nil // Optional field
	}
	// Check for newlines, carriage returns, or null bytes
	if strings.ContainsAny(modelName, "\n\r\x00") {
		return fmt.Errorf("%s contains invalid characters (newlines or null bytes)", fieldName)
	}
	// Check for spaces (could cause quoting issues)
	if strings.Contains(modelName, " ") {
		return fmt.Errorf("%s must not contain spaces", fieldName)
	}
	// Check for shell metacharacters that could cause issues
	if strings.ContainsAny(modelName, "$`\\;|&<>(){}[]") {
		return fmt.Errorf("%s contains shell metacharacters", fieldName)
	}
	// Reasonable length limit
	if len(modelName) > 256 {
		return fmt.Errorf("%s too long (max 256 characters)", fieldName)
	}
	return nil
}

// ToShellExports returns shell export commands for this provider.
// The templateID is required to look up the base URL from the template registry.
func (p *Provider) ToShellExports(templateID string) (string, error) {
	tmpl, ok := GetTemplate(templateID)
	if !ok {
		return "", fmt.Errorf("unknown provider template: %s", templateID)
	}

	var buf strings.Builder

	fmt.Fprintf(&buf, "export ANTHROPIC_BASE_URL=%q\n", tmpl.BaseURL)
	fmt.Fprintf(&buf, "export ANTHROPIC_AUTH_TOKEN=%q\n", p.APIKey)

	// Helper to get model value: use provider value if set, else template default
	getModel := func(providerModel string) string {
		if providerModel != "" {
			return providerModel
		}
		return tmpl.DefaultModel
	}

	// Model variables: export if we have a value (from provider or template default), else unset
	if model := getModel(p.OpusModel); model != "" {
		fmt.Fprintf(&buf, "export ANTHROPIC_DEFAULT_OPUS_MODEL=%q\n", model)
	} else {
		buf.WriteString("unset ANTHROPIC_DEFAULT_OPUS_MODEL\n")
	}
	if model := getModel(p.SonnetModel); model != "" {
		fmt.Fprintf(&buf, "export ANTHROPIC_DEFAULT_SONNET_MODEL=%q\n", model)
	} else {
		buf.WriteString("unset ANTHROPIC_DEFAULT_SONNET_MODEL\n")
	}
	if model := getModel(p.HaikuModel); model != "" {
		fmt.Fprintf(&buf, "export ANTHROPIC_DEFAULT_HAIKU_MODEL=%q\n", model)
	} else {
		buf.WriteString("unset ANTHROPIC_DEFAULT_HAIKU_MODEL\n")
	}
	if model := getModel(p.SubagentModel); model != "" {
		fmt.Fprintf(&buf, "export CLAUDE_CODE_SUBAGENT_MODEL=%q\n", model)
	} else {
		buf.WriteString("unset CLAUDE_CODE_SUBAGENT_MODEL\n")
	}

	// Always export hardcoded values for timeout and telemetry
	fmt.Fprintf(&buf, "export API_TIMEOUT_MS=%d\n", 6000000)
	buf.WriteString("export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1\n")

	return buf.String(), nil
}
