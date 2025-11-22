package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var backupBioCmd = &cobra.Command{
	Use:   "bio [CODE]",
	Short: "Backup and restore .bio directory",
	Long: `Backup and restore your ~/.bio directory.

Without arguments: Creates a compressed archive of ~/.bio and uploads it
With CODE argument: Downloads and restores .bio from the uploaded archive

Examples:
  zzk backup bio              # Upload .bio and get a code
  zzk backup bio a1b2c3       # Restore .bio from code a1b2c3`,
	Args: cobra.ArbitraryArgs,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		target := backupTargets["bio"]

		// Check OS compatibility
		if err := isOSAllowed(target); err != nil {
			return err
		}

		if len(args) == 0 {
			// Upload mode
			return uploadBackup(target)
		} else if len(args) == 1 {
			// Restore mode
			return restoreBackup(target, args[0])
		} else {
			return fmt.Errorf("too many arguments")
		}
	},
}

func init() {
	backupCmd.AddCommand(backupBioCmd)
}
