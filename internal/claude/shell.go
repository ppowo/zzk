package claude

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"al.essio.dev/pkg/shellescape"
	"github.com/ppowo/zzk/internal/fileutil"
)

// EnvFilePath returns the path to the environment file
func EnvFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home unavailable
		cwd, _ := os.Getwd()
		if cwd != "" {
			return filepath.Join(cwd, ".config", "zzk", "claude-env.sh")
		}
		// Last resort fallback
		return filepath.Join("/tmp", "zzk-claude-env.sh")
	}
	return filepath.Join(home, ".config", "zzk", "claude-env.sh")
}

// WriteEnvFile writes the provider configuration to the env file
func WriteEnvFile(provider Provider) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString("# Managed by zzk - do not edit manually\n")
	buf.WriteString("# Generated for Claude Code provider configuration\n\n")
	buf.WriteString(provider.ToShellExports())

	return fileutil.AtomicWrite(EnvFilePath(), buf.Bytes(), 0600)
}

// ClearEnvFile clears the environment file
func ClearEnvFile() error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	// Write unset commands to clear any existing environment variables
	content := `# No active provider - using official Anthropic API
# Unset any previously set Claude variables
unset ANTHROPIC_BASE_URL
unset ANTHROPIC_AUTH_TOKEN
unset ANTHROPIC_DEFAULT_OPUS_MODEL
unset ANTHROPIC_DEFAULT_SONNET_MODEL
unset ANTHROPIC_DEFAULT_HAIKU_MODEL
unset CLAUDE_CODE_SUBAGENT_MODEL
unset CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC
`

	return fileutil.AtomicWrite(EnvFilePath(), []byte(content), 0600)
}

// DetectShell detects the current shell
func DetectShell() string {
	// Check $SHELL
	if shell := os.Getenv("SHELL"); shell != "" {
		return filepath.Base(shell)
	}

	// Try to detect from parent process on Unix
	if ppid := os.Getppid(); ppid > 0 {
		// On Linux, could check /proc/$PPID/exe for more accurate detection
		// For now, we rely on $SHELL being set
		_ = ppid // Use ppid to avoid unused variable
	}

	// Final fallback to bash (more common than sh)
	return "bash"
}

// GetRCFilePath returns the RC file path for the given shell
func GetRCFilePath(shell string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	switch shell {
	case "zsh":
		return filepath.Join(home, ".zshrc")
	case "bash":
		// Check for .bashrc first, fall back to .bash_profile on macOS
		bashrc := filepath.Join(home, ".bashrc")
		if _, err := os.Stat(bashrc); err == nil {
			return bashrc
		}
		return filepath.Join(home, ".bash_profile")
	case "fish":
		return filepath.Join(home, ".config", "fish", "config.fish")
	default:
		return ""
	}
}

// CheckRCFileSetup checks if the RC file has the source line
func CheckRCFileSetup() (bool, string, error) {
	shell := DetectShell()
	rcFile := GetRCFilePath(shell)

	if rcFile == "" {
		return false, "", fmt.Errorf("unsupported shell: %s", shell)
	}

	// Try to read RC file directly (no TOCTOU race)
	data, err := os.ReadFile(rcFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, rcFile, nil
		}
		return false, rcFile, fmt.Errorf("failed to read RC file: %w", err)
	}

	envPath := EnvFilePath()
	// Check for actual source command, not just substring match
	// This prevents false positives from comments or other strings
	lines := strings.SplitSeq(string(data), "\n")
	for line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip comments
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Check for source command with our env file (exact match with word boundaries)
		// Match patterns:
		// - "source <path>", ". <path>"
		// - "source <path> ", ". <path> " (with trailing space/semicolon)
		// - "[ -f <path> ] && source <path>" (conditional sourcing)
		sourceLine := fmt.Sprintf("source %s", envPath)
		dotSourceLine := fmt.Sprintf(". %s", envPath)
		conditionalSourceLine := fmt.Sprintf("[ -f %s ] && source %s", envPath, envPath)

		if trimmed == sourceLine || strings.HasPrefix(trimmed, sourceLine+" ") ||
			strings.HasPrefix(trimmed, sourceLine+";") ||
			trimmed == dotSourceLine || strings.HasPrefix(trimmed, dotSourceLine+" ") ||
			strings.HasPrefix(trimmed, dotSourceLine+";") ||
			trimmed == conditionalSourceLine || strings.HasPrefix(trimmed, conditionalSourceLine+" ") ||
			strings.HasPrefix(trimmed, conditionalSourceLine+";") {
			return true, rcFile, nil
		}
	}

	return false, rcFile, nil
}

