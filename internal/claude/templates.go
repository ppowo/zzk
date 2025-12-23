package claude

import (
	"fmt"
	"strings"
)

// ProviderTemplate represents a hardcoded Claude API provider.
type ProviderTemplate struct {
	ID           string // Unique identifier (e.g., "synthetic", "openrouter")
	Name         string // Display name
	BaseURL      string // Fixed API base URL
	AllowModels  bool   // Whether model overrides are allowed
	DefaultModel string // Default model for all model types (used when user doesn't specify)
}

// Templates is the registry of all known Claude API providers.
var Templates = []ProviderTemplate{
	{
		ID:           "synthetic",
		Name:         "Synthetic",
		BaseURL:      "https://api.synthetic.new/anthropic",
		AllowModels:  true,
		DefaultModel: "hf:zai-org/GLM-4.7",
	},
	{
		ID:           "openrouter",
		Name:         "OpenRouter",
		BaseURL:      "https://openrouter.ai/api",
		AllowModels:  true,
		DefaultModel: "openai/gpt-oss-120b:free",
	},
	{
		ID:           "zai",
		Name:         "Z.AI",
		BaseURL:      "https://api.z.ai/api/anthropic",
		AllowModels:  false,
		DefaultModel: "",
	},
}

// GetTemplate returns a provider template by ID.
// Returns nil and false if the template doesn't exist.
func GetTemplate(id string) (*ProviderTemplate, bool) {
	for i := range Templates {
		if Templates[i].ID == id {
			return &Templates[i], true
		}
	}
	return nil, false
}

// ListTemplates returns all available provider templates.
func ListTemplates() []ProviderTemplate {
	return Templates
}

// IsValidTemplate checks if a template ID exists.
func IsValidTemplate(id string) bool {
	_, ok := GetTemplate(id)
	return ok
}

// TemplateIDs returns a list of all valid template IDs.
func TemplateIDs() []string {
	ids := make([]string, len(Templates))
	for i, t := range Templates {
		ids[i] = t.ID
	}
	return ids
}

// ResolveTemplateID resolves a prefix to a full template ID.
// Returns the full ID if found, or an error if no match or ambiguous.
func ResolveTemplateID(prefix string) (string, error) {
	// Try exact match first
	if IsValidTemplate(prefix) {
		return prefix, nil
	}

	// Try prefix matching
	var matches []string
	for _, t := range Templates {
		if len(prefix) <= len(t.ID) && t.ID[:len(prefix)] == prefix {
			matches = append(matches, t.ID)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("unknown provider '%s'\n\nAvailable providers: %s",
			prefix, strings.Join(TemplateIDs(), ", "))
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("ambiguous provider '%s' matches: %s",
			prefix, strings.Join(matches, ", "))
	}
}
