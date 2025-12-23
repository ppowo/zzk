package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var claudeCmd = &cobra.Command{
	Use:   "claude",
	Short: "Manage Claude API providers for Claude Code",
	Long: `Manage Claude API providers to switch between different API-compatible services.

Claude Code uses Anthropic as the default provider, but you can configure and
switch between alternative providers like Synthetic or other API-compatible services.

Provider IDs support prefix matching (e.g., 'syn' matches 'synthetic').

Configuration file: ~/.claude-providers.json
Environment file: ~/.config/zzk/claude-env.sh

Examples:
  zzk claude ls                   # List providers (shows active)
  zzk claude set synthetic        # Configure a provider (add or update)
  zzk claude use syn              # Switch to a provider (prefix matching)
  zzk claude reset                # Reset to official Anthropic
  zzk claude rm synthetic         # Remove a provider`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if runtime.GOOS == "windows" {
			return fmt.Errorf("this command is not supported on Windows - it requires Unix-style shell environment management")
		}

		// Show help if no subcommand provided
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(claudeCmd)
}
