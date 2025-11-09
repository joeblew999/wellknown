package wellknown

import (
	"fmt"
	"os"
	"strings"

	"github.com/joeblew999/wellknown/pkg/env"
)

// GetSecretVars returns only environment variables that should be in Fly.io secrets
func GetSecretVars() []env.EnvVar {
	return EnvRegistry.GetSecrets()
}

// GetRequiredVars returns only required environment variables
func GetRequiredVars() []env.EnvVar {
	return EnvRegistry.GetRequired()
}

// GetVarsByGroup returns environment variables grouped by their Group field
func GetVarsByGroup() map[string][]env.EnvVar {
	return EnvRegistry.GetByGroup()
}

// ValidateEnv checks if all required environment variables are set
func ValidateEnv() error {
	return EnvRegistry.ValidateRequired()
}

// ExportSecretsFormat outputs secret environment variables in NAME=VALUE format for flyctl secrets import
func ExportSecretsFormat() string {
	return EnvRegistry.ExportSecrets()
}

// GenerateEnvExample generates a .env.example file content from the registry
func GenerateEnvExample() string {
	return EnvRegistry.GenerateEnvExample("Wellknown")
}

// ListEnvVars returns a human-readable list of all environment variables
func ListEnvVars() string {
	return EnvRegistry.GenerateEnvList("Environment Variables Registry")
}

// ================================================================
// File Generation and Synchronization
// ================================================================

// GenerateDockerfileEnvDocs generates Dockerfile-style environment variable documentation
func GenerateDockerfileEnvDocs() string {
	return EnvRegistry.GenerateDockerfileDocs(env.DockerfileDocsOptions{
		AppName:            "Wellknown",
		UpdateCommand:      "make env-sync-dockerfile",
		DeploymentPlatform: "Fly.io",
		NonSecretEnvSource: "fly.toml [env] section",
		SecretSource:       "Fly.io secrets",
		SyncCommand:        "make fly-secrets",
	})
}

// SyncDockerfileEnvDocs updates the Dockerfile environment variable documentation section
func SyncDockerfileEnvDocs(dockerfilePath string, dryRun bool) error {
	return env.SyncFileSection(env.SyncOptions{
		FilePath:     dockerfilePath,
		StartMarker:  "# ================================================================\n# Environment Variables (injected at runtime by Fly.io)",
		EndMarker:    "# Sync secrets: make fly-secrets\n# ================================================================",
		Content:      GenerateDockerfileEnvDocs(),
		DryRun:       dryRun,
		CreateBackup: true,
	})
}

// GenerateFlyTomlEnv generates the [env] section for fly.toml
func GenerateFlyTomlEnv() string {
	return EnvRegistry.GenerateTOMLEnv("env", []string{
		"PocketBase configuration (non-secret)",
		"Secrets (OAuth, SMTP, etc.) are set via: make fly-secrets",
	})
}

// SyncFlyTomlEnv updates the fly.toml [env] section
func SyncFlyTomlEnv(flytomlPath string, dryRun bool) error {
	// Read file to find section boundaries
	data, err := os.ReadFile(flytomlPath)
	if err != nil {
		return fmt.Errorf("failed to read fly.toml: %w", err)
	}
	content := string(data)

	// Find the [env] section
	startMarker := "[env]"
	startIdx := strings.Index(content, startMarker)
	if startIdx == -1 {
		return fmt.Errorf("could not find [env] section in fly.toml")
	}

	// Find the end of the [env] section (next section or end of file)
	endIdx := startIdx + len(startMarker)
	nextSectionIdx := strings.Index(content[endIdx:], "\n[")
	var endMarker string
	if nextSectionIdx != -1 {
		// Use the next section as end marker
		endMarker = content[startIdx : endIdx+nextSectionIdx]
	} else {
		// Use entire remaining content as the section
		endMarker = content[startIdx:]
	}

	// Use generic sync with the dynamically determined markers
	return env.SyncFileSection(env.SyncOptions{
		FilePath:     flytomlPath,
		StartMarker:  startMarker,
		EndMarker:    endMarker,
		Content:      startMarker + "\n" + GenerateFlyTomlEnv() + "\n",
		DryRun:       dryRun,
		CreateBackup: true,
	})
}

// GenerateEnvLocal generates .env.local template for development
func GenerateEnvLocal() string {
	return EnvRegistry.GenerateTemplate(env.TemplateOptions{
		Header: []string{
			"# ================================================================",
			"# Wellknown Environment Variables - LOCAL DEVELOPMENT",
			"# ================================================================",
			"# AUTO-GENERATED from pkg/pb/env.go",
			"# To update: make env-generate-local",
			"# ================================================================\n",
		},
		GroupOrder: []string{
			"Google OAuth",
			"HTTPS (Development)",
			"Server",
			"AI",
			"Apple OAuth",
			"PocketBase Admin",
			"SMTP",
			"S3",
			"Deployment",
			"Binary Update",
		},
		ValueOverrides: func(v env.EnvVar) (string, bool) {
			// Development-specific overrides for localhost
			switch v.Name {
			case "GOOGLE_REDIRECT_URL":
				return "https://localhost:8443/auth/google/callback", true
			case "APPLE_REDIRECT_URL":
				return "https://localhost:8443/auth/apple/callback", true
			case "HTTPS_ENABLED":
				return "true", true
			case "CERT_FILE":
				return ".data/certs/cert.pem", true
			case "KEY_FILE":
				return ".data/certs/key.pem", true
			case "HTTPS_PORT":
				return "8443", true
			default:
				return "", false // Use default value
			}
		},
		IncludeComments:     true,
		IncludeGroupHeaders: true,
	})
}

// GenerateEnvProduction generates .env.production template for Fly.io
func GenerateEnvProduction() string {
	return EnvRegistry.GenerateTemplate(env.TemplateOptions{
		Header: []string{
			"# ================================================================",
			"# Wellknown Environment Variables - PRODUCTION (Fly.io)",
			"# ================================================================",
			"# AUTO-GENERATED from pkg/pb/env.go",
			"# To update: make env-generate-production",
			"# This file is for Fly.io deployment only",
			"# For local development, use .env.local",
			"# ================================================================\n",
		},
		GroupOrder: []string{
			"Google OAuth",
			"HTTPS (Development)", // Special note for production
			"AI",
			"Apple OAuth",
			"PocketBase Admin",
			"SMTP",
			"S3",
		},
		ValueOverrides: func(v env.EnvVar) (string, bool) {
			// Production-specific overrides for Fly.io
			switch v.Name {
			case "GOOGLE_REDIRECT_URL":
				return "https://wellknown-pb.fly.dev/auth/google/callback", true
			case "APPLE_REDIRECT_URL":
				return "https://wellknown-pb.fly.dev/auth/apple/callback", true
			case "HTTPS_ENABLED":
				return "false", true
			default:
				return "", false // Use default value
			}
		},
		IncludeComments:     true,
		IncludeGroupHeaders: true,
		GroupHeaderFormat: func(groupName string) string {
			// Special note for HTTPS group in production
			if groupName == "HTTPS (Development)" {
				return fmt.Sprintf("# ----------------------------------------------------------------\n# %s\n# Production: DISABLED - Fly.io handles HTTPS\n# ----------------------------------------------------------------\n", groupName)
			}
			return fmt.Sprintf("# ----------------------------------------------------------------\n# %s\n# ----------------------------------------------------------------\n", groupName)
		},
	})
}

// ================================================================
// Secrets Merging (Git-Tracked Secrets with Auto-Sync)
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
		FilePath:        secretsPath,
		PreferEncrypted: true,
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
