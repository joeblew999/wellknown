// Package workflow provides high-level orchestration functions for environment
// management workflows. It combines multiple env package primitives into
// complete, production-ready workflows.
//
// # Overview
//
// The workflow package implements the standard 3-phase environment management
// workflow:
//
//  1. Registry Sync: Registry → Templates → Deployment Configs
//  2. Environments Sync: Secrets → Environments → Validation
//  3. Finalize: Encryption → Git Staging
//
// Each workflow is implemented as a single function that handles all steps,
// error handling, and returns structured results.
//
// # The Three Workflows
//
// SyncRegistryWorkflow - Phase 1: After editing registry.go
//
//	result, err := workflow.SyncRegistryWorkflow(workflow.RegistrySyncOptions{
//	    Registry:           AppRegistry,
//	    AppName:            "My Application",
//	    CreateSecretsFiles: true,
//	    DeploymentConfigs:  deploymentConfigs,
//	    // Optional: filter specific configs
//	    SyncOnlyConfigs:    []string{"Dockerfile"},  // Only sync Dockerfile
//	    SkipEnvironments:   false,                   // Skip .env generation
//	})
//
// What it does:
//   - Syncs deployment configuration files (Dockerfile, fly.toml, etc.)
//   - Optionally filters to sync only specific configs (via SyncOnlyConfigs)
//   - Generates environment templates (.env.local, .env.production)
//   - Optionally skips environment generation (via SkipEnvironments)
//   - Creates secrets templates if they don't exist
//
// SyncEnvironmentsWorkflow - Phase 2: After editing secrets files
//
//	result, err := workflow.SyncEnvironmentsWorkflow(workflow.EnvironmentsSyncOptions{
//	    Registry:          AppRegistry,
//	    AppName:           "My Application",
//	    LocalEnv:          env.Local,       // nil = skip local sync
//	    ProductionEnv:     env.Production,  // nil = skip production sync
//	    ValidateRequired:  true,
//	})
//
// What it does:
//   - Loads secrets from .env.secrets.* files (prefers encrypted .age versions)
//   - Merges secrets into .env.local and/or .env.production templates
//   - Supports syncing only local (ProductionEnv: nil) or only production (LocalEnv: nil)
//   - Optionally validates that all required variables are set
//
// FinalizeWorkflow - Phase 3: Encrypt and prepare for commit
//
//	result, err := workflow.FinalizeWorkflow(workflow.FinalizeOptions{
//	    Environments:      env.AllEnvironmentFiles(),
//	    EncryptionKeyPath: ".age/key.txt",
//	    GitAdd:            true,
//	})
//
// What it does:
//   - Encrypts all environment files using Age encryption
//   - Optionally adds encrypted files to git staging area
//   - Returns list of files ready for commit
//
// # Design Philosophy
//
// Library vs CLI Separation:
//
// Library functions (this package):
//   - Pure functions with no side effects beyond file I/O
//   - Return errors instead of calling os.Exit()
//   - Accept io.Writer for testable output
//   - Return structured WorkflowResult objects
//   - Reusable by any Go project
//
// CLI implementations (example/workflow.go):
//   - User-friendly output with emojis and progress
//   - Interactive prompts when needed
//   - Call os.Exit() on fatal errors
//   - Pretty-print results
//
// # Workflow Results
//
// All workflow functions return a WorkflowResult:
//
//	type WorkflowResult struct {
//	    GeneratedFiles []string // Files that were created
//	    UpdatedFiles   []string // Files that were updated
//	    SkippedFiles   []string // Files that were skipped
//	    Warnings       []string // Non-fatal warnings
//	    Errors         []error  // Errors (workflow may continue)
//	}
//
// Check results:
//
//	if result.HasErrors() {
//	    for _, err := range result.Errors {
//	        log.Printf("Error: %v", err)
//	    }
//	}
//
//	if result.HasWarnings() {
//	    for _, warn := range result.Warnings {
//	        log.Printf("Warning: %s", warn)
//	    }
//	}
//
// # Options Patterns
//
// All workflows use Options structs for clean, extensible APIs:
//
//	type RegistrySyncOptions struct {
//	    Registry           *env.Registry
//	    AppName            string
//	    DeploymentConfigs  []DeploymentConfig
//	    CreateSecretsFiles bool
//	    OutputWriter       io.Writer
//	}
//
// This allows:
//   - Optional parameters with sensible defaults
//   - Easy addition of new options without breaking changes
//   - Clear documentation of what each workflow needs
//
// # Custom Deployment Configs
//
// Define custom deployment configurations:
//
//	deploymentConfigs := []workflow.DeploymentConfig{
//	    {
//	        FilePath:    "Dockerfile",
//	        StartMarker: "# === AUTO-GENERATED ===",
//	        EndMarker:   "# === END ===",
//	        Generator: func(r *env.Registry) (string, error) {
//	            return r.GenerateDockerfileDocs(env.DockerfileDocsOptions{}), nil
//	        },
//	    },
//	    {
//	        FilePath:    "docker-compose.yml",
//	        StartMarker: "# === AUTO-GENERATED ===",
//	        EndMarker:   "# === END ===",
//	        Generator: func(r *env.Registry) (string, error) {
//	            return r.GenerateDockerComposeEnv([]string{}), nil
//	        },
//	    },
//	}
//
// # Error Handling
//
// Workflows handle errors gracefully:
//
//   - Fatal errors: Return error immediately
//   - Recoverable errors: Add to result.Warnings and continue
//   - Partial success: Complete as much as possible
//
// Example:
//
//	result, err := workflow.SyncRegistryWorkflow(opts)
//	if err != nil {
//	    // Fatal error - workflow could not complete
//	    log.Fatal(err)
//	}
//
//	// Check for warnings (partial failures)
//	for _, warn := range result.Warnings {
//	    log.Printf("Warning: %s", warn)
//	}
//
// # Testing Workflows
//
// Workflows are designed to be testable:
//
//	func TestMyWorkflow(t *testing.T) {
//	    // Create temp directory
//	    tmpDir, _ := os.MkdirTemp("", "test-*")
//	    defer os.RemoveAll(tmpDir)
//	    os.Chdir(tmpDir)
//
//	    // Create test registry
//	    registry := env.NewRegistry([]env.EnvVar{
//	        {Name: "TEST_VAR", Description: "Test"},
//	    })
//
//	    // Run workflow with custom output
//	    var buf bytes.Buffer
//	    result, err := workflow.SyncRegistryWorkflow(workflow.RegistrySyncOptions{
//	        Registry:     registry,
//	        OutputWriter: &buf,
//	    })
//
//	    // Verify results
//	    if err != nil {
//	        t.Fatal(err)
//	    }
//	    if len(result.UpdatedFiles) == 0 {
//	        t.Error("Expected some updated files")
//	    }
//	}
//
// # Integration Patterns
//
// Build tools:
//
//	//go:generate go run tools/sync-env.go
//
// CI/CD pipelines:
//
//	if result, err := workflow.SyncRegistryWorkflow(opts); err != nil {
//	    os.Exit(1)
//	}
//
// Custom CLIs:
//
//	cmd := &cobra.Command{
//	    Use: "sync",
//	    Run: func(cmd *cobra.Command, args []string) {
//	        result, err := workflow.SyncRegistryWorkflow(opts)
//	        // handle result...
//	    },
//	}
//
// # Complete Example
//
//	package main
//
//	import (
//	    "log"
//	    "github.com/joeblew999/wellknown/pkg/env"
//	    "github.com/joeblew999/wellknown/pkg/env/workflow"
//	)
//
//	func main() {
//	    registry := env.NewRegistry([]env.EnvVar{
//	        {Name: "DATABASE_URL", Secret: true, Required: true},
//	        {Name: "PORT", Default: "8080"},
//	    })
//
//	    // Phase 1: Sync registry
//	    result, err := workflow.SyncRegistryWorkflow(workflow.RegistrySyncOptions{
//	        Registry:           registry,
//	        AppName:            "My App",
//	        CreateSecretsFiles: true,
//	    })
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // User edits secrets files...
//
//	    // Phase 2: Sync environments
//	    result, err = workflow.SyncEnvironmentsWorkflow(workflow.EnvironmentsSyncOptions{
//	        Registry:         registry,
//	        ValidateRequired: true,
//	    })
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // Phase 3: Finalize (encrypt + git)
//	    result, err = workflow.FinalizeWorkflow(workflow.FinalizeOptions{
//	        Environments:      env.AllEnvironmentFiles(),
//	        EncryptionKeyPath: ".age/key.txt",
//	        GitAdd:            true,
//	    })
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    log.Printf("Ready to commit: %v", result.GeneratedFiles)
//	}
//
// # Further Reading
//
// See also:
//   - pkg/env: Core environment management primitives
//   - pkg/env/LIBRARY_USAGE.md: Complete library documentation
//   - example/workflow.go: Full CLI implementation example
//   - *_test.go: Comprehensive test examples
package workflow
