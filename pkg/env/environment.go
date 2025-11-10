// Package env provides environment file generation with smart defaults
package env

import (
	"fmt"
	"os"
	"path/filepath"
)

// Age encryption constants
const (
	// DefaultAgeKeyPath is the default location for Age encryption identity key
	DefaultAgeKeyPath = ".age/key.txt"
)

// Environment represents a target environment type (local, production, secrets, etc.)
// with smart defaults that require zero configuration
type Environment struct {
	Name     string // Environment name: "local", "production", "secrets", etc.
	FileName string // Target filename: ".env.local", ".env.production", etc.
	BaseDir  string // Base directory for files (defaults to "." for backward compatibility)
}

// Generate generates an environment file template with smart defaults based on environment type
// The appName is used in headers to identify the application
func (e *Environment) Generate(registry *Registry, appName string) string {
	return registry.GenerateTemplate(e.defaultOptions(appName))
}

// EncryptedFileName returns the encrypted version of the environment filename
// Example: ".env.local" → ".env.local.age"
func (e *Environment) EncryptedFileName() string {
	return e.FileName + ".age"
}

// FullPath returns the complete path to the environment file
// Combines BaseDir with FileName. If BaseDir is empty, defaults to current directory (".")
func (e *Environment) FullPath() string {
	baseDir := e.BaseDir
	if baseDir == "" {
		baseDir = "."
	}
	return filepath.Join(baseDir, e.FileName)
}

// FullEncryptedPath returns the complete path to the encrypted environment file
func (e *Environment) FullEncryptedPath() string {
	baseDir := e.BaseDir
	if baseDir == "" {
		baseDir = "."
	}
	return filepath.Join(baseDir, e.EncryptedFileName())
}

// WithBaseDir returns a new Environment with the specified base directory
// This enables fluent API usage: env.Local.WithBaseDir("./config")
func (e *Environment) WithBaseDir(dir string) *Environment {
	return &Environment{
		Name:     e.Name,
		FileName: e.FileName,
		BaseDir:  dir,
	}
}

// Exists checks if the environment file exists on disk
func (e *Environment) Exists() bool {
	_, err := os.Stat(e.FullPath())
	return err == nil
}

// defaultOptions returns smart default TemplateOptions based on environment type
func (e *Environment) defaultOptions(appName string) TemplateOptions {
	switch e.Name {
	case "local":
		return TemplateOptions{
			Header: e.defaultHeader(appName, "LOCAL DEVELOPMENT",
				"This file is for local development only",
				"Copy real secrets from .env.secrets"),
			IncludeComments:     true,
			IncludeGroupHeaders: true,
		}

	case "production":
		return TemplateOptions{
			Header: e.defaultHeader(appName, "PRODUCTION",
				"This file is for production deployment",
				"Secrets should be set via your deployment platform"),
			IncludeComments:     true,
			IncludeGroupHeaders: true,
		}

	case "secrets", "secrets-local", "secrets-production":
		envType := "SECRETS"
		if e.Name == "secrets-local" {
			envType = "SECRETS - LOCAL"
		} else if e.Name == "secrets-production" {
			envType = "SECRETS - PRODUCTION"
		}

		return TemplateOptions{
			Header: e.defaultHeader(appName, envType,
				"Generated from registry - FILL IN REAL VALUES",
				"This file should NOT be committed (add to .gitignore)",
				"Use age-encrypt to create "+e.EncryptedFileName()+" for git"),
			ValueOverrides: func(v EnvVar) (string, bool) {
				return "", true // Empty values for secrets template
			},
			IncludeComments:     true,
			IncludeGroupHeaders: true,
		}

	case "example":
		return TemplateOptions{
			Header: e.defaultHeader(appName, "EXAMPLE",
				"Example environment file - copy to .env.local and fill in values"),
			ValueOverrides: func(v EnvVar) (string, bool) {
				if v.Secret {
					return "your-secret-here", true
				}
				return v.Default, false
			},
			IncludeComments:     true,
			IncludeGroupHeaders: true,
		}

	default:
		// Generic environment with minimal headers
		return TemplateOptions{
			Header: []string{
				fmt.Sprintf("# %s - %s\n", appName, e.Name),
			},
			IncludeComments:     true,
			IncludeGroupHeaders: true,
		}
	}
}

// defaultHeader generates a standard header block for environment files
func (e *Environment) defaultHeader(appName, envType string, description ...string) []string {
	header := []string{
		"# ================================================================",
		fmt.Sprintf("# %s - %s", appName, envType),
		"# ================================================================",
	}

	for _, desc := range description {
		header = append(header, fmt.Sprintf("# %s", desc))
	}

	header = append(header, "# ================================================================\n")
	return header
}

