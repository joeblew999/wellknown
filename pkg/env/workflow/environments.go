package workflow

import (
	"fmt"
	"io"
	"os"

	"github.com/joeblew999/wellknown/pkg/env"
)

// SyncEnvironmentsWorkflow orchestrates the environment synchronization process
// This workflow:
// 1. Loads secrets from .env.secrets.local and .env.secrets.production
// 2. Merges secrets into .env.local and .env.production templates
// 3. Optionally validates that all required variables are set
//
// Returns a WorkflowResult with details about files updated and validation status
func SyncEnvironmentsWorkflow(opts EnvironmentsSyncOptions) (*WorkflowResult, error) {
	result := &WorkflowResult{}

	// Use discard writer if none provided
	w := opts.OutputWriter
	if w == nil {
		w = io.Discard
	}

	// Validate inputs
	if opts.Registry == nil {
		return nil, fmt.Errorf("registry cannot be nil")
	}
	if opts.AppName == "" {
		opts.AppName = "Application"
	}

	// Step 1: Sync local environment (if provided)
	if opts.LocalEnv != nil {
		secretsEnv, usedFallback := env.ResolveSecretsFile(opts.LocalEnv)
		if secretsEnv == nil {
			return nil, fmt.Errorf("no secrets file found for %s", opts.LocalEnv.Name)
		}

		if usedFallback {
			result.AddWarning(fmt.Sprintf("Using fallback secrets file: %s", secretsEnv.FileName))
		}

		secrets, err := env.LoadSecrets(env.SecretsSource{
			FilePath:        secretsEnv.FileName,
			PreferEncrypted: true,
		})
		if err != nil {
			return result, fmt.Errorf("failed to load secrets from %s: %w", secretsEnv.FileName, err)
		}

		template := opts.LocalEnv.Generate(opts.Registry, opts.AppName)
		mergedContent := env.MergeIntoTemplate(template, secrets)

		if err := os.WriteFile(opts.LocalEnv.FullPath(), []byte(mergedContent), 0600); err != nil {
			return result, fmt.Errorf("failed to write %s: %w", opts.LocalEnv.FileName, err)
		}
		result.AddUpdated(opts.LocalEnv.FileName)
	}

	// Step 2: Sync production environment (if provided)
	if opts.ProductionEnv != nil {
		secretsEnvProd, usedFallbackProd := env.ResolveSecretsFile(opts.ProductionEnv)
		if secretsEnvProd == nil {
			return result, fmt.Errorf("no secrets file found for %s", opts.ProductionEnv.Name)
		}

		if usedFallbackProd {
			result.AddWarning(fmt.Sprintf("Using fallback secrets file: %s", secretsEnvProd.FileName))
		}

		secretsProd, err := env.LoadSecrets(env.SecretsSource{
			FilePath:        secretsEnvProd.FileName,
			PreferEncrypted: true,
		})
		if err != nil {
			return result, fmt.Errorf("failed to load secrets from %s: %w", secretsEnvProd.FileName, err)
		}

		templateProd := opts.ProductionEnv.Generate(opts.Registry, opts.AppName)
		mergedContentProd := env.MergeIntoTemplate(templateProd, secretsProd)

		if err := os.WriteFile(opts.ProductionEnv.FullPath(), []byte(mergedContentProd), 0600); err != nil {
			return result, fmt.Errorf("failed to write %s: %w", opts.ProductionEnv.FileName, err)
		}
		result.AddUpdated(opts.ProductionEnv.FileName)
	}

	// Step 3: Validate required variables (optional)
	if opts.ValidateRequired {
		if err := opts.Registry.ValidateRequired(); err != nil {
			result.AddWarning(fmt.Sprintf("Validation failed: %v", err))
		}
	}

	return result, nil
}
