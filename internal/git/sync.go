package git

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type SyncResult struct {
	OrphansRemoved []string
	Created        []string
	Updated        []string
	Verified       []string
	Failed         map[string]error
}

func Sync(config *Config) (*SyncResult, error) {
	result := &SyncResult{
		OrphansRemoved: []string{},
		Created:        []string{},
		Updated:        []string{},
		Verified:       []string{},
		Failed:         make(map[string]error),
	}

	fmt.Println("Reading config:", ConfigPath())
	fmt.Printf("Found %d identities: %s\n\n", len(config.Identities), identityNames(config))

	fmt.Println("Detecting orphans...")
	orphans, err := detectOrphans(config)
	if err != nil {
		return nil, fmt.Errorf("failed to detect orphans: %w", err)
	}

	if len(orphans) > 0 {
		fmt.Printf("  Found %d orphaned identities: %s\n", len(orphans), strings.Join(orphans, ", "))
		for _, orphan := range orphans {
			if err := cleanupIdentity(orphan); err != nil {
				fmt.Printf("  ⚠ Warning: failed to clean up %s: %v\n", orphan, err)
			} else {
				fmt.Printf("  ✓ Removed orphan: %s\n", orphan)
				result.OrphansRemoved = append(result.OrphansRemoved, orphan)
			}
		}
	} else {
		fmt.Println("  No orphans found")
	}
	fmt.Println()

	for _, identity := range config.Identities {
		fmt.Printf("Processing: %s\n", identity.Name)

		for _, folder := range identity.Folders {
			expandedFolder := expandPath(folder)
			if err := os.MkdirAll(expandedFolder, 0755); err != nil {
				fmt.Printf("  ⚠ Warning: failed to create folder %s: %v\n", folder, err)
			} else {
				if _, err := os.Stat(expandedFolder); err == nil {
					fmt.Printf("  ✓ Folder exists: %s\n", folder)
				} else {
					fmt.Printf("  ✓ Created folder: %s\n", folder)
				}
			}
		}

		keyWasCreated := false
		if !SSHKeyExists(identity) {
			if err := GenerateSSHKey(identity); err != nil {
				fmt.Printf("  ✗ Failed to generate SSH key: %v\n", err)
				result.Failed[identity.Name] = err
				fmt.Println()
				continue
			}
			fmt.Printf("  ✓ Generated SSH key: %s [zzk:%s]\n", identity.SSHKeyPath(), identity.Name)
			result.Created = append(result.Created, identity.Name)
			keyWasCreated = true
		} else {
			fmt.Printf("  ✓ SSH key exists: %s [zzk:%s]\n", identity.SSHKeyPath(), identity.Name)
		}

		// Only copy public key if a new key was just created
		if keyWasCreated {
			copied, err := CopyPublicKeyToHome(identity)
			if err != nil {
				fmt.Printf("  ⚠ Warning: failed to copy public key: %v\n", err)
			} else if copied {
				fmt.Printf("  ✓ Copied public key to ~/%s_key.pub\n", identity.Name)
			}
		}

		if err := CreateIdentityGitConfig(identity); err != nil {
			fmt.Printf("  ✗ Failed to create git config: %v\n", err)
			result.Failed[identity.Name] = err
			fmt.Println()
			continue
		}
		fmt.Printf("  ✓ Updated %s\n", identity.GitConfigPath())

		if err := AddKeyToSSHAgent(identity); err != nil {
			fmt.Printf("  ⚠ Warning: failed to add key to SSH agent: %v\n", err)
		} else {
			fmt.Printf("  ✓ Added key to SSH agent\n")
		}

		var testFromDir string
		for _, folder := range identity.Folders {
			expandedFolder := expandPath(folder)
			if _, err := os.Stat(expandedFolder); err == nil {
				testFromDir = expandedFolder
				break
			}
		}

		if testFromDir != "" {
			fmt.Printf("  Testing SSH connection to %s...\n", identity.Domain)
			if err := TestSSHConnection(identity, testFromDir); err != nil {
				fmt.Printf("  ⚠ SSH test failed: %v\n", err)
				fmt.Printf("    → Your SSH key may not be added to %s yet\n", identity.Domain)
				fmt.Printf("    → Add it: cat %s | pbcopy\n", identity.SSHPubKeyPath())
			} else {
				fmt.Printf("  ✓ SSH connection verified\n")
				result.Verified = append(result.Verified, identity.Name)
			}
		} else {
			fmt.Printf("  ⚠ SSH test skipped (no valid folders)\n")
		}

		fmt.Println()
	}

	fmt.Println("Updating global configurations...")
	if err := UpdateGlobalGitConfig(config); err != nil {
		return nil, fmt.Errorf("failed to update global git config: %w", err)
	}
	fmt.Println("  ✓ Updated ~/.gitconfig")

	if err := UpdateSSHConfig(config); err != nil {
		return nil, fmt.Errorf("failed to update SSH config: %w", err)
	}
	fmt.Println("  ✓ Updated ~/.ssh/config")

	if err := UpdateAllowedSigners(config); err != nil {
		return nil, fmt.Errorf("failed to update allowed signers: %w", err)
	}
	fmt.Println("  ✓ Updated ~/.ssh/allowed_signers")
	fmt.Println()
	printSyncSummary(result)

	return result, nil
}

