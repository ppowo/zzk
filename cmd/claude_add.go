package cmd

import (
	"fmt"

	"github.com/ppowo/zzk/internal/claude"
	"github.com/spf13/cobra"
)

var claudeAddCmd = &cobra.Command{
	Use:   "add <provider-name>",
	Short: "Add a new Claude API provider",
	Long: `Add a new Claude API provider by opening your editor with a configuration template.

The provider name must be alphanumeric with optional hyphens or underscores.
Reserved names (anthropic, official, reset, default) cannot be used.

After running this command, your editor will open with a template.
Fill in the BASE_URL and API_TOKEN (required), and optionally configure
model overrides and telemetry settings.

Examples:
  zzk claude add synthetic        # Add provider named "synthetic"
  zzk claude add my-provider      # Add provider named "my-provider"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		providerName := args[0]

		// Validate provider name
		if err := claude.ValidateProviderName(providerName); err != nil {
			return err
		}

		// Load config
		config, err := claude.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Check if provider already exists
		if config.HasProvider(providerName) {
			return fmt.Errorf("provider '%s' already exists. Use 'zzk claude edit %s' to modify it", providerName, providerName)
		}

		fmt.Printf("Opening editor to configure provider '%s'...\n", providerName)

		// Open editor for new provider, retry on validation failure
		provider, err := claude.EditProviderWithRetry(nil)
		if err != nil {
			return fmt.Errorf("failed to create provider: %w", err)
		}

		// Add provider to config
		if err := config.AddProvider(providerName, *provider); err != nil {
			return fmt.Errorf("failed to add provider: %w", err)
		}

		// Save config
		if err := claude.SaveConfig(config); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("\n✓ Provider '%s' added successfully!\n", providerName)
		fmt.Printf("\nConfiguration saved to: %s\n", claude.ConfigPath())
		fmt.Printf("\nTo use this provider, run:\n")
		fmt.Printf("  zzk claude use %s\n", providerName)

		return nil
	},
}

func init() {
	claudeCmd.AddCommand(claudeAddCmd)
}