// GetSourceLine returns the appropriate source line for the current shell
func GetSourceLine() string {
	shell := DetectShell()
	envPath := EnvFilePath()

	if shell == "fish" {
		return fmt.Sprintf("[ -f %s ]; and source %s", envPath, envPath)
	}
	return fmt.Sprintf("[ -f %s ] && source %s", envPath, envPath)
}

// shellQuote quotes a string for safe use in shell commands using POSIX shell quoting
func shellQuote(s string) string {
	return shellescape.Quote(s)
}

// ShowSetupInstructions shows instructions for setting up the RC file
func ShowSetupInstructions() {
	shell := DetectShell()
	rcFile := GetRCFilePath(shell)
	sourceLine := GetSourceLine()

	fmt.Println("\n⚠️  One-time setup required:")
	fmt.Println("Add this line to your shell configuration file:")
	fmt.Printf("\n  %s\n\n", sourceLine)

	if rcFile != "" {
		fmt.Printf("Your shell config file: %s\n", rcFile)
		fmt.Println("\nYou can add it manually or run:")
		// Properly quote for shell safety
		fmt.Printf("  grep -q 'claude-env.sh' %s || echo %s >> %s\n",
			shellQuote(rcFile),
			shellQuote(sourceLine),
			shellQuote(rcFile))
		fmt.Println("\nAfter adding the line, reload your shell:")
		fmt.Printf("  source %s\n", shellQuote(rcFile))
	} else {
		fmt.Println("Add this to your shell's configuration file and reload.")
	}
}

// GetReloadInstructions returns instructions for reloading the shell environment
func GetReloadInstructions() string {
	shell := DetectShell()
	rcFile := GetRCFilePath(shell)

	if rcFile == "" {
		rcFile = "your shell config file"
	}

	return fmt.Sprintf(`⚠️  ACTION REQUIRED: Reload your shell to apply changes

Run this command:
  source %s`, rcFile)
}

// ResetToOfficialAPI resets the Claude environment to use the official Anthropic API.
// It clears the env file, updates the config, checks shell sync, and shows RC file setup warnings if needed.
func ResetToOfficialAPI() error {
	// Load config
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Clear env file
	if err := ClearEnvFile(); err != nil {
		return fmt.Errorf("failed to clear env file: %w", err)
	}

	// Clear active provider
	wasActive := config.Active
	config.ClearActive()

	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	if wasActive != "" {
		fmt.Printf("✓ Cleared active provider: %s\n", wasActive)
	} else {
		fmt.Println("✓ No active provider to clear")
	}

	fmt.Println("✓ Reset to official Anthropic API")
	fmt.Println(GetReloadInstructions())

	// Check if RC file is set up
	isSetup, rcFile, err := CheckRCFileSetup()
	if err != nil {
		// Non-fatal, just warn
		fmt.Printf("\nWarning: %v\n", err)
		return nil
	}

	if !isSetup {
		// One-time setup needed
		fmt.Println("\n⚠️  One-time setup: Add this line to your", rcFile)
		fmt.Printf("  [ -f %s ] && source %s\n", EnvFilePath(), EnvFilePath())
		fmt.Println("\nThen reload your shell.")
	}

	return nil
}

// ReloadClaudeEnvironment reloads the Claude environment when a provider is edited.
// It writes the env file, checks shell sync, and shows warnings if needed.
func ReloadClaudeEnvironment(providerName string, provider Provider) error {
	// Write env file
	if err := WriteEnvFile(provider); err != nil {
		return fmt.Errorf("failed to write env file: %w", err)
	}

	// Update active in config
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := config.SetActive(providerName); err != nil {
		return fmt.Errorf("failed to set active provider: %w", err)
	}

	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	fmt.Printf("✓ Switched to provider: %s\n", providerName)
	fmt.Printf("  Base URL: %s\n", provider.BaseURL)
	fmt.Println(GetReloadInstructions())

	// Check if RC file is set up
	isSetup, rcFile, err := CheckRCFileSetup()
	if err != nil {
		// Non-fatal, just warn
		fmt.Printf("\nWarning: %v\n", err)
		return nil
	}

	if !isSetup {
		// One-time setup needed
		fmt.Println("\n⚠️  One-time setup: Add this line to your", rcFile)
		fmt.Printf("  [ -f %s ] && source %s\n", EnvFilePath(), EnvFilePath())
		fmt.Println("\nThen reload your shell.")
	}

	return nil
}
