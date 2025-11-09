package scaffold

import (
	"fmt"
	"os"

	"github.com/joeblew999/wellknown/pkg/env"
)

// GitHooksOptions configures git hooks installation.
type GitHooksOptions struct {
	HookPath        string      // Path to install hook (default: ".git/hooks/pre-commit")
	OverwritePrompt func() bool // Optional: callback to prompt for overwrite confirmation
}

// GitHooksResult contains the result of git hooks installation.
type GitHooksResult struct {
	HookPath  string // Path where hook was installed
	Installed bool   // Whether hook was installed (false if aborted)
}

// InstallGitHooks installs a pre-commit hook that prevents committing plaintext secrets.
//
// The hook blocks commits of:
//   - Plaintext .env files (.env.local, .env.production, .env.secrets.*)
//   - Age encryption keys (path from env.DefaultAgeKeyPath)
//   - Allows encrypted *.age files
//
// Example:
//
//	result, err := scaffold.InstallGitHooks(scaffold.GitHooksOptions{
//	    OverwritePrompt: func() bool {
//	        fmt.Print("Overwrite existing hook? (y/N): ")
//	        var response string
//	        fmt.Scanln(&response)
//	        return response == "y" || response == "Y"
//	    },
//	})
func InstallGitHooks(opts GitHooksOptions) (*GitHooksResult, error) {
	// Set defaults
	if opts.HookPath == "" {
		opts.HookPath = ".git/hooks/pre-commit"
	}

	// Pre-commit hook content (using DefaultAgeKeyPath)
	preCommitHook := fmt.Sprintf(`#!/bin/bash
# Pre-commit hook to prevent committing plaintext secrets

# Check for plaintext .env files (not *.age)
if git diff --cached --name-only | grep -E "^\.env\.(local|production)$|^\.env\.secrets\.(local|production)$|^\.env\.secrets$" | grep -v "\.age$"; then
  echo "❌ ERROR: Attempting to commit plaintext secrets!"
  echo "   Plaintext .env files must NOT be committed"
  echo "   Did you mean to add the .age files instead?"
  echo ""
  echo "   Run: go run . age-encrypt"
  echo "   Then: git add *.age"
  exit 1
fi

# Check for age encryption key
if git diff --cached --name-only | grep -E "%s|\.age-key\.txt"; then
  echo "❌ ERROR: Attempting to commit age encryption key!"
  echo "   The %s file must NEVER be committed"
  echo "   This would expose all your encrypted secrets!"
  exit 1
fi

exit 0
`, env.DefaultAgeKeyPath, env.DefaultAgeKeyPath)

	// Check if .git directory exists
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		return nil, fmt.Errorf("not a git repository (no .git directory found)")
	}

	// Check if hook already exists
	if _, err := os.Stat(opts.HookPath); err == nil {
		// Hook exists - check if we should overwrite
		if opts.OverwritePrompt != nil {
			if !opts.OverwritePrompt() {
				return &GitHooksResult{
					HookPath:  opts.HookPath,
					Installed: false,
				}, nil
			}
		} else {
			// No prompt provided - don't overwrite
			return nil, fmt.Errorf("hook already exists at %s", opts.HookPath)
		}
	}

	// Create hooks directory if it doesn't exist
	hooksDir := ".git/hooks"
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create hooks directory: %w", err)
	}

	// Write pre-commit hook with executable permissions
	if err := os.WriteFile(opts.HookPath, []byte(preCommitHook), 0755); err != nil {
		return nil, fmt.Errorf("failed to write pre-commit hook: %w", err)
	}

	return &GitHooksResult{
		HookPath:  opts.HookPath,
		Installed: true,
	}, nil
}