func detectOrphans(config *Config) ([]string, error) {
	orphans := []string{}

	managedKeys, err := FindZZKManagedKeys()
	if err != nil {
		return nil, err
	}

	for identity := range managedKeys {
		if !config.HasIdentity(identity) {
			orphans = append(orphans, identity)
		}
	}

	managedConfigs, err := FindZZKManagedGitConfigs()
	if err != nil {
		return nil, err
	}

	for identity := range managedConfigs {
		if !config.HasIdentity(identity) {
			found := slices.Contains(orphans, identity)
			if !found {
				orphans = append(orphans, identity)
			}
		}
	}

	return orphans, nil
}

func cleanupIdentity(identityName string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	sshKeyPath := filepath.Join(home, ".ssh", fmt.Sprintf("%s_key", identityName))
	sshPubKeyPath := sshKeyPath + ".pub"
	os.Remove(sshKeyPath)
	os.Remove(sshPubKeyPath)

	homePubKey := filepath.Join(home, fmt.Sprintf("%s_key.pub", identityName))
	os.Remove(homePubKey)

	gitConfigPath := filepath.Join(home, fmt.Sprintf(".gitconfig-%s", identityName))
	os.Remove(gitConfigPath)

	caser := cases.Title(language.English)
	serviceFolder := filepath.Join(home, caser.String(identityName))
	os.Remove(serviceFolder) // Only succeeds if empty

	return nil
}

func identityNames(config *Config) string {
	names := []string{}
	for name := range config.Identities {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}

func printSyncSummary(result *SyncResult) {
	fmt.Println("Sync complete!")
	fmt.Println()

	if len(result.OrphansRemoved) > 0 {
		fmt.Printf("Orphans removed: %d\n", len(result.OrphansRemoved))
	}
	if len(result.Created) > 0 {
		fmt.Printf("Identities created: %d\n", len(result.Created))
	}
	if len(result.Verified) > 0 {
		fmt.Printf("SSH connections verified: %d\n", len(result.Verified))
	}
	if len(result.Failed) > 0 {
		fmt.Printf("Failed: %d\n", len(result.Failed))
		for identity, err := range result.Failed {
			fmt.Printf("  - %s: %v\n", identity, err)
		}
	}
	needsKeyUpload := len(result.Created) > 0

	if needsKeyUpload {
		fmt.Println()
		fmt.Println("Next steps for new identities:")
		fmt.Println("1. Add your public keys to your accounts")
		fmt.Println("2. Run 'zzk git sync' again to verify connections")
	}
}
