package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/ppowo/zzk/internal/git"
	"github.com/spf13/cobra"
)

var gitStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of all git identities",
	Long:  `Shows the status of all git identities from ~/.git-identities.json with their last sync time.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := git.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			fmt.Fprintf(os.Stderr, "Run 'zzk git sync' to create example config\n")
			os.Exit(1)
		}

		// Load state to get last sync times
		state, err := git.LoadState()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not load state: %v\n", err)
			state = nil
		}

		if len(config.Identities) == 0 {
			fmt.Println("No identities configured")
			return
		}

		fmt.Printf("%-20s %-15s %-25s %-15s %-20s %-15s %s\n",
			"IDENTITY", "USER", "EMAIL", "DOMAIN", "FOLDERS", "STATUS", "LAST SYNC")
		fmt.Println(strings.Repeat("-", 135))

		for _, identity := range config.Identities {
			status := getIdentityStatus(identity)

			// Get last sync time from state
			lastSync := "Never"
			if state != nil {
				if identityState, ok := state.Identities[identity.Name]; ok {
					if !identityState.LastSync.IsZero() {
						lastSync = humanize.Time(identityState.LastSync)
					}
				}
			}

			firstFolder := ""
			if len(identity.Folders) > 0 {
				firstFolder = identity.Folders[0]
			}

			fmt.Printf("%-20s %-15s %-25s %-15s %-20s %-15s %s\n",
				identity.Name,
				truncate(identity.User, 15),
				truncate(identity.Email, 25),
				identity.Domain,
				truncate(firstFolder, 20),
				status,
				lastSync)

			for i := 1; i < len(identity.Folders); i++ {
				fmt.Printf("%-20s %-15s %-25s %-15s %-20s\n",
					"", "", "", "", truncate(identity.Folders[i], 20))
			}
		}

		fmt.Println()

		// Print summary
		activeCount := 0
		for _, identity := range config.Identities {
			if git.SSHKeyExists(identity) {
				activeCount++
			}
		}

		if state != nil && !state.LastSync.IsZero() {
			fmt.Printf("Summary: %d identities active | Last global sync: %s\n",
				activeCount, humanize.Time(state.LastSync))
		} else {
			fmt.Printf("Summary: %d identities active | Never synced\n", activeCount)
		}

		fmt.Println()
		fmt.Println("Status Legend:")
		fmt.Println("  ✓ Active       - Fully configured and ready")
		fmt.Println("  ⚠ Key missing  - SSH key not found (run: zzk git sync)")
		fmt.Println("  ✗ Config error - Git config file missing or invalid")
	},
}

func init() {
	gitCmd.AddCommand(gitStatusCmd)
}

func getIdentityStatus(identity git.Identity) string {
	if !git.SSHKeyExists(identity) {
		return "⚠ Key missing"
	}

	gitConfigPath := git.ExpandPath(identity.GitConfigPath())
	if _, err := os.Stat(gitConfigPath); os.IsNotExist(err) {
		return "✗ Config error"
	}

	return "✓ Active"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
