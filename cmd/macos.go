package cmd

import (
	"github.com/spf13/cobra"
)

var macosCmd = &cobra.Command{
	Use:   "macos",
	Short: "macOS-specific utilities and operations",
	Long: `macOS-specific utilities and operations.

Examples:
  zzk macos vol       # Reset volume to default (17)
  zzk macos vol 50    # Set volume to 50`,
}

func init() {
	rootCmd.AddCommand(macosCmd)
}
