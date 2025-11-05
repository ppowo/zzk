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
		if err := claude.ResetToOfficialAPI(); err != nil {
			return fmt.Errorf("failed to reset to official API: %w", err)
		}

		return nil
	},
}

func init() {
	claudeCmd.AddCommand(claudeResetCmd)
}
