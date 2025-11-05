package cmd

import (
	"fmt"

	"github.com/ppowo/zzk/internal/claude"
	"github.com/spf13/cobra"
)

var claudeUseCmd = &cobra.Command{
	Use:   "use <provider-name>",
	Short: "Switch to a Claude API provider",
	Long: `Switch to a Claude API provider and configure your shell to use it.

This command will:
1. Write the provider configuration to ~/.config/zzk/claude-env.sh
2. Mark the provider as active
3. Show instructions if shell setup is needed

After switching, you'll need to reload your shell for changes to take effect.

Examples:
  zzk claude use synthetic        # Switch to provider "synthetic"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		providerName := args[0]

		// Load config
		config, err := claude.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Check if provider exists
		provider, exists := config.GetProvider(providerName)
		if !exists {
			return fmt.Errorf("provider '%s' not found. Use 'zzk claude ls' to list providers", providerName)
		}

		if err := claude.ReloadClaudeEnvironment(providerName, provider); err != nil {
			return fmt.Errorf("failed to reload Claude environment: %w", err)
		}

		return nil
	},
}

func init() {
	claudeCmd.AddCommand(claudeUseCmd)
}
