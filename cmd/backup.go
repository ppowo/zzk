package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

const backupServiceURL = "https://envs.sh"

// Global exclusion patterns applied to all backups
var globalExcludeGlobs = []string{
	"*.DS_Store",
	"._*",
	"Thumbs.db",
	"desktop.ini",
	"*.swp",
	"*.swo",
	"*~",
	".Spotlight-V100",
	".Trashes",
	".fseventsd",
	".TemporaryItems",
	"__pycache__",
	"*.pyc",
	".git",
	".svn",
	"node_modules",
	".claude",
	".claude/",
	"*.claude",
}

// BackupTarget defines a backup target configuration
type BackupTarget struct {
	Name         string   // Name of the target (e.g., "bio", "openemu")
	Path         string   // Path relative to home (e.g., ".bio", "Library/Application Support/OpenEmu")
	AllowedOS    []string // Allowed operating systems (darwin, linux, windows)
	BackupPrefix string   // Prefix for backup directories (e.g., ".bio.backup-")
	KeepBackups  int      // Number of backups to keep
}

var backupTargets = map[string]BackupTarget{
	"bio": {
		Name:         "bio",
		Path:         ".bio",
		AllowedOS:    []string{"darwin", "linux"},
		BackupPrefix: ".bio.backup-",
		KeepBackups:  3,
	},
	"openemu": {
		Name:         "openemu",
		Path:         "Library/Application Support/OpenEmu",
		AllowedOS:    []string{"darwin"},
		BackupPrefix: ".openemu.backup-",
		KeepBackups:  3,
	},
}

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup and restore operations",
	Long: `Backup and restore various directories and configurations.

Available targets:
  bio      - Backup/restore ~/.bio directory (macOS/Linux)
  openemu  - Backup/restore OpenEmu data (macOS only)

Examples:
  zzk backup bio              # Upload .bio and get a code
  zzk backup bio a1b2c3       # Restore .bio from code a1b2c3
  zzk backup openemu          # Upload OpenEmu data
  zzk backup openemu xyz123   # Restore OpenEmu from code xyz123`,
}

func init() {
	rootCmd.AddCommand(backupCmd)
}

// isOSAllowed checks if the current OS is allowed for a target
func isOSAllowed(target BackupTarget) error {
	currentOS := runtime.GOOS
	for _, allowedOS := range target.AllowedOS {
		if currentOS == allowedOS {
			return nil
		}
	}
	return fmt.Errorf("%s backup is only supported on %v (current OS: %s)",
		target.Name, target.AllowedOS, currentOS)
}
