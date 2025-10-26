package cmd

import (
	"fmt"
	"os"

	"github.com/ppowo/zzk/internal/git"
	"github.com/spf13/cobra"
)

var gitSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize git identities from config file",
	Long: `Reads ~/.git-identities.json and synchronizes your system:
  - Creates/updates SSH keys
  - Updates git configs
  - Cleans up orphaned identities
  - Verifies SSH connections

Run this command after editing ~/.git-identities.json`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := git.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
			fmt.Println("Creating example configuration...")
			if err := git.CreateExampleConfig(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create example config: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Created example config at: %s\n\n", git.ConfigPath())
			fmt.Println("Please edit this file with your identities, then run 'zzk git sync' again.")
			fmt.Println()
			fmt.Println("Example identity:")
			fmt.Println(`{
  "identities": {
    "github-work": {
      "user": "your-username",
      "email": "work@company.com",
      "domain": "github.com",
      "folders": ["~/Work/Github"]
    }
  }
}`)
			os.Exit(0)
		}

		_, err = git.Sync(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Sync failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	gitCmd.AddCommand(gitSyncCmd)
}
