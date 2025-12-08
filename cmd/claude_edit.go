package cmd

import (
	"fmt"
	"os"

	"github.com/ppowo/zzk/internal/claude"
	"github.com/spf13/cobra"
)

var claudeEditCmd = &cobra.Command{
	Use:   "edit <provider-name>",
	Short: "Edit an existing Claude API provider",
	Long: `Edit an existing Claude API provider configuration.

Opens your editor with the current configuration so you can modify it.

Examples:
  zzk claude edit synthetic       # Edit provider named "synthetic"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		providerName := args[0]

		// Load config
		config, err := claude.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Check if provider exists
		existingProvider, exists := config.GetProvider(providerName)
		if !exists {
			return fmt.Errorf("provider '%s' not found. Use 'zzk claude add %s' to create it", providerName, providerName)
		}

		fmt.Printf("Opening editor to modify provider '%s'...\n", providerName)

		// Open editor with existing provider, retry on validation failure
		provider, err := claude.EditProviderWithRetry(&existingProvider)
		if err != nil {
			// Don't show error details if user simply didn't make changes
			if err.Error() == "no changes made" {
				fmt.Println("No changes made.")
				return nil
			}
			return fmt.Errorf("failed to edit provider: %w", err)
		}

		// Update provider in config
		if err := config.AddProvider(providerName, *provider); err != nil {
			return fmt.Errorf("failed to update provider: %w", err)
		}

		// Save config
		if err := claude.SaveConfig(config); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("\n✓ Provider '%s' updated successfully!\n", providerName)

		// If this was the active provider, update env file and check shell sync
		// Also reload if this provider matches what's in the shell (even if active field is missing)
		shouldReload := config.Active == providerName
		if !shouldReload {
			// Check if shell environment is using this provider
			envBaseURL := os.Getenv("ANTHROPIC_BASE_URL")
			if envBaseURL == existingProvider.BaseURL {
				shouldReload = true
			}
		}

		if shouldReload {
			if err := claude.ReloadClaudeEnvironment(providerName, *provider); err != nil {
				return fmt.Errorf("failed to reload Claude environment: %w", err)
			}
		}

		return nil
	},
}

func init() {
	claudeCmd.AddCommand(claudeEditCmd)
}
