package workflow

import (
	"fmt"
	"io"
	"os"

	"github.com/joeblew999/wellknown/pkg/env"
)

// SyncRegistryWorkflow orchestrates the registry synchronization process
// This workflow:
// 1. Syncs deployment configuration files (Dockerfile, fly.toml, etc.)
// 2. Generates environment templates (.env.local, .env.production)
// 3. Creates secrets templates if they don't exist
//
// Returns a WorkflowResult with details about files created/updated/skipped
func SyncRegistryWorkflow(opts RegistrySyncOptions) (*WorkflowResult, error) {
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

	// Step 1: Sync deployment configs
	for _, cfg := range opts.DeploymentConfigs {
		// Filter: skip if not in SyncOnlyConfigs list
		if len(opts.SyncOnlyConfigs) > 0 {
			found := false
			for _, allowed := range opts.SyncOnlyConfigs {
				if cfg.FilePath == allowed {
					found = true
					break
				}
			}
			if !found {
				continue // Skip this config
			}
		}

		content, err := cfg.Generator(opts.Registry)
		if err != nil {
			result.AddWarning(fmt.Sprintf("Failed to generate %s: %v", cfg.FilePath, err))
			continue
		}

		err = env.SyncFileSection(env.SyncOptions{
			FilePath:    cfg.FilePath,
			StartMarker: cfg.StartMarker,
			EndMarker:   cfg.EndMarker,
			Content:     content,
		})

		if err != nil {
			result.AddWarning(fmt.Sprintf("Failed to sync %s: %v", cfg.FilePath, err))
		} else {
			result.AddUpdated(cfg.FilePath)
		}
	}

	// Step 2: Update environment templates (unless skipped)
	if !opts.SkipEnvironments {
		// Update local environment template
		localContent := env.Local.Generate(opts.Registry, opts.AppName)
		if err := os.WriteFile(env.Local.FullPath(), []byte(localContent), 0600); err != nil {
			return result, fmt.Errorf("failed to write %s: %w", env.Local.FileName, err)
		}
		result.AddUpdated(env.Local.FileName)

		// Update production environment template
		prodContent := env.Production.Generate(opts.Registry, opts.AppName)
		if err := os.WriteFile(env.Production.FullPath(), []byte(prodContent), 0600); err != nil {
			return result, fmt.Errorf("failed to write %s: %w", env.Production.FileName, err)
		}
		result.AddUpdated(env.Production.FileName)
	}

	// Step 4: Generate secrets templates if they don't exist (and requested)
	if opts.CreateSecretsFiles {
		secretsRegistry := env.NewRegistry(opts.Registry.GetSecrets())

		// Local secrets
		if !env.SecretsLocal.Exists() {
			secretsContent := env.SecretsLocal.Generate(secretsRegistry, opts.AppName)
			if err := os.WriteFile(env.SecretsLocal.FullPath(), []byte(secretsContent), 0600); err != nil {
				return result, fmt.Errorf("failed to write %s: %w", env.SecretsLocal.FileName, err)
			}
			result.AddGenerated(env.SecretsLocal.FileName)
		} else {
			result.AddSkipped(env.SecretsLocal.FileName)
		}

		// Production secrets
		if !env.SecretsProduction.Exists() {
			secretsContentProd := env.SecretsProduction.Generate(secretsRegistry, opts.AppName)
			if err := os.WriteFile(env.SecretsProduction.FullPath(), []byte(secretsContentProd), 0600); err != nil {
				return result, fmt.Errorf("failed to write %s: %w", env.SecretsProduction.FileName, err)
			}
			result.AddGenerated(env.SecretsProduction.FileName)
		} else {
			result.AddSkipped(env.SecretsProduction.FileName)
		}
	}

	return result, nil
}
