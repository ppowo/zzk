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

// EditProviderWithRetry opens the editor to create/edit a provider,
// retrying on validation failures if the user wants to try again
func EditProviderWithRetry(existing *Provider) (*Provider, error) {
	for {
		provider, err := EditProvider(existing)
		if err == nil && provider == nil {
			// User cancelled (no modifications made)
			return nil, fmt.Errorf("no changes made")
		}
		if err == nil {
			return provider, nil
		}

		// Check if it's a validation error
		if !strings.Contains(err.Error(), "validation failed") {
			return nil, err
		}

		// Show the error
		fmt.Printf("\n%v\n", err)

		// Ask if they want to retry
		retry, retryErr := PromptYesNo("Would you like to try again?", true)
		if retryErr != nil {
			return nil, fmt.Errorf("prompt failed: %w", retryErr)
		}

		if !retry {
			return nil, fmt.Errorf("operation cancelled")
		}
	}
}
