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

		// Write env file
		if err := claude.WriteEnvFile(provider); err != nil {
			return fmt.Errorf("failed to write env file: %w", err)
		}

		// Update active in config
		if err := config.SetActive(providerName); err != nil {
			return fmt.Errorf("failed to set active provider: %w", err)
		}

		if err := claude.SaveConfig(config); err != nil {
			return fmt.Errorf("failed to update config: %w", err)
		}

		fmt.Printf("✓ Switched to provider: %s\n", providerName)
		fmt.Printf("  Base URL: %s\n", provider.BaseURL)

		// Check if shell is in sync
		if needsReload, warning := claude.CheckShellSync(provider.BaseURL); needsReload {
			fmt.Println(warning)
		}

		// Check if RC file is set up
		isSetup, rcFile, err := claude.CheckRCFileSetup()
		if err != nil {
			// Non-fatal, just warn
			fmt.Printf("\nWarning: %v\n", err)
			return nil
		}

		if !isSetup {
			// One-time setup needed
			fmt.Println("\n⚠️  One-time setup: Add this line to your", rcFile)
			fmt.Printf("  [ -f %s ] && source %s\n", claude.EnvFilePath(), claude.EnvFilePath())
			fmt.Println("\nThen reload your shell.")
		}

		return nil
	},
}

func init() {
	claudeCmd.AddCommand(claudeUseCmd)
}
