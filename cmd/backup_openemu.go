package cmd

import (
	"github.com/spf13/cobra"
)

var backupOpenemuCmd = &cobra.Command{
	Use:   "openemu [CODE]",
	Short: "Backup and restore OpenEmu data (macOS only)",
	Long: `Backup and restore your OpenEmu application data.

Without arguments: Creates a compressed archive of OpenEmu data and uploads it
With CODE argument: Downloads and restores OpenEmu data from the uploaded archive

Examples:
  zzk backup openemu              # Upload OpenEmu data and get a code
  zzk backup openemu xyz123       # Restore OpenEmu from code xyz123`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := backupTargets["openemu"]

		// Check OS compatibility
		if err := isOSAllowed(target); err != nil {
			return err
		}

		if len(args) == 0 {
			// Upload mode
			return uploadBackup(target)
		} else {
			// Restore mode
			return restoreBackup(target, args[0])
		}
	},
}

func init() {
	backupCmd.AddCommand(backupOpenemuCmd)
}
