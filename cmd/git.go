package cmd

import (
	"github.com/spf13/cobra"
)

var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "Git identity manager - manage multiple git identities",
	Long: `Git identity manager helps you manage multiple git identities for different services.

Features:
  - Folder-based identity selection
  - SSH key generation and management
  - Automatic commit signing
  - HTTPS to SSH URL rewriting
  - Multiple identities per domain (e.g., work and personal GitHub)

Configuration file: ~/.git-identities.json

Examples:
  zzk git sync                    # Apply configuration and cleanup orphans
  zzk git status                  # Show status of all identities
  zzk git where                   # Show current identity
  zzk git info github-work        # Show identity details`,
}

func init() {
	rootCmd.AddCommand(gitCmd)
}
