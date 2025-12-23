package claude

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// PromptForProvider prompts the user for provider configuration.
// If existingProvider is not nil, it pre-fills with existing values.
func PromptForProvider(templateID string, existingProvider *Provider) (*Provider, error) {
	tmpl, ok := GetTemplate(templateID)
	if !ok {
		return nil, fmt.Errorf("unknown provider template: %s", templateID)
	}

	reader := bufio.NewReader(os.Stdin)

	// Prompt for API key
	apiKey, err := promptForAPIKey(reader, existingProvider)
	if err != nil {
		return nil, err
	}

	provider := &Provider{
		APIKey: apiKey,
	}

	// Prompt for model overrides if the template allows it
	if tmpl.AllowModels {
		models, err := promptForModels(reader, tmpl, existingProvider)
		if err != nil {
			return nil, err
		}
		provider.OpusModel = models.OpusModel
		provider.SonnetModel = models.SonnetModel
		provider.HaikuModel = models.HaikuModel
		provider.SubagentModel = models.SubagentModel
	}

	// Validate the provider
	if err := provider.Validate(templateID); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return provider, nil
}

// promptForAPIKey prompts for and reads the API key
func promptForAPIKey(reader *bufio.Reader, existing *Provider) (string, error) {
	var defaultVal string
	if existing != nil && existing.APIKey != "" {
		// Show masked version of existing key
		maskedKey := maskAPIKey(existing.APIKey)
		fmt.Printf("API key [current: %s]: ", maskedKey)
		defaultVal = existing.APIKey
	} else {
		fmt.Print("API key: ")
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	line = strings.TrimSpace(line)
	if line == "" {
		if defaultVal != "" {
			return defaultVal, nil
		}
		return "", fmt.Errorf("API key is required")
	}

	return line, nil
}

// maskAPIKey returns a masked version of an API key for display
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "********"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

// modelConfig holds the model override values
type modelConfig struct {
	OpusModel     string
	SonnetModel   string
	HaikuModel    string
	SubagentModel string
}

// promptForModels prompts for model overrides
func promptForModels(reader *bufio.Reader, tmpl *ProviderTemplate, existing *Provider) (modelConfig, error) {
	var models modelConfig
	var err error

	// Returns (value, isCurrent) - isCurrent true if from existing config, false if from template
	getDefaultWithSource := func(modelType string) (string, bool) {
		if existing != nil {
			if val := getExistingModel(existing, modelType); val != "" {
				return val, true // current setting
			}
		}
		return tmpl.DefaultModel, false // template default
	}

	fmt.Println("\nModel overrides (leave empty to keep shown value):")

	opusVal, opusIsCurrent := getDefaultWithSource("opus")
	models.OpusModel, err = promptForModelWithSource(reader, "Opus model", opusVal, opusIsCurrent)
	if err != nil {
		return models, err
	}

	sonnetVal, sonnetIsCurrent := getDefaultWithSource("sonnet")
	models.SonnetModel, err = promptForModelWithSource(reader, "Sonnet model", sonnetVal, sonnetIsCurrent)
	if err != nil {
		return models, err
	}

	haikuVal, haikuIsCurrent := getDefaultWithSource("haiku")
	models.HaikuModel, err = promptForModelWithSource(reader, "Haiku model", haikuVal, haikuIsCurrent)
	if err != nil {
		return models, err
	}

	subagentVal, subagentIsCurrent := getDefaultWithSource("subagent")
	models.SubagentModel, err = promptForModelWithSource(reader, "Subagent model", subagentVal, subagentIsCurrent)
	if err != nil {
		return models, err
	}

	return models, nil
}

// getExistingModel returns the existing model value or empty string
func getExistingModel(existing *Provider, modelType string) string {
	if existing == nil {
		return ""
	}
	switch modelType {
	case "opus":
		return existing.OpusModel
	case "sonnet":
		return existing.SonnetModel
	case "haiku":
		return existing.HaikuModel
	case "subagent":
		return existing.SubagentModel
	}
	return ""
}

// promptForModelWithSource prompts for a single model override, showing source label
func promptForModelWithSource(reader *bufio.Reader, label string, defaultVal string, isCurrent bool) (string, error) {
	if defaultVal != "" {
		sourceLabel := "default"
		if isCurrent {
			sourceLabel = "current"
		}
		fmt.Printf("  %s [%s: %s] ('default' to reset): ", label, sourceLabel, defaultVal)
	} else {
		fmt.Printf("  %s: ", label)
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return defaultVal, nil
	}
	if line == "default" {
		return "", nil // Reset to template default
	}

	return line, nil
}
