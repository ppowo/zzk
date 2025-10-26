package cmd

import (
	"github.com/spf13/cobra"
)

var fontInstallCmd = &cobra.Command{
	Use:   "font-install",
	Short: "Install fonts to user font directory",
	Long: `Install fonts to user font directory (no admin/sudo required).

Fonts will be installed to:
  - macOS: ~/Library/Fonts
  - Linux: ~/.local/share/fonts
  - Windows: %LOCALAPPDATA%\Microsoft\Windows\Fonts

Examples:
  zzk font-install dmca    # Install DMCA Sans Serif font`,
}

func init() {
	rootCmd.AddCommand(fontInstallCmd)
}
