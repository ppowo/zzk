package cmd

import (
	"fmt"
	"sort"

	"github.com/ppowo/zzk/internal/claude"
	"github.com/spf13/cobra"
)


var claudeLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all Claude API providers",
	Long: `List all configured Claude API providers.

Shows provider names with an asterisk (*) marking the currently active provider.

Example:
  zzk claude ls`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		config, err := claude.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if len(config.Providers) == 0 {
			fmt.Println("No providers configured.")
			fmt.Println("\nTo add a provider, run:")
			fmt.Println("  zzk claude add <provider-name>")
			return nil
		}

		// Sort provider names for consistent output
		names := make([]string, 0, len(config.Providers))
		for name := range config.Providers {
			names = append(names, name)
		}
		sort.Strings(names)

		// Show active provider in config
		if config.Active != "" {
			fmt.Printf("Active: %s\n\n", config.Active)
		} else {
			fmt.Println("Active: none (using official Anthropic API)")
		}

		// Show all providers list
		fmt.Printf("Providers (%d):\n", len(config.Providers))
		for _, name := range names {
			provider := config.Providers[name]
			marker := " "
			if name == config.Active {
				marker = "*"
			}
			fmt.Printf("  %s %s\n", marker, name)
			fmt.Printf("    %s\n", provider.BaseURL)
		}

		return nil
	},
}

func init() {
	claudeCmd.AddCommand(claudeLsCmd)
}
