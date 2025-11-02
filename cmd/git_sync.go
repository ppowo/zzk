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
			// Check if the config file exists
			configPath := git.ConfigPath()
			if _, statErr := os.Stat(configPath); statErr == nil {
				// File exists but has errors - report them without overwriting
				fmt.Fprintf(os.Stderr, "Error in %s:\n", configPath)
				fmt.Fprintf(os.Stderr, "%v\n\n", err)
				fmt.Println("Please fix the errors in the config file and run 'zzk git sync' again.")
				fmt.Println()
				fmt.Println("Example identity structure:")
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
				os.Exit(1)
			} else {
				// File doesn't exist - create example config
				fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
				fmt.Println("Creating example configuration...")
				if err := git.CreateExampleConfig(); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to create example config: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("Created example config at: %s\n\n", configPath)
				fmt.Println("Please edit this file with your identities, then run 'zzk git sync' again.")
				os.Exit(0)
			}
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
