package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func GenerateSSHKey(identity Identity) error {
	keyPath := ExpandPath(identity.SSHKeyPath())
	pubKeyPath := ExpandPath(identity.SSHPubKeyPath())

	os.Remove(keyPath)
	os.Remove(pubKeyPath)

	sshDir := filepath.Dir(keyPath)
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	cmd := exec.Command("ssh-keygen",
		"-t", "ed25519",
		"-C", identity.SSHKeyComment(),
		"-f", keyPath,
		"-N", "",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate SSH key: %w", err)
	}

	return nil
}

func SSHKeyExists(identity Identity) bool {
	keyPath := ExpandPath(identity.SSHKeyPath())
	pubKeyPath := ExpandPath(identity.SSHPubKeyPath())

	_, err1 := os.Stat(keyPath)
	_, err2 := os.Stat(pubKeyPath)

	return err1 == nil && err2 == nil
}

func IsZZKManagedKey(pubKeyPath string) (bool, string) {
	data, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return false, ""
	}

	content := string(data)
	if !strings.Contains(content, "[zzk:") {
		return false, ""
	}

	re := regexp.MustCompile(`\[zzk:([^\]]+)\]`)
	matches := re.FindStringSubmatch(content)
	if len(matches) < 2 {
		return false, ""
	}

	return true, matches[1]
}

func FindZZKManagedKeys() (map[string]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	sshDir := filepath.Join(home, ".ssh")
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("failed to read .ssh directory: %w", err)
	}

	managedKeys := make(map[string]string)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".pub") {
			continue
		}

		pubKeyPath := filepath.Join(sshDir, name)
		isManaged, identity := IsZZKManagedKey(pubKeyPath)
		if isManaged {
			keyPath := strings.TrimSuffix(pubKeyPath, ".pub")
			managedKeys[identity] = keyPath
		}
	}

	return managedKeys, nil
}

func CopyPublicKeyToHome(identity Identity) (bool, error) {
	pubKeyPath := ExpandPath(identity.SSHPubKeyPath())
	home, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("failed to get home directory: %w", err)
	}

	destPath := filepath.Join(home, fmt.Sprintf("%s_key.pub", identity.Name))

	data, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return false, fmt.Errorf("failed to read public key: %w", err)
	}

	// Check if destination exists and has same content
	if existingData, err := os.ReadFile(destPath); err == nil {
		if string(existingData) == string(data) {
			return false, nil // No change needed
		}
	}

	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return false, fmt.Errorf("failed to copy public key: %w", err)
	}

	return true, nil
}

func AddKeyToSSHAgent(identity Identity) error {
	keyPath := ExpandPath(identity.SSHKeyPath())

	exec.Command("ssh-add", "-d", keyPath).Run()

	cmd := exec.Command("ssh-add", keyPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add key to SSH agent: %w", err)
	}

	return nil
}

func TestSSHConnection(identity Identity, fromDir string) error {
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(fromDir); err != nil {
		return fmt.Errorf("failed to change to directory %s: %w", fromDir, err)
	}

	cmd := exec.Command("ssh", "-T", fmt.Sprintf("git@%s", identity.Domain))
	output, err := cmd.CombinedOutput()

	outputStr := string(output)

	successPatterns := []string{
		"successfully authenticated",
		"You've successfully authenticated",
		"Hi ",
		"Welcome to ",
	}

	for _, pattern := range successPatterns {
		if strings.Contains(outputStr, pattern) {
			return nil
		}
	}

	if strings.Contains(outputStr, "Permission denied") {
		return fmt.Errorf("permission denied - key not added to %s", identity.Domain)
	}

	if err != nil {
		return fmt.Errorf("SSH test failed: %s", outputStr)
	}

	return nil
}

func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

