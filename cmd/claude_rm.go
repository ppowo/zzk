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
	Use:   "rm <provider>",
	Short: "Remove a Claude API provider configuration",
	Long: `Remove a Claude API provider configuration.

This removes the API key and any model overrides for the specified provider.
If the provider is currently active, it will be automatically reset
to the official Anthropic API.

Provider IDs support prefix matching (e.g., 'syn' matches 'synthetic').
By default, you will be prompted to confirm deletion. Use -f to skip confirmation.

Examples:
  zzk claude rm synthetic      # Remove Synthetic configuration
  zzk claude rm syn -f         # Remove with prefix matching, no confirmation`,
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

		// Check if provider is configured
		if !config.HasProvider(templateID) {
			return fmt.Errorf("provider '%s' not configured", templateID)
		}

		// Confirm deletion unless forced
		if !forceRemove {
			confirmed, err := claude.PromptYesNo(fmt.Sprintf("Remove configuration for '%s'?", tmpl.Name), false)
			if err != nil {
				return fmt.Errorf("%w. Use -f to force", err)
			}
			if !confirmed {
				fmt.Println("Cancelled")
				return nil
			}
		}

		wasActive := config.Active == templateID

		// Remove provider
		if err := config.RemoveProvider(templateID); err != nil {
			return fmt.Errorf("failed to remove provider: %w", err)
		}

		// Save config
		if err := claude.SaveConfig(config); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Provider '%s' configuration removed\n", tmpl.Name)

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
