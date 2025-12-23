package cmd

import (
	"fmt"
	"os"

	"github.com/ppowo/zzk/internal/claude"
	"github.com/spf13/cobra"
)

var claudeSetCmd = &cobra.Command{
	Use:   "set <provider>",
	Short: "Configure a Claude API provider",
	Long: `Configure a Claude API provider (add new or update existing).

If the provider is already configured, existing values are shown as defaults.
If the provider is currently active, the environment is automatically reloaded.

Provider IDs support prefix matching (e.g., 'syn' matches 'synthetic').

Examples:
  zzk claude set synthetic    # Configure Synthetic provider
  zzk claude set syn          # Same (prefix matching)
  zzk claude set openrouter   # Configure OpenRouter provider`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Resolve prefix to full template ID
		templateID, err := claude.ResolveTemplateID(args[0])
		if err != nil {
			return err
		}

		tmpl, _ := claude.GetTemplate(templateID)

		// Load config
		config, err := claude.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Check if provider exists (determines prompt behavior)
		existingProvider, exists := config.GetProvider(templateID)
		var existing *claude.Provider
		if exists {
			existing = &existingProvider
			fmt.Printf("Updating %s (%s)\n\n", tmpl.Name, tmpl.BaseURL)
		} else {
			fmt.Printf("Configuring %s (%s)\n\n", tmpl.Name, tmpl.BaseURL)
		}

		// Prompt for provider configuration
		provider, err := claude.PromptForProvider(templateID, existing)
		if err != nil {
			return fmt.Errorf("failed to configure provider: %w", err)
		}

		// Save provider to config
		if err := config.AddProvider(templateID, *provider); err != nil {
			return fmt.Errorf("failed to save provider: %w", err)
		}
		if err := claude.SaveConfig(config); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		if exists {
			fmt.Printf("\nProvider '%s' updated successfully!\n", tmpl.Name)
		} else {
			fmt.Printf("\nProvider '%s' configured successfully!\n", tmpl.Name)
		}

		// Reload if this is the active provider
		shouldReload := config.Active == templateID
		if !shouldReload {
			// Also check if shell environment is using this provider's base URL
			envBaseURL := os.Getenv("ANTHROPIC_BASE_URL")
			if envBaseURL == tmpl.BaseURL {
				shouldReload = true
			}
		}

		if shouldReload {
			if err := claude.ReloadClaudeEnvironment(templateID, *provider); err != nil {
				return fmt.Errorf("failed to reload Claude environment: %w", err)
			}
		} else if !exists {
			fmt.Printf("\nTo activate this provider, run:\n")
			fmt.Printf("  zzk claude use %s\n", templateID)
		}

		return nil
	},
}

func init() {
	claudeCmd.AddCommand(claudeSetCmd)
}
