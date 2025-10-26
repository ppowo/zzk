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
		// Match patterns: "source <path>", ". <path>", "source <path> ", ". <path> "
		if (trimmed == "source "+envPath || strings.HasPrefix(trimmed, "source "+envPath+" ") ||
		    strings.HasPrefix(trimmed, "source "+envPath+";") ||
		    trimmed == ". "+envPath || strings.HasPrefix(trimmed, ". "+envPath+" ") ||
		    strings.HasPrefix(trimmed, ". "+envPath+";")) {
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

// CheckShellSync checks if the shell environment matches the expected state
// Returns true if shell needs reloading, along with a warning message
func CheckShellSync(expectedBaseURL string) (needsReload bool, warning string) {
	currentBaseURL := os.Getenv("ANTHROPIC_BASE_URL")

	if expectedBaseURL == "" {
		// We expect no variables set (after reset or no active provider)
		if currentBaseURL != "" {
			return true, "⚠️  Shell has old variables. Run 'ss' to reload."
		}
		return false, ""
	}

	// We expect specific provider to be active
	if currentBaseURL == "" {
		return true, "⚠️  Shell not reloaded. Run 'ss' to apply changes."
	}

	if currentBaseURL != expectedBaseURL {
		return true, "⚠️  Shell has different provider. Run 'ss' to apply changes."
	}

	return false, ""
}
