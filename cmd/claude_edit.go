package cmd

import (
	"fmt"

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

		// If this was the active provider, update env file
		if config.Active == providerName {
			if err := claude.WriteEnvFile(*provider); err != nil {
				return fmt.Errorf("failed to update env file: %w", err)
			}
			fmt.Println("✓ Active provider updated")

			// Check if shell is in sync
			if needsReload, warning := claude.CheckShellSync(provider.BaseURL); needsReload {
				fmt.Println(warning)
			}
		}

		return nil
	},
}

func init() {
	claudeCmd.AddCommand(claudeEditCmd)
}
