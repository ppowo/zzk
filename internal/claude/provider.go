package claude

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Provider represents a Claude API provider configuration
type Provider struct {
	BaseURL          string `json:"base_url"`
	APIToken         string `json:"api_token"`
	OpusModel        string `json:"opus_model,omitempty"`
	SonnetModel      string `json:"sonnet_model,omitempty"`
	HaikuModel       string `json:"haiku_model,omitempty"`
	SubagentModel    string `json:"subagent_model,omitempty"`
	DisableTelemetry bool   `json:"disable_telemetry,omitempty"`
}

var (
	reservedNames = map[string]bool{
		"anthropic": true,
		"official":  true,
		"reset":     true,
		"default":   true,
	}
	providerNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// ValidateProviderName validates a provider name
func ValidateProviderName(name string) error {
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	// Check length bounds first (most basic validation)
	if len(name) < 2 {
		return fmt.Errorf("provider name too short (min 2 characters)")
	}
	if len(name) > 64 {
		return fmt.Errorf("provider name too long (max 64 characters)")
	}

	// Only alphanumeric, hyphens, underscores
	if !providerNameRegex.MatchString(name) {
		return fmt.Errorf("provider name must contain only letters, numbers, hyphens, and underscores")
	}

	// Check it doesn't start or end with hyphen or underscore
	if name[0] == '-' || name[0] == '_' || name[len(name)-1] == '-' || name[len(name)-1] == '_' {
		return fmt.Errorf("provider name cannot start or end with hyphen or underscore")
	}

	// Check reserved names last (semantic validation)
	if reservedNames[name] {
		return fmt.Errorf("'%s' is a reserved name, please choose another", name)
	}

	return nil
}

// Validate validates a provider configuration
func (p *Provider) Validate() error {
	if p.BaseURL == "" {
		return fmt.Errorf("BASE_URL is required")
	}
	parsedURL, err := url.Parse(p.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid BASE_URL %q: %w", p.BaseURL, err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("BASE_URL must use http or https scheme, got %q", parsedURL.Scheme)
	}
	// Warn if using insecure HTTP
	if parsedURL.Scheme == "http" {
		return fmt.Errorf("BASE_URL uses insecure HTTP scheme - API tokens will be sent in plaintext!\nFor security, use HTTPS instead")
	}
	if parsedURL.Host == "" {
		return fmt.Errorf("BASE_URL must have a valid host")
	}
	if parsedURL.User != nil {
		return fmt.Errorf("BASE_URL must not contain credentials (user:pass@)")
	}
	// Check for trailing slash (will be stripped before saving)
	if strings.HasSuffix(p.BaseURL, "/") {
		return fmt.Errorf("BASE_URL must not have trailing slash")
	}
	if p.APIToken == ""  {
		return fmt.Errorf("API_TOKEN is required")
	}
	// Check for newlines (would break env file format)
	if strings.ContainsAny(p.APIToken, "\n\r") {
		return fmt.Errorf("API_TOKEN must not contain newlines")
	}
	// Check actual token length (before trimming)
	if len(p.APIToken) < 8 {
		return fmt.Errorf("API_TOKEN must be at least 8 characters")
	}
	// Ensure no leading/trailing whitespace
	if strings.TrimSpace(p.APIToken) != p.APIToken {
		return fmt.Errorf("API_TOKEN must not have leading or trailing whitespace")
	}

	// Validate model names if provided
	if err := validateModelName("OPUS_MODEL", p.OpusModel); err != nil {
		return err
	}
	if err := validateModelName("SONNET_MODEL", p.SonnetModel); err != nil {
		return err
	}
	if err := validateModelName("HAIKU_MODEL", p.HaikuModel); err != nil {
		return err
	}
	if err := validateModelName("SUBAGENT_MODEL", p.SubagentModel); err != nil {
		return err
	}

	return nil
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

// Template returns a template for creating a new provider
func Template() string {
	return `# Provider configuration for Claude Code

[required]
base_url = "https://api.example.com/anthropic"
api_token = "your-token-here"

[optional]
# Uncomment and set to override default models
# opus_model = ""
# sonnet_model = ""
# haiku_model = ""
# subagent_model = ""

# Set to true to disable telemetry
# disable_telemetry = false
`
}

// ToTemplate converts a provider to template format for editing
func (p *Provider) ToTemplate() string {
	var buf strings.Builder

	buf.WriteString("# Provider configuration for Claude Code\n\n")
	buf.WriteString("[required]\n")
	fmt.Fprintf(&buf, "base_url = %q\n", p.BaseURL)
	fmt.Fprintf(&buf, "api_token = %q\n\n", p.APIToken)
	buf.WriteString("[optional]\n")

	if p.OpusModel != "" {
		fmt.Fprintf(&buf, "opus_model = %q\n", p.OpusModel)
	} else {
		buf.WriteString("# opus_model = \"\"\n")
	}

	if p.SonnetModel != "" {
		fmt.Fprintf(&buf, "sonnet_model = %q\n", p.SonnetModel)
	} else {
		buf.WriteString("# sonnet_model = \"\"\n")
	}

	if p.HaikuModel != "" {
		fmt.Fprintf(&buf, "haiku_model = %q\n", p.HaikuModel)
	} else {
		buf.WriteString("# haiku_model = \"\"\n")
	}

	if p.SubagentModel != "" {
		fmt.Fprintf(&buf, "subagent_model = %q\n", p.SubagentModel)
	} else {
		buf.WriteString("# subagent_model = \"\"\n")
	}

	fmt.Fprintf(&buf, "disable_telemetry = %t\n", p.DisableTelemetry)

	return buf.String()
}

// ToShellExports returns shell export commands for this provider
func (p *Provider) ToShellExports() string {
	var buf strings.Builder

	fmt.Fprintf(&buf, "export ANTHROPIC_BASE_URL=%q\n", p.BaseURL)
	fmt.Fprintf(&buf, "export ANTHROPIC_AUTH_TOKEN=%q\n", p.APIToken)

	// Model variables: either set them or unset them to clear previous values
	if p.OpusModel != "" {
		fmt.Fprintf(&buf, "export ANTHROPIC_DEFAULT_OPUS_MODEL=%q\n", p.OpusModel)
	} else {
		buf.WriteString("unset ANTHROPIC_DEFAULT_OPUS_MODEL\n")
	}
	if p.SonnetModel != "" {
		fmt.Fprintf(&buf, "export ANTHROPIC_DEFAULT_SONNET_MODEL=%q\n", p.SonnetModel)
	} else {
		buf.WriteString("unset ANTHROPIC_DEFAULT_SONNET_MODEL\n")
	}
	if p.HaikuModel != "" {
		fmt.Fprintf(&buf, "export ANTHROPIC_DEFAULT_HAIKU_MODEL=%q\n", p.HaikuModel)
	} else {
		buf.WriteString("unset ANTHROPIC_DEFAULT_HAIKU_MODEL\n")
	}
	if p.SubagentModel != "" {
		fmt.Fprintf(&buf, "export CLAUDE_CODE_SUBAGENT_MODEL=%q\n", p.SubagentModel)
	} else {
		buf.WriteString("unset CLAUDE_CODE_SUBAGENT_MODEL\n")
	}
	if p.DisableTelemetry {
		buf.WriteString("export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1\n")
	} else {
		buf.WriteString("unset CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC\n")
	}

	return buf.String()
}
