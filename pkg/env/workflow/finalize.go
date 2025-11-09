package workflow

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/joeblew999/wellknown/pkg/env"
)

// FinalizeWorkflow orchestrates the finalization process
// This workflow:
// 1. Encrypts all environment files using age encryption
// 2. Optionally adds encrypted files to git staging area
//
// Returns a WorkflowResult with details about encrypted files
func FinalizeWorkflow(opts FinalizeOptions) (*WorkflowResult, error) {
	result := &WorkflowResult{}

	// Use discard writer if none provided
	w := opts.OutputWriter
	if w == nil {
		w = io.Discard
	}

	// Validate inputs
	if opts.EncryptionKeyPath == "" {
		opts.EncryptionKeyPath = env.DefaultAgeKeyPath
	}
	if opts.Environments == nil || len(opts.Environments) == 0 {
		opts.Environments = env.AllEnvironmentFiles()
	}

	// Step 1: Encrypt all environment files using library function
	encryptResult, err := env.EncryptEnvironments(env.EncryptionOptions{
		KeyPath:      opts.EncryptionKeyPath,
		Environments: opts.Environments,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to encrypt environments: %w", err)
	}

	// Transfer results from encryption to workflow result
	for _, file := range encryptResult.ProcessedFiles {
		result.AddGenerated(file)
	}
	for _, file := range encryptResult.SkippedFiles {
		result.AddSkipped(file)
	}
	for _, err := range encryptResult.Errors {
		result.AddWarning(err.Error())
	}

	// Step 2: Git add (optional)
	if opts.GitAdd && len(encryptResult.ProcessedFiles) > 0 {
		// Build full paths for git add
		var encryptedPaths []string
		for _, envFile := range opts.Environments {
			if envFile.Exists() {
				encryptedPaths = append(encryptedPaths, envFile.FullEncryptedPath())
			}
		}

		if len(encryptedPaths) > 0 {
			args := append([]string{"add"}, encryptedPaths...)
			cmd := exec.Command("git", args...)
			if err := cmd.Run(); err != nil {
				result.AddWarning(fmt.Sprintf("Failed to git add files: %v. You can manually add: git add %s",
					err, strings.Join(encryptedPaths, " ")))
			}
		}
	}

	return result, nil
}
