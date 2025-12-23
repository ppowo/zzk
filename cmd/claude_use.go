package cmd

import (
	"fmt"

	"github.com/ppowo/zzk/internal/claude"
	"github.com/spf13/cobra"
)

var claudeUseCmd = &cobra.Command{
	Use:   "use <provider>",
	Short: "Switch to a Claude API provider",
	Long: `Switch to a Claude API provider and configure your shell to use it.

This command will:
1. Write the provider configuration to ~/.config/zzk/claude-env.sh
2. Mark the provider as active
3. Show instructions if shell setup is needed

Provider IDs support prefix matching (e.g., 'syn' matches 'synthetic').

Examples:
  zzk claude use synthetic    # Switch to Synthetic provider
  zzk claude use syn          # Same (prefix matching)
  zzk claude use openrouter   # Switch to OpenRouter provider`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Resolve prefix to full template ID
		templateID, err := claude.ResolveTemplateID(args[0])
		if err != nil {
			return err
		}

		// Load config
		config, err := claude.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Check if provider is configured
		provider, exists := config.GetProvider(templateID)
		if !exists {
			return fmt.Errorf("provider '%s' not configured. Use 'zzk claude set %s' to configure it",
				templateID, templateID)
		}

		if err := claude.ReloadClaudeEnvironment(templateID, provider); err != nil {
			return fmt.Errorf("failed to reload Claude environment: %w", err)
		}

		return nil
	},
}

func init() {
	claudeCmd.AddCommand(claudeUseCmd)
}