// Predefined common environments (zero config required)
var (
	// Local is the local development environment (.env.local)
	Local = &Environment{Name: "local", FileName: ".env.local", BaseDir: "."}

	// Production is the production deployment environment (.env.production)
	Production = &Environment{Name: "production", FileName: ".env.production", BaseDir: "."}

	// Secrets is the secrets-only template environment (.env.secrets)
	// DEPRECATED: Use SecretsLocal or SecretsProduction for environment-specific secrets
	Secrets = &Environment{Name: "secrets", FileName: ".env.secrets", BaseDir: "."}

	// SecretsLocal is the local secrets file (.env.secrets.local)
	SecretsLocal = &Environment{Name: "secrets-local", FileName: ".env.secrets.local", BaseDir: "."}

	// SecretsProduction is the production secrets file (.env.secrets.production)
	SecretsProduction = &Environment{Name: "secrets-production", FileName: ".env.secrets.production", BaseDir: "."}

	// Example is the example/template environment (.env.example)
	Example = &Environment{Name: "example", FileName: ".env.example", BaseDir: "."}
)

// NewEnvironment creates a custom environment with the given name and filename
// Use this if the predefined environments don't meet your needs
func NewEnvironment(name, fileName string) *Environment {
	return &Environment{Name: name, FileName: fileName, BaseDir: "."}
}

// NewEnvironmentWithBase creates a custom environment with a specific base directory
func NewEnvironmentWithBase(name, fileName, baseDir string) *Environment {
	return &Environment{Name: name, FileName: fileName, BaseDir: baseDir}
}

// ================================================================
// Centralized File Management
// ================================================================

// AllEnvironmentFiles returns a list of all standard environment files
// Used for operations like clean, list, etc.
func AllEnvironmentFiles() []*Environment {
	return []*Environment{
		Local,
		Production,
		SecretsLocal,
		SecretsProduction,
	}
}

// AllEncryptedFiles returns a list of all encrypted environment files
// Used for encryption/decryption operations
func AllEncryptedFiles() []string {
	var files []string
	for _, env := range AllEnvironmentFiles() {
		files = append(files, env.EncryptedFileName())
	}
	return files
}

// ================================================================
// Secrets Fallback Logic
// ================================================================

// ResolveSecretsFile resolves the appropriate secrets file for an environment
// with fallback logic: .env.secrets.{env} → .env.secrets
//
// For local:      .env.secrets.local → .env.secrets
// For production: .env.secrets.production → .env.secrets
//
// Returns the resolved environment and whether the fallback was used
func ResolveSecretsFile(targetEnv *Environment) (*Environment, bool) {
	var primarySecrets *Environment
	var fallbackSecrets = Secrets

	// Determine the primary secrets file based on target environment
	switch targetEnv.Name {
	case "local":
		primarySecrets = SecretsLocal
	case "production":
		primarySecrets = SecretsProduction
	default:
		// Unknown environment, no fallback available
		return nil, false
	}

	// Check if primary secrets file exists
	if primarySecrets.Exists() {
		return primarySecrets, false
	}

	// Fallback to generic .env.secrets if it exists
	if fallbackSecrets.Exists() {
		return fallbackSecrets, true
	}

	// Neither exists, return primary (caller will handle error)
	return primarySecrets, false
}

// ================================================================
// Utility Functions
// ================================================================

// CleanEnvironmentFiles removes all generated environment files and their encrypted versions.
//
// This function:
//  1. Iterates through all standard environment files
//  2. Removes plaintext versions if they exist
//  3. Removes encrypted (.age) versions if they exist
//  4. Returns count of files removed
//
// Example:
//
//	removed, err := env.CleanEnvironmentFiles()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Removed %d files\n", removed)
func CleanEnvironmentFiles() (int, error) {
	removed := 0

	for _, envFile := range AllEnvironmentFiles() {
		// Remove plaintext version
		if envFile.Exists() {
			if err := os.Remove(envFile.FullPath()); err != nil {
				return removed, fmt.Errorf("failed to remove %s: %w", envFile.FileName, err)
			}
			removed++
		}

		// Remove encrypted version
		encryptedPath := envFile.FullEncryptedPath()
		if _, err := os.Stat(encryptedPath); err == nil {
			if err := os.Remove(encryptedPath); err != nil {
				return removed, fmt.Errorf("failed to remove %s: %w", envFile.EncryptedFileName(), err)
			}
			removed++
		}
	}

	return removed, nil
}

// SetupEnvironment creates an environment file from a registry template.
//
// This is a convenience function that:
//  1. Generates the template from the registry
//  2. Writes it to the environment file
//  3. Sets appropriate file permissions (0600)
//
// Example:
//
//	err := env.SetupEnvironment(registry, env.Local, "My Application")
//	if err != nil {
//	    log.Fatal(err)
//	}
func SetupEnvironment(registry *Registry, environment *Environment, appName string) error {
	content := environment.Generate(registry, appName)
	return os.WriteFile(environment.FullPath(), []byte(content), 0600)
}

// DetectEnvironment determines the current runtime environment.
// Returns one of: "fly.io", "docker", "kubernetes", or "local"
//
// Detection logic:
//   - fly.io: Checks for FLY_APP_NAME environment variable
//   - docker: Checks for /.dockerenv file
//   - kubernetes: Checks for KUBERNETES_SERVICE_HOST environment variable
//   - local: Default when none of the above match
func DetectEnvironment() string {
	if os.Getenv("FLY_APP_NAME") != "" {
		return "fly.io"
	}
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return "docker"
	}
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return "kubernetes"
	}
	return "local"
}
