package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joeblew999/wellknown/pkg/env"
)

// ================================================================
// Helper Functions (adapted from pkg/pb/env.go)
// ================================================================

// GetSecretVars returns only environment variables that should be secrets
func GetSecretVars() []env.EnvVar {
	return AppRegistry.GetSecrets()
}

// GetRequiredVars returns only required environment variables
func GetRequiredVars() []env.EnvVar {
	return AppRegistry.GetRequired()
}

// GetVarsByGroup returns environment variables grouped by their Group field
func GetVarsByGroup() map[string][]env.EnvVar {
	return AppRegistry.GetByGroup()
}

// ValidateEnv checks if all required environment variables are set
func ValidateEnv() error {
	return AppRegistry.ValidateRequired()
}

// ExportSecretsFormat outputs secret environment variables in NAME=VALUE format
func ExportSecretsFormat() string {
	return AppRegistry.ExportSecrets()
}

// GenerateEnvExample generates a .env.example file content from the registry
func GenerateEnvExample() string {
	return AppRegistry.GenerateEnvExample("Sample Application")
}

// ListEnvVars returns a human-readable list of all environment variables
func ListEnvVars() string {
	return AppRegistry.GenerateEnvList("Sample Application Environment Variables")
}

// ================================================================
// Template Generation Helpers
// ================================================================

// GenerateEnvLocal generates .env.local template for development
func GenerateEnvLocal() string {
	return AppRegistry.GenerateTemplate(env.TemplateOptions{
		Header: []string{
			"# ================================================================",
			"# Sample Application - LOCAL DEVELOPMENT",
			"# ================================================================",
			"# This file is for local development only",
			"# Copy real secrets from .env.secrets",
			"# ================================================================\n",
		},
		ValueOverrides: func(v env.EnvVar) (string, bool) {
			// Development-specific overrides
			switch v.Name {
			case "OAUTH_GOOGLE_REDIRECT_URL":
				return "http://localhost:8080/auth/google/callback", true
			case "SERVER_HOST":
				return "127.0.0.1", true
			case "LOG_LEVEL":
				return "debug", true
			default:
				return "", false
			}
		},
		IncludeComments:     true,
		IncludeGroupHeaders: true,
	})
}

// GenerateEnvProduction generates .env.production template for deployment
func GenerateEnvProduction() string {
	return AppRegistry.GenerateTemplate(env.TemplateOptions{
		Header: []string{
			"# ================================================================",
			"# Sample Application - PRODUCTION",
			"# ================================================================",
			"# This file is for production deployment",
			"# Secrets should be set via your deployment platform",
			"# ================================================================\n",
		},
		ValueOverrides: func(v env.EnvVar) (string, bool) {
			// Production-specific overrides
			switch v.Name {
			case "OAUTH_GOOGLE_REDIRECT_URL":
				return "https://example.com/auth/google/callback", true
			case "LOG_LEVEL":
				return "warn", true
			default:
				return "", false
			}
		},
		IncludeComments:     true,
		IncludeGroupHeaders: true,
	})
}

// GenerateEnvSecrets generates .env.secrets template with ONLY secret variables
// This ensures .env.secrets is generated from registry (forward engineering)
// Users then fill in the real secret values
func GenerateEnvSecrets() string {
	secrets := AppRegistry.GetSecrets()

	// Build template with only secret vars
	var filteredVars []env.EnvVar
	for _, v := range secrets {
		filteredVars = append(filteredVars, v)
	}

	// Create a temporary registry with only secrets
	secretsRegistry := env.NewRegistry(filteredVars)

	return secretsRegistry.GenerateTemplate(env.TemplateOptions{
		Header: []string{
			"# ================================================================",
			"# Sample Application - SECRETS",
			"# ================================================================",
			"# Generated from registry - FILL IN REAL VALUES",
			"# This file should NOT be committed (add to .gitignore)",
			"# Use age-encrypt to create .env.secrets.age for git",
			"# ================================================================\n",
		},
		ValueOverrides: func(v env.EnvVar) (string, bool) {
			// Leave all values empty - user must fill
			return "", true
		},
		IncludeComments:     true,
		IncludeGroupHeaders: true,
	})
}

// ================================================================
// Secrets Merging (adapted from pkg/pb/env.go)
// ================================================================

// MergeSecretsIntoEnv merges secrets from .env.secrets or .env.secrets.age into an environment template
// This enables git-tracked secrets (encrypted with Age) to auto-sync into .env.local or .env.production
//
// Workflow:
//  1. Try .env.secrets.age first (git-tracked, encrypted)
//  2. Fall back to .env.secrets (local plaintext)
//  3. Generate template based on templateType (local/production)
//  4. Merge secret values into template (preserving structure, comments, alphabetical order)
//  5. Write to outputPath
//
// Example:
//
//	MergeSecretsIntoEnv(".env.secrets", "local", ".env.local")   // Tries .env.secrets.age, then .env.secrets
//	MergeSecretsIntoEnv(".env.secrets", "production", ".env.production")
func MergeSecretsIntoEnv(secretsPath, templateType, outputPath string) error {
	// 1. Load secrets (with automatic .age detection and decryption)
	secrets, err := env.LoadSecrets(env.SecretsSource{
		FilePath:     secretsPath,
		TryEncrypted: true,
	})
	if err != nil {
		return err
	}

	// 2. Generate template based on type
	var template string
	switch templateType {
	case "local":
		template = GenerateEnvLocal()
	case "production":
		template = GenerateEnvProduction()
	default:
		return fmt.Errorf("invalid template type: %s (must be 'local' or 'production')", templateType)
	}

	// 3. Merge secrets into template
	mergedContent := env.MergeIntoTemplate(template, secrets)

	// 4. Write merged content to output file
	if err := os.WriteFile(outputPath, []byte(mergedContent), 0600); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// ================================================================
// Deployment File Sync Helpers
// ================================================================

// GenerateDockerfileEnvDocs generates Dockerfile environment documentation
func GenerateDockerfileEnvDocs() string {
	return AppRegistry.GenerateDockerfileDocs(env.DockerfileDocsOptions{
		AppName:            "Sample Application",
		UpdateCommand:      "go run . dockerfile-sync",
		DeploymentPlatform: "Docker",
		NonSecretEnvSource: "docker-compose.yml or Dockerfile ENV",
		SecretSource:       "Docker secrets or environment",
		SyncCommand:        "docker-compose up",
	})
}

// GenerateFlyTomlEnv generates fly.toml [env] section
func GenerateFlyTomlEnv() string {
	return AppRegistry.GenerateTOMLEnv("env", []string{
		"# AUTO-GENERATED - DO NOT EDIT MANUALLY",
		"# Update with: go run . fly-sync",
	})
}

// GenerateDockerComposeEnv generates docker-compose.yml environment section
func GenerateDockerComposeEnv() string {
	// Get non-secret vars with defaults for docker-compose
	vars := AppRegistry.All()

	var lines []string
	lines = append(lines, "    environment:")
	lines = append(lines, "      # AUTO-GENERATED - DO NOT EDIT MANUALLY")
	lines = append(lines, "      # Update with: go run . compose-sync")

	// Add non-secret vars that have defaults
	for _, v := range vars {
		if !v.Secret && v.Default != "" {
			lines = append(lines, fmt.Sprintf("      %s: \"%s\"", v.Name, v.Default))
		}
	}

	return strings.Join(lines, "\n")
}

