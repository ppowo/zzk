package cmd

import (
	"fmt"

	"github.com/ppowo/zzk/internal/claude"
	"github.com/spf13/cobra"
)

var (
	forceRemove bool
)

var claudeRmCmd = &cobra.Command{
	Use:   "rm <provider-name>",
	Short: "Remove a Claude API provider",
	Long: `Remove a Claude API provider from your configuration.

If the provider is currently active, it will be automatically reset
to the official Anthropic API.

By default, you will be prompted to confirm deletion. Use -f to skip confirmation.

Examples:
  zzk claude rm synthetic         # Remove provider with confirmation
  zzk claude rm synthetic -f      # Remove without confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		providerName := args[0]

		// Load config
		config, err := claude.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Check if provider exists
		if !config.HasProvider(providerName) {
			return fmt.Errorf("provider '%s' not found", providerName)
		}

		// Confirm deletion unless forced
		if !forceRemove {
			confirmed, err := claude.PromptYesNo(fmt.Sprintf("Remove provider '%s'?", providerName), false)
			if err != nil {
				return fmt.Errorf("%w. Use -f to force", err)
			}
			if !confirmed {
				fmt.Println("Cancelled")
				return nil
			}
		}

		wasActive := config.Active == providerName

		// Remove provider
		if err := config.RemoveProvider(providerName); err != nil {
			return fmt.Errorf("failed to remove provider: %w", err)
		}

		// Save config
		if err := claude.SaveConfig(config); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("✓ Provider '%s' removed\n", providerName)

		if wasActive {
			if err := claude.ResetToOfficialAPI(); err != nil {
				return fmt.Errorf("failed to reset to official API: %w", err)
			}
		}

		return nil
	},
}

func init() {
	claudeRmCmd.Flags().BoolVarP(&forceRemove, "force", "f", false, "Skip confirmation prompt")
	claudeCmd.AddCommand(claudeRmCmd)
}
