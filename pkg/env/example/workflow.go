package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/joeblew999/wellknown/pkg/env"
	"github.com/joeblew999/wellknown/pkg/env/deploy"
	"github.com/joeblew999/wellknown/pkg/env/workflow"
)

// ================================================================
// Workflow Automation Commands
// ================================================================
// These commands combine multiple steps to simplify common workflows.
// They clearly separate USER ACTIONS (editing) from SYSTEM ACTIONS (automation).

// cmdSyncRegistry syncs all configs after editing registry.go
// Phase 1: USER edits registry.go → run this → edits secrets
func cmdSyncRegistry() {
	fmt.Println("🔄 Syncing from registry...")
	fmt.Println()

	// Build deployment configs for this example
	deploymentConfigs := []workflow.DeploymentConfig{
		{
			FilePath:    "Dockerfile",
			StartMarker: "# === AUTO-GENERATED ENVIRONMENT (do not edit between markers) ===",
			EndMarker:   "# === END AUTO-GENERATED ===",
			Generator:   func(r *env.Registry) (string, error) { return r.GenerateDockerfileDocs(env.DockerfileDocsOptions{}), nil },
		},
		{
			FilePath:    "fly.toml",
			StartMarker: "# === AUTO-GENERATED ENVIRONMENT (do not edit between markers) ===",
			EndMarker:   "# === END AUTO-GENERATED ===",
			Generator: func(r *env.Registry) (string, error) {
				tomlEnv := r.GenerateTOMLEnv("env", []string{})
				secretsList := r.GenerateTOMLSecretsList("secrets")
				return tomlEnv + "\n" + secretsList, nil
			},
		},
		{
			FilePath:    "docker-compose.yml",
			StartMarker: "# === AUTO-GENERATED ENVIRONMENT (do not edit between markers) ===",
			EndMarker:   "# === END AUTO-GENERATED ===",
			Generator: func(r *env.Registry) (string, error) {
				content := r.GenerateDockerComposeEnv([]string{
					"AUTO-GENERATED - DO NOT EDIT MANUALLY",
					"Update with: go run . sync-registry",
				})
				// Add markers back since SyncFileSection replaces them
				return "# === AUTO-GENERATED ENVIRONMENT (do not edit between markers) ===\n" + content + "    # === END AUTO-GENERATED ===", nil
			},
		},
	}

	// Call workflow function
	result, err := workflow.SyncRegistryWorkflow(workflow.RegistrySyncOptions{
		Registry:           AppRegistry,
		AppName:            "Sample Application",
		DeploymentConfigs:  deploymentConfigs,
		CreateSecretsFiles: true,
		OutputWriter:       nil, // Use default (discard)
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to sync registry: %v\n", err)
		os.Exit(1)
	}

	// CLI-specific output formatting
	fmt.Println("📝 Step 1/5: Syncing deployment configs")
	for _, file := range result.UpdatedFiles {
		if strings.Contains(file, "Dockerfile") || strings.Contains(file, "fly.toml") || strings.Contains(file, "docker-compose.yml") {
			fmt.Printf("   ✅ Synced %s\n", file)
		}
	}
	for _, warn := range result.Warnings {
		if strings.Contains(warn, "Dockerfile") || strings.Contains(warn, "fly.toml") || strings.Contains(warn, "docker-compose.yml") {
			fmt.Printf("   ⚠️  %s\n", warn)
		}
	}
	fmt.Println()

	fmt.Println("📝 Step 2/5: Updating .env.local template")
	fmt.Printf("   ✅ Updated %s\n", env.Local.FileName)
	fmt.Println()

	fmt.Println("📝 Step 3/5: Updating .env.production template")
	fmt.Printf("   ✅ Updated %s\n", env.Production.FileName)
	fmt.Println()

	fmt.Println("📝 Step 4/5: Checking secrets templates")
	for _, file := range result.GeneratedFiles {
		fmt.Printf("   ✅ Created %s\n", file)
	}
	for _, file := range result.SkippedFiles {
		fmt.Printf("   ℹ️  %s exists (not overwriting)\n", file)
	}
	fmt.Println()

	// Step 5: Summary
	fmt.Println("📝 Step 5/5: Summary")
	secrets := AppRegistry.GetSecrets()
	fmt.Printf("   Registry: %d total variables (%d secrets)\n", len(AppRegistry.All()), len(secrets))
	fmt.Println()

	// Tell user what to do next
	fmt.Println("✅ Registry synced successfully!")
	fmt.Println()
	fmt.Println("📝 NEXT: Edit secrets files with real values:")
	fmt.Printf("   - %s (for local development)\n", env.SecretsLocal.FileName)
	fmt.Printf("   - %s (for production)\n", env.SecretsProduction.FileName)
	fmt.Println()
	fmt.Println("Then run: go run . sync-environments")
}

// cmdSyncEnvironments merges secrets into environments and validates
// Phase 2: USER edits secrets → run this → automation
func cmdSyncEnvironments() {
	fmt.Println("🔄 Syncing environments from secrets...")
	fmt.Println()

	// Call workflow function
	result, err := workflow.SyncEnvironmentsWorkflow(workflow.EnvironmentsSyncOptions{
		Registry:          AppRegistry,
		AppName:           "Sample Application",
		LocalEnv:          env.Local,
		ProductionEnv:     env.Production,
		LocalSecrets:      env.SecretsLocal,
		ProductionSecrets: env.SecretsProduction,
		ValidateRequired:  true,
		OutputWriter:      nil, // Use default (discard)
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to sync environments: %v\n", err)
		os.Exit(1)
	}

	// CLI-specific output formatting
	fmt.Println("📝 Step 1/3: Syncing local environment")
	for _, warn := range result.Warnings {
		if strings.Contains(warn, "fallback") && strings.Contains(warn, "local") {
			fmt.Printf("   ⚠️  %s\n", warn)
		}
	}
	// Calculate secrets count from file (approximate)
	fmt.Printf("   ✅ Merged secrets → %s\n", env.Local.FileName)
	fmt.Println()

	fmt.Println("📝 Step 2/3: Syncing production environment")
	for _, warn := range result.Warnings {
		if strings.Contains(warn, "fallback") && strings.Contains(warn, "production") {
			fmt.Printf("   ⚠️  %s\n", warn)
		}
	}
	fmt.Printf("   ✅ Merged secrets → %s\n", env.Production.FileName)
	fmt.Println()

	// Step 3: Validate
	fmt.Println("📝 Step 3/3: Validating environments")
	validationWarnings := false
	for _, warn := range result.Warnings {
		if strings.Contains(warn, "Validation failed") {
			fmt.Printf("   ⚠️  %s\n", warn)
			validationWarnings = true
		}
	}
	if validationWarnings {
		fmt.Println()
		fmt.Println("   Fill these in before deploying!")
	} else {
		fmt.Println("   ✅ All required variables set")
	}
	fmt.Println()

	fmt.Println("✅ Environments synced successfully!")
	fmt.Println()
	fmt.Println("📝 NEXT: Encrypt and commit:")
	fmt.Println("   go run . finalize")
}

// cmdFinalize encrypts all files and prepares for git commit
// Phase 3: AUTOMATION - encrypt + git add + show commit message
func cmdFinalize() {
	fmt.Println("🔒 Finalizing for git commit...")
	fmt.Println()

	// Step 1: Check for age key (CLI-specific interactive prompt)
	fmt.Println("📝 Step 1/3: Checking encryption key")
	keyPath := env.DefaultAgeKeyPath
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		fmt.Printf("   ⚠️  No age key found at %s\n", keyPath)
		fmt.Print("   Generate key now? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "y" || response == "yes" {
			cmdAgeKeygen()
			fmt.Println()
		} else {
			fmt.Println("   Skipped key generation")
			fmt.Println("   Run 'go run . age-keygen' to create a key")
			os.Exit(1)
		}
	} else {
		fmt.Printf("   ✅ Found key at %s\n", keyPath)
	}
	fmt.Println()

	// Call workflow function
	result, err := workflow.FinalizeWorkflow(workflow.FinalizeOptions{
		Environments:      env.AllEnvironmentFiles(),
		EncryptionKeyPath: keyPath,
		GitAdd:            true,
		OutputWriter:      nil, // Use default (discard)
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to finalize: %v\n", err)
		os.Exit(1)
	}

	// CLI-specific output formatting
	fmt.Println("📝 Step 2/3: Encrypting environment files")
	for _, file := range result.GeneratedFiles {
		// Extract base name from .age extension
		baseName := strings.TrimSuffix(file, ".age")
		fmt.Printf("   ✅ %s → %s\n", baseName, file)
	}
	for _, warn := range result.Warnings {
		fmt.Printf("   ⚠️  %s\n", warn)
	}
	fmt.Println()

	// Step 3: Git add
	fmt.Println("📝 Step 3/3: Preparing git commit")
	if len(result.GeneratedFiles) > 0 {
		fmt.Printf("   ✅ Added %d files to git\n", len(result.GeneratedFiles))
	}
	fmt.Println()

	// Summary
	fmt.Println("✅ Finalized successfully!")
	fmt.Println()
	fmt.Println("📝 NEXT: Commit and push:")
	fmt.Println("   git commit -m \"chore: update encrypted environments\"")
	fmt.Println("   git push")
	fmt.Println()
	fmt.Println("⚠️  REMINDER:")
	fmt.Println("   - .age files are SAFE to commit")
	fmt.Println("   - NEVER commit plaintext .env files")
	fmt.Printf("   - NEVER commit %s\n", env.DefaultAgeKeyPath)
}

// ================================================================
// Helper Functions
// ================================================================

// cmdFlySyncInternal syncs fly.toml without exiting on error
func cmdFlySyncInternal() {
	tomlContent := AppRegistry.GenerateTOMLEnv("env", []string{})
	secretsList := AppRegistry.GenerateTOMLSecretsList("secrets")

	if err := env.SyncFileSection(env.SyncOptions{
		FilePath:    "fly.toml",
		StartMarker: "# === AUTO-GENERATED ENVIRONMENT (do not edit between markers) ===",
		EndMarker:   "# === END AUTO-GENERATED ===",
		Content:     tomlContent + "\n" + secretsList,
		DryRun:      false,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "   ⚠️  Failed to sync fly.toml: %v\n", err)
		return
	}
	fmt.Println("   ✅ Synced fly.toml")
}

// cmdComposeSyncInternal syncs docker-compose.yml without exiting on error
func cmdComposeSyncInternal() {
	yamlContent := AppRegistry.GenerateDockerComposeEnv([]string{})

	if err := env.SyncFileSection(env.SyncOptions{
		FilePath:    "docker-compose.yml",
		StartMarker: "# === AUTO-GENERATED ENVIRONMENT (do not edit between markers) ===",
		EndMarker:   "# === END AUTO-GENERATED ===",
		Content:     yamlContent,
		DryRun:      false,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "   ⚠️  Failed to sync docker-compose.yml: %v\n", err)
		return
	}
	fmt.Println("   ✅ Synced docker-compose.yml")
}

// ================================================================
// Ko Build Workflow
// ================================================================

// cmdKoBuild builds the application with ko for fast local Docker development
func cmdKoBuild() {
	fmt.Println("🏗️  Building with ko...")
	fmt.Println()

	// Check if ko is installed
	if !deploy.KoInstalled() {
		fmt.Println("📦 Ko not found. Installing...")
		if err := deploy.InstallKo(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to install ko: %v\n", err)
			fmt.Fprintln(os.Stderr, "💡 Install manually: go install github.com/google/ko@latest")
			os.Exit(1)
		}
		fmt.Println()
	}

	// Build with ko
	fmt.Println("🔨 Building image with ko...")
	imageName, err := deploy.BuildLocal(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Ko build failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ Build successful!")
	fmt.Printf("   Image: %s\n", imageName)
	fmt.Printf("   Size: ~12MB (distroless static base)\n")
	fmt.Println()
	fmt.Println("📝 NEXT STEPS:")
	fmt.Println()
	fmt.Println("   Run with Docker:")
	fmt.Printf("   docker run -p 8080:8080 --env-file .env.local %s serve\n", imageName)
	fmt.Println()
	fmt.Println("   Or use with docker-compose:")
	fmt.Printf("   IMAGE=%s docker-compose up\n", imageName)
	fmt.Println()
}
