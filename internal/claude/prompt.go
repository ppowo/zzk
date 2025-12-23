package claude

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// PromptYesNo prompts the user for a yes/no answer
// Returns true for yes, false for no
// Returns error if not in an interactive terminal or if reading fails
func PromptYesNo(question string, defaultYes bool) (bool, error) {
	// Check if stdin is a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false, fmt.Errorf("cannot prompt for confirmation in non-interactive mode")
	}

	// Build prompt
	prompt := question
	if defaultYes {
		prompt += " [Y/n]: "
	} else {
		prompt += " [y/N]: "
	}

	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))

	// Handle default on empty response
	if response == "" {
		return defaultYes, nil
	}

	return response == "y" || response == "yes", nil
}
