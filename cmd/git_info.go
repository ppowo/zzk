package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ppowo/zzk/internal/git"
	"github.com/spf13/cobra"
)

var gitInfoCmd = &cobra.Command{
	Use:   "info <identity>",
	Short: "Show detailed information about a git identity",
	Long:  `Displays detailed information about a specific git identity including SSH keys, folders, and status.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		identityName := args[0]

		config, err := git.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			fmt.Fprintf(os.Stderr, "Run 'zzk git sync' to create example config\n")
			os.Exit(1)
		}

		identity, ok := config.GetIdentity(identityName)
		if !ok {
			fmt.Fprintf(os.Stderr, "Identity '%s' not found\n\n", identityName)
			fmt.Fprintf(os.Stderr, "Available identities:\n")
			for name := range config.Identities {
				fmt.Fprintf(os.Stderr, "  - %s\n", name)
			}
			os.Exit(1)
		}

		fmt.Printf("Identity: %s\n", identity.Name)
		fmt.Printf("Domain:   %s\n", identity.Domain)
		fmt.Printf("User:     %s\n", identity.User)
		fmt.Printf("Email:    %s\n", identity.Email)
		fmt.Println()

		sshKeyPath := git.ExpandPath(identity.SSHKeyPath())
		fmt.Printf("SSH Key:        %s\n", identity.SSHKeyPath())

		if _, err := os.Stat(sshKeyPath); err == nil {
			cmd := exec.Command("ssh-keygen", "-l", "-f", sshKeyPath)
			if output, err := cmd.Output(); err == nil {
				fingerprint := strings.TrimSpace(string(output))
				parts := strings.Fields(fingerprint)
				if len(parts) >= 2 {
					fmt.Printf("  Fingerprint:  %s\n", parts[1])
				}
			}

			info, _ := os.Stat(sshKeyPath)
			fmt.Printf("  Modified:     %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("  Status:       ⚠ Not found\n")
		}

		fmt.Printf("  Public key:   %s\n", identity.SSHPubKeyPath())
		fmt.Println()

		gitConfigPath := git.ExpandPath(identity.GitConfigPath())
		fmt.Printf("Git Config:     %s\n", identity.GitConfigPath())
		if _, err := os.Stat(gitConfigPath); err == nil {
			fmt.Printf("  Status:       ✓ Exists\n")
			fmt.Printf("  Signing:      Enabled (SSH)\n")
			fmt.Printf("  SSH command:  ssh -i %s\n", identity.SSHKeyPath())
		} else {
			fmt.Printf("  Status:       ⚠ Not found\n")
		}
		fmt.Println()

		fmt.Printf("Folders (%d):\n", len(identity.Folders))
		for i, folder := range identity.Folders {
			expandedFolder := git.ExpandPath(folder)
			fmt.Printf("  %d. %s", i+1, folder)

			if _, err := os.Stat(expandedFolder); err == nil {
				repoCount := countGitRepos(expandedFolder)
				if repoCount > 0 {
					fmt.Printf("  ✓ exists (%d repos)", repoCount)
				} else {
					fmt.Printf("  ✓ exists")
				}
			} else {
				fmt.Printf("  ⚠ does not exist")
			}
			fmt.Println()
		}
		fmt.Println()

		status := "✓ Fully configured"
		if !git.SSHKeyExists(identity) {
			status = "⚠ SSH key missing"
		} else if _, err := os.Stat(gitConfigPath); os.IsNotExist(err) {
			status = "⚠ Git config missing"
		}

		fmt.Printf("Status: %s\n", status)

		if status != "✓ Fully configured" {
			fmt.Println()
			fmt.Println("Run 'zzk git sync' to fix issues")
		}
	},
}

func init() {
	gitCmd.AddCommand(gitInfoCmd)
}

func countGitRepos(dir string) int {
	count := 0
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		gitDir := filepath.Join(dir, entry.Name(), ".git")
		if _, err := os.Stat(gitDir); err == nil {
			count++
		}
	}

	return count
}
