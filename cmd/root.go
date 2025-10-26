package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zzk",
	Short: "A swiss army knife CLI toolkit",
	Long: `zzk is a command-line swiss army knife with diverse functionality.

Currently includes:
  - Backup/restore .bio directory
  - Claude API provider management for Claude Code
  - Git identity management for multiple services
  - Media downloading (YouTube) with aria2c acceleration
  - Automatic screen resolution detection for video quality
  - Font installation utilities

Examples:
  zzk backup                                    # Upload .bio and get a code
  zzk backup a1b2c3                             # Restore .bio from code
  zzk claude add synthetic                      # Add Claude provider
  zzk claude use synthetic                      # Switch to provider
  zzk git sync                                  # Sync git identities
  zzk git ls                                    # List identities
  zzk yt aud https://youtube.com/watch?v=...    # Download audio
  zzk yt alb https://youtube.com/playlist?...   # Download album/playlist
  zzk yt vid https://youtube.com/watch?v=...    # Download video
  zzk font-install dmca                         # Install DMCA Sans Serif font`,
}

var UseTmpDir bool

func init() {
	rootCmd.PersistentFlags().BoolVar(&UseTmpDir, "tmp", false, "Use temporary directory for operations")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
