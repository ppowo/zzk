package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ppowo/zzk/internal/git"
	"github.com/spf13/cobra"
)

var gitWhereCmd = &cobra.Command{
	Use:   "where",
	Short: "Show which git identity applies to the current directory",
	Long:  `Detects which git identity applies to the current directory based on folder patterns.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := git.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			fmt.Fprintf(os.Stderr, "Run 'zzk git sync' to create example config\n")
			os.Exit(1)
		}

		identity, err := git.GetCurrentIdentity(config)
		if err != nil {
			fmt.Printf("⚠ No identity detected for current directory\n\n")

			cwd, _ := os.Getwd()
			fmt.Printf("Current directory: %s\n\n", cwd)

			fmt.Println("You are not in a folder managed by any identity.")
			fmt.Println()
			fmt.Println("Available identities:")
			for _, id := range config.Identities {
				fmt.Printf("  %s: %s\n", id.Name, strings.Join(id.Folders, ", "))
			}
			fmt.Println()
			fmt.Println("Move your repository to one of these folders to use an identity.")
			os.Exit(1)
		}

		fmt.Printf("✓ Identity detected: %s\n\n", identity.Name)

		fmt.Printf("User:        %s\n", identity.User)
		fmt.Printf("Email:       %s\n", identity.Email)
		fmt.Printf("Domain:      %s\n", identity.Domain)
		fmt.Printf("SSH Key:     %s\n", identity.SSHKeyPath())

		cwd, _ := os.Getwd()
		matchedFolder := git.MatchingFolder(*identity, cwd)
		if matchedFolder != "" {
			fmt.Printf("Folder:      %s (matches %s/)\n", cwd, matchedFolder)
		}

		fmt.Println()
		fmt.Printf("Git config:  %s\n", identity.GitConfigPath())
		fmt.Printf("Applied via: [includeIf \"gitdir:%s/\"]\n", matchedFolder)
		fmt.Println()

		fmt.Println("Verification:")
		if isInGitRepo() {
			if verifyGitConfig(identity) {
				fmt.Println("  ✓ Git configuration matches identity")
			} else {
				fmt.Println("  ⚠ Git configuration does not match (run 'zzk git sync')")
			}
		} else {
			fmt.Println("  ℹ Not in a git repository")
		}

		if git.SSHKeyExists(*identity) {
			fmt.Println("  ✓ SSH key exists")
		} else {
			fmt.Println("  ⚠ SSH key missing (run 'zzk git sync')")
		}
	},
}

func init() {
	gitCmd.AddCommand(gitWhereCmd)
}

func isInGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	return err == nil
}

func verifyGitConfig(identity *git.Identity) bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	cmd := exec.Command("git", "config", "user.email")
	cmd.Dir = cwd
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	email := strings.TrimSpace(string(output))
	if email != identity.Email {
		return false
	}

	cmd = exec.Command("git", "config", "user.name")
	cmd.Dir = cwd
	output, err = cmd.Output()
	if err != nil {
		return false
	}

	name := strings.TrimSpace(string(output))
	if name != identity.User {
		return false
	}

	return true
}
