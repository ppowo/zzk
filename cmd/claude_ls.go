package cmd

import (
	"fmt"

	"github.com/ppowo/zzk/internal/claude"
	"github.com/spf13/cobra"
)

var claudeLsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List all Claude API providers",
	Long: `List all available Claude API providers and their configuration status.

Shows all provider templates with:
  * - currently active provider
  + - configured (has API key)
  - - not configured

Example:
  zzk claude ls`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		config, err := claude.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Show active provider
		if config.Active != "" {
			if tmpl, ok := claude.GetTemplate(config.Active); ok {
				fmt.Printf("Active: %s (%s)\n\n", tmpl.Name, tmpl.BaseURL)
			} else {
				fmt.Printf("Active: %s\n\n", config.Active)
			}
		} else {
			fmt.Println("Active: Official Anthropic API")
			fmt.Println()
		}

		// Show all templates with status
		fmt.Println("Providers:")
		for _, tmpl := range claude.ListTemplates() {
			marker := "-"
			status := "not configured"

			if config.HasProvider(tmpl.ID) {
				marker = "+"
				status = "configured"
				if tmpl.ID == config.Active {
					marker = "*"
					status = "active"
				}
			}

			fmt.Printf("  %s %-12s %-15s %s\n", marker, tmpl.ID, "("+status+")", tmpl.BaseURL)
		}

		// Show help for unconfigured providers
		var unconfigured []string
		for _, tmpl := range claude.ListTemplates() {
			if !config.HasProvider(tmpl.ID) {
				unconfigured = append(unconfigured, tmpl.ID)
			}
		}

		if len(unconfigured) > 0 {
			fmt.Println("\nTo configure a provider:")
			fmt.Printf("  zzk claude set <provider>\n")
		}

		return nil
	},
}

func init() {
	claudeCmd.AddCommand(claudeLsCmd)
}
