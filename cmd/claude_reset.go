package cmd

import (
	"fmt"

	"github.com/ppowo/zzk/internal/claude"
	"github.com/spf13/cobra"
)

var claudeResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset to official Anthropic API",
	Long: `Reset to the official Anthropic API by clearing the active provider.

This command will:
1. Clear the environment file
2. Clear the active provider in config
3. Instruct you to reload your shell

Your provider configurations will be preserved for future use.

Example:
  zzk claude reset`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		config, err := claude.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Clear env file
		if err := claude.ClearEnvFile(); err != nil {
			return fmt.Errorf("failed to clear env file: %w", err)
		}

		// Clear active provider
		wasActive := config.Active
		config.ClearActive()

		if err := claude.SaveConfig(config); err != nil {
			return fmt.Errorf("failed to update config: %w", err)
		}

		if wasActive != "" {
			fmt.Printf("✓ Cleared active provider: %s\n", wasActive)
		} else {
			fmt.Println("✓ No active provider to clear")
		}

		fmt.Println("✓ Reset to official Anthropic API")

		// Check if shell is in sync
		if needsReload, warning := claude.CheckShellSync(""); needsReload {
			fmt.Println(warning)
		}

		return nil
	},
}

func init() {
	claudeCmd.AddCommand(claudeResetCmd)
}
