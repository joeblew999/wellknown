// Package env provides registry-driven environment variable management with
// support for templates, secrets separation, and Age encryption.
//
// # Core Concepts
//
// The package follows a registry-driven architecture where all environment
// variables are defined once in a central Registry, which then generates
// templates, validates requirements, and manages secrets.
//
// Key components:
//   - Registry: Central definition of all environment variables
//   - EnvVar: Individual variable with metadata (name, description, default, secret, required)
//   - Environment: Represents a .env file (local, production, secrets)
//   - Secrets: Separate management for sensitive values with Age encryption support
//
// # Quick Start
//
// Define your environment variables in a registry:
//
//	registry := env.NewRegistry([]env.EnvVar{
//	    {
//	        Name:        "DATABASE_URL",
//	        Description: "PostgreSQL connection string",
//	        Secret:      true,
//	        Required:    true,
//	        Group:       "Database",
//	    },
//	    {
//	        Name:        "PORT",
//	        Description: "Server port",
//	        Default:     "8080",
//	        Group:       "Server",
//	    },
//	})
//
// Access values with type safety:
//
//	dbURL := registry.ByName("DATABASE_URL").GetString()
//	port := registry.ByName("PORT").GetInt()
//	debug := registry.ByName("DEBUG").GetBool()
//
// # Environment Files
//
// Pre-defined environment file types:
//
//	env.Local              // .env.local (development)
//	env.Production         // .env.production (production)
//	env.SecretsLocal       // .env.secrets.local (local secrets)
//	env.SecretsProduction  // .env.secrets.production (production secrets)
//
// Generate environment templates:
//
//	content := env.Local.Generate(registry, "My Application")
//	os.WriteFile(env.Local.FileName, []byte(content), 0600)
//
// # Secrets Management
//
// Load secrets from files (prefers encrypted .age versions):
//
//	secrets, err := env.LoadSecrets(env.SecretsSource{
//	    FilePath:        env.SecretsLocal.FileName,
//	    PreferEncrypted: true,
//	})
//
// Merge secrets into templates:
//
//	template := env.Local.Generate(registry, "My App")
//	merged := env.MergeIntoTemplate(template, secrets)
//	os.WriteFile(env.Local.FileName, []byte(merged), 0600)
//
// # Validation
//
// Validate that all required variables are set:
//
//	if err := registry.ValidateRequired(); err != nil {
//	    log.Fatalf("Missing required variables: %v", err)
//	}
//
// # Deployment Configuration
//
// Generate deployment-specific formats:
//
//	// Dockerfile documentation
//	docs := registry.GenerateDockerfileDocs(env.DockerfileDocsOptions{})
//
//	// docker-compose.yml environment section
//	yaml := registry.GenerateDockerComposeEnv([]string{})
//
//	// fly.toml environment and secrets
//	toml := registry.GenerateTOMLEnv("env", []string{})
//	secrets := registry.GenerateTOMLSecretsList("secrets")
//
// Sync auto-generated sections in files:
//
//	err := env.SyncFileSection(env.SyncOptions{
//	    FilePath:    "Dockerfile",
//	    StartMarker: "# === AUTO-GENERATED ===",
//	    EndMarker:   "# === END ===",
//	    Content:     docs,
//	})
//
// # Workflow Functions
//
// For high-level orchestration, see the workflow subpackage:
//
//	import "github.com/joeblew999/wellknown/pkg/env/workflow"
//
//	result, err := workflow.SyncRegistryWorkflow(workflow.RegistrySyncOptions{
//	    Registry: registry,
//	    AppName:  "My Application",
//	})
//
// # File Organization
//
// Typical project structure:
//
//	.env.local                    # Local environment (generated from registry)
//	.env.production               # Production environment (generated from registry)
//	.env.secrets.local            # Local secrets (user-edited)
//	.env.secrets.production       # Production secrets (user-edited)
//	.env.secrets.local.age        # Encrypted local secrets (git-safe)
//	.env.secrets.production.age   # Encrypted production secrets (git-safe)
//	.age/key.txt                  # Age encryption key (DO NOT COMMIT)
//
// # Security Best Practices
//
//   - Mark sensitive variables with Secret: true
//   - Store secrets in separate .env.secrets.* files
//   - Encrypt secrets with Age before committing
//   - Never commit plaintext .env files
//   - Never commit .age/key.txt
//   - Use .gitignore to prevent accidents
//
// # Advanced Usage
//
// Custom environments:
//
//	staging := &env.Environment{
//	    Name:     "staging",
//	    FileName: ".env.staging",
//	    BaseDir:  "./config",
//	}
//
// Filter variables:
//
//	secrets := registry.GetSecrets()
//	required := registry.GetRequired()
//	byGroup := registry.GetByGroup()
//
// For complete examples and library usage patterns, see:
//   - LIBRARY_USAGE.md: Comprehensive library documentation
//   - example/: Working CLI implementation
//   - workflow/: High-level orchestration functions
//
// # Package Organization
//
// Main files:
//   - registry.go: Registry and EnvVar types with accessors
//   - environment.go: Environment file abstraction
//   - template.go: Template generation functions
//   - secrets.go: Secrets loading and encryption
//   - export.go: Deployment format generators (Dockerfile, TOML, YAML)
//   - sync.go: File section synchronization
//
// Subpackages:
//   - workflow/: High-level workflow orchestration functions
//
// Testing:
//   - *_test.go: Unit tests for all functions
//   - example_*_test.go: Testable examples for documentation
package env
