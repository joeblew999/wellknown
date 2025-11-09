package main

// commands.go - Low-level direct command implementations
// These commands provide direct access to library functions for fine-grained control.
// For high-level orchestrated workflows, see workflow.go

import (
	"fmt"
	"os"
	"strings"

	"github.com/joeblew999/wellknown/pkg/env"
	"github.com/joeblew999/wellknown/pkg/env/deploy"
	"github.com/joeblew999/wellknown/pkg/env/scaffold"
)

// ================================================================
// Setup & Validation Commands
// ================================================================

func cmdList() {
	output := AppRegistry.GenerateEnvList("Sample Application Environment Variables")
	fmt.Print(output)
}

func cmdValidate() {
	if err := AppRegistry.ValidateRequired(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Validation failed: %v\n", err)
		fmt.Fprintln(os.Stderr, "\n💡 Tip: Set missing variables or use 'go run . list' to see all variables")
		os.Exit(1)
	}
	fmt.Println("✅ All required environment variables are set!")
}

func cmdClean() {
	// Use library function to remove all environment files
	removed, err := env.CleanEnvironmentFiles()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to clean: %v\n", err)
		os.Exit(1)
	}

	if removed == 0 {
		fmt.Println("✅ Already clean - no generated files found")
	} else {
		fmt.Printf("✅ Cleaned %d generated file(s)\n", removed)
		fmt.Println("\n📝 Next step:")
		fmt.Println("  go run . setup")
	}
}

func cmdSetup() {
	// Use library function to create .env.local
	if err := env.SetupEnvironment(AppRegistry, env.Local, "Sample Application"); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to create %s: %v\n", env.Local.FileName, err)
		os.Exit(1)
	}

	fmt.Printf("✅ Created %s from registry\n", env.Local.FileName)
	fmt.Println("\n📝 Next steps:")
	fmt.Printf("  1. Edit %s with your real values\n", env.Local.FileName)
	fmt.Println("  2. Run: go run . validate")
	fmt.Println("\n🔐 Optional - Encrypt for git:")
	fmt.Println("  go run . age-keygen")
	fmt.Println("  go run . age-encrypt")
	fmt.Printf("\n⚠️  DO NOT commit %s directly (contains secrets when filled)\n", env.Local.FileName)
	fmt.Printf("   Encrypt first, then commit %s\n", env.Local.EncryptedFileName())
}

func cmdSetupProd() {
	// Use library function to create .env.production
	if err := env.SetupEnvironment(AppRegistry, env.Production, "Sample Application"); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to create %s: %v\n", env.Production.FileName, err)
		os.Exit(1)
	}

	fmt.Printf("✅ Created %s from registry\n", env.Production.FileName)
	fmt.Println("\n📝 Next steps:")
	fmt.Printf("  1. Edit %s with your production values\n", env.Production.FileName)
	fmt.Println("  2. Run: go run . age-encrypt")
	fmt.Println("  3. Run: go run . fly-secrets-import")
	fmt.Printf("\n⚠️  DO NOT commit %s directly (contains secrets when filled)\n", env.Production.FileName)
	fmt.Printf("   Encrypt first, then commit %s\n", env.Production.EncryptedFileName())
}

// ================================================================
// Template Generation Commands
// ================================================================

func cmdGenerateExample() {
	output := AppRegistry.GenerateEnvExample("Sample Application")
	fmt.Print(output)
}

func cmdGenerateLocal() {
	output := env.Local.Generate(AppRegistry, "Sample Application")
	fmt.Print(output)
}

func cmdGenerateProd() {
	output := env.Production.Generate(AppRegistry, "Sample Application")
	fmt.Print(output)
}

func cmdGenerateSecrets() {
	// Get only secret vars from registry
	secrets := AppRegistry.GetSecrets()
	var filteredVars []env.EnvVar
	for _, v := range secrets {
		filteredVars = append(filteredVars, v)
	}

	// Create temporary registry with only secrets
	secretsRegistry := env.NewRegistry(filteredVars)

	output := env.Secrets.Generate(secretsRegistry, "Sample Application")
	fmt.Print(output)
}

// ================================================================
// Secrets Sync Commands
// ================================================================

func cmdSyncSecrets() {
	// Use library function to sync secrets to local environment
	result, err := env.SyncSecretsToEnvironment(env.SecretsSyncOptions{
		Registry:    AppRegistry,
		TargetEnv:   env.Local,
		AppName:     "Sample Application",
		AutoResolve: true, // Use ResolveSecretsFile() to find best secrets source
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to sync secrets: %v\n", err)
		if strings.Contains(err.Error(), "no secrets file found") {
			fmt.Fprintln(os.Stderr, "\n💡 Create one of these files:")
			fmt.Fprintf(os.Stderr, "   - %s (recommended for local development)\n", env.SecretsLocal.FileName)
			fmt.Fprintf(os.Stderr, "   - %s (encrypted version)\n", env.SecretsLocal.EncryptedFileName())
		}
		os.Exit(1)
	}

	// Display fallback warning if used
	if result.UsedFallback {
		fmt.Printf("⚠️  Using %s (consider creating %s for local development)\n",
			result.FallbackFile, env.SecretsLocal.FileName)
	}

	fmt.Printf("✅ Successfully synced secrets from %s to %s\n", result.SecretsFile, result.TargetFile)
	fmt.Printf("📝 Merged %d secret values\n", result.SecretsCount)
}

func cmdSyncSecretsProd() {
	// Use library function to sync secrets to production environment
	result, err := env.SyncSecretsToEnvironment(env.SecretsSyncOptions{
		Registry:    AppRegistry,
		TargetEnv:   env.Production,
		AppName:     "Sample Application",
		AutoResolve: true, // Use ResolveSecretsFile() to find best secrets source
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to sync secrets: %v\n", err)
		if strings.Contains(err.Error(), "no secrets file found") {
			fmt.Fprintln(os.Stderr, "\n💡 Create one of these files:")
			fmt.Fprintf(os.Stderr, "   - %s (recommended for production)\n", env.SecretsProduction.FileName)
			fmt.Fprintf(os.Stderr, "   - %s (encrypted version)\n", env.SecretsProduction.EncryptedFileName())
		}
		os.Exit(1)
	}

	// Display fallback warning if used
	if result.UsedFallback {
		fmt.Printf("⚠️  Using %s (consider creating %s for production)\n",
			result.FallbackFile, env.SecretsProduction.FileName)
	}

	fmt.Printf("✅ Successfully synced secrets from %s to %s\n", result.SecretsFile, result.TargetFile)
	fmt.Printf("📝 Merged %d secret values\n", result.SecretsCount)
}

// ================================================================
// Config File Sync Commands
// ================================================================

func cmdDockerfileDocs() {
	output := AppRegistry.GenerateDockerfileDocs(env.DockerfileDocsOptions{
		AppName:            "Sample Application",
		UpdateCommand:      "go run . dockerfile-docs",
		DeploymentPlatform: "Docker",
		NonSecretEnvSource: "Dockerfile ENV or docker-compose.yml",
		SecretSource:       "Docker secrets or environment",
		SyncCommand:        "docker-compose up",
	})
	fmt.Print(output)
}

func cmdDockerfileSync() {
	content := AppRegistry.GenerateDockerfileDocs(env.DockerfileDocsOptions{
		AppName:            "Sample Application",
		UpdateCommand:      "go run . dockerfile-sync",
		DeploymentPlatform: "Docker",
		NonSecretEnvSource: "docker-compose.yml or Dockerfile ENV",
		SecretSource:       "Docker secrets or environment",
		SyncCommand:        "docker-compose up",
	})
	fullContent := fmt.Sprintf("# === START AUTO-GENERATED ENV DOCS ===\n%s# === END AUTO-GENERATED ENV DOCS ===", content)

	err := env.SyncFileSection(env.SyncOptions{
		FilePath:     "Dockerfile",
		StartMarker:  "# === START AUTO-GENERATED ENV DOCS ===",
		EndMarker:    "# === END AUTO-GENERATED ENV DOCS ===",
		Content:      fullContent,
		CreateBackup: true,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to sync Dockerfile: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Synced Dockerfile environment docs from registry")
	fmt.Println("\n💡 Safe to commit (no secrets):")
	fmt.Println("  git add Dockerfile")
	fmt.Println("  git commit -m \"sync: update Dockerfile from registry\"")
}

func cmdFlySync() {
	// Sync [env] section
	content := AppRegistry.GenerateTOMLEnv("env", []string{
		"# AUTO-GENERATED - DO NOT EDIT MANUALLY",
		"# Update with: go run . fly-sync",
	})
	fullContent := fmt.Sprintf("# === START AUTO-GENERATED [env] ===\n%s\n# === END AUTO-GENERATED [env] ===", content)

	err := env.SyncFileSection(env.SyncOptions{
		FilePath:     "fly.toml",
		StartMarker:  "# === START AUTO-GENERATED [env] ===",
		EndMarker:    "# === END AUTO-GENERATED [env] ===",
		Content:      fullContent,
		CreateBackup: true,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to sync fly.toml [env]: %v\n", err)
		os.Exit(1)
	}

	// Sync secrets comment list
	secretsList := AppRegistry.GenerateTOMLSecretsList("go run . fly-secrets-import")
	err = env.SyncFileSection(env.SyncOptions{
		FilePath:     "fly.toml",
		StartMarker:  "# === START AUTO-GENERATED SECRETS LIST ===",
		EndMarker:    "# === END AUTO-GENERATED SECRETS LIST ===",
		Content:      secretsList,
		CreateBackup: false, // Already created backup above
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to sync fly.toml secrets list: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Synced fly.toml [env] section and secrets list from registry")
	fmt.Println("\n💡 Safe to commit (no secrets):")
	fmt.Println("  git add fly.toml")
	fmt.Println("  git commit -m \"sync: update fly.toml from registry\"")
}

func cmdComposeSync() {
	content := AppRegistry.GenerateDockerComposeEnv([]string{
		"AUTO-GENERATED - DO NOT EDIT MANUALLY",
		"Update with: go run . compose-sync",
	})
	fullContent := fmt.Sprintf("    # === START AUTO-GENERATED environment ===\n%s\n    # === END AUTO-GENERATED environment ===", content)

	err := env.SyncFileSection(env.SyncOptions{
		FilePath:     "docker-compose.yml",
		StartMarker:  "    # === START AUTO-GENERATED environment ===",
		EndMarker:    "    # === END AUTO-GENERATED environment ===",
		Content:      fullContent,
		CreateBackup: true,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to sync docker-compose.yml: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Synced docker-compose.yml environment section from registry")
	fmt.Println("\n💡 Safe to commit (no secrets):")
	fmt.Println("  git add docker-compose.yml")
	fmt.Println("  git commit -m \"sync: update docker-compose.yml from registry\"")
}

// ================================================================
// Age Encryption Commands
// ================================================================

func cmdAgeKeygen() {
	// Use library function with interactive prompt
	result, err := env.GenerateAgeKey(env.KeygenOptions{
		KeyPath: env.DefaultAgeKeyPath,
		OverwritePrompt: func() bool {
			fmt.Printf("⚠️  Key already exists at %s\n", env.DefaultAgeKeyPath)
			fmt.Print("Overwrite? (y/N): ")
			var response string
			fmt.Scanln(&response)
			return response == "y" || response == "Y"
		},
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to generate key: %v\n", err)
		os.Exit(1)
	}

	if !result.Created {
		fmt.Println("Aborted")
		return
	}

	fmt.Println("✅ Generated age keypair")
	fmt.Printf("📁 Saved to: %s\n", result.KeyPath)
	fmt.Printf("\n🔑 Public key: %s\n", result.PublicKey)
	fmt.Printf("\n⚠️  CRITICAL: NEVER commit %s to git!\n", env.DefaultAgeKeyPath)
	fmt.Println("   Store this key in your password manager (1Password, LastPass, etc.)")
	fmt.Println("   For CI/CD, add it as a GitHub Secret: AGE_KEY")
}

func cmdAgeEncrypt() {
	// Use library function for encryption
	result, err := env.EncryptEnvironments(env.EncryptionOptions{
		KeyPath:      env.DefaultAgeKeyPath,
		Environments: env.AllEnvironmentFiles(),
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to encrypt: %v\n", err)

		// Provide helpful guidance
		if strings.Contains(err.Error(), "no Age key") {
			fmt.Fprintln(os.Stderr, "\n💡 Generate a key first:")
			fmt.Fprintln(os.Stderr, "   go run . age-keygen")
		}

		os.Exit(1)
	}

	// Display results
	for _, err := range result.Errors {
		fmt.Fprintf(os.Stderr, "⚠️  %v\n", err)
	}

	for _, file := range result.ProcessedFiles {
		baseName := strings.TrimSuffix(file, ".age")
		fmt.Printf("✅ Encrypted %s → %s\n", baseName, file)
	}

	if len(result.ProcessedFiles) == 0 {
		fmt.Fprintln(os.Stderr, "❌ No environment files found to encrypt")
		fmt.Fprintln(os.Stderr, "   Run: go run . setup  (for .env.local)")
		fmt.Fprintln(os.Stderr, "   Run: go run . setup-prod  (for .env.production)")
		fmt.Fprintln(os.Stderr, "   Run: go run . generate-secrets > .env.secrets.local")
		fmt.Fprintln(os.Stderr, "   Run: go run . generate-secrets > .env.secrets.production")
		os.Exit(1)
	}

	fmt.Printf("\n✅ Encrypted %d file(s)\n", len(result.ProcessedFiles))
	fmt.Println("\n💡 Safe to commit encrypted files:")
	fmt.Println("  git add *.age")
	fmt.Println("  git commit -m \"chore: update encrypted environments\"")
	fmt.Println("\n🔒 Encrypted *.age files are SAFE to commit!")
	fmt.Println("⚠️  NEVER commit plaintext: .env.local, .env.production, .env.secrets.*")
	fmt.Printf("⚠️  NEVER commit: %s\n", env.DefaultAgeKeyPath)
}

func cmdAgeDecrypt() {
	// Use library function for decryption
	result, err := env.DecryptEnvironments(env.EncryptionOptions{
		KeyPath:      env.DefaultAgeKeyPath,
		Environments: env.AllEnvironmentFiles(),
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to decrypt: %v\n", err)
		// Provide helpful guidance
		if strings.Contains(err.Error(), "no Age identities") {
			fmt.Fprintln(os.Stderr, "\n💡 Generate a key first:")
			fmt.Fprintln(os.Stderr, "   go run . age-keygen")
		}
		os.Exit(1)
	}

	// Display any non-fatal errors
	for _, err := range result.Errors {
		fmt.Fprintf(os.Stderr, "⚠️  %v\n", err)
	}

	// Show successfully decrypted files
	for _, file := range result.ProcessedFiles {
		encryptedFile := file + ".age"
		fmt.Printf("✅ Decrypted %s → %s\n", encryptedFile, file)
	}

	if len(result.ProcessedFiles) == 0 {
		fmt.Fprintln(os.Stderr, "❌ No encrypted files found to decrypt")
		fmt.Fprintln(os.Stderr, "   Looking for: .env.local.age or .env.production.age")
		os.Exit(1)
	}

	fmt.Printf("\n✅ Decrypted %d file(s)\n", len(result.ProcessedFiles))
	fmt.Println("💡 You can now run: go run . validate")
}

func cmdInstallGitHooks() {
	// Use library function to install git hooks
	result, err := scaffold.InstallGitHooks(scaffold.GitHooksOptions{
		OverwritePrompt: func() bool {
			fmt.Printf("⚠️  Pre-commit hook already exists at .git/hooks/pre-commit\n")
			fmt.Print("Overwrite? (y/N): ")
			var response string
			fmt.Scanln(&response)
			return response == "y" || response == "Y"
		},
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to install git hooks: %v\n", err)
		if strings.Contains(err.Error(), "not a git repository") {
			fmt.Fprintln(os.Stderr, "\n💡 Initialize git first:")
			fmt.Fprintln(os.Stderr, "   git init")
		}
		os.Exit(1)
	}

	if !result.Installed {
		fmt.Println("Aborted")
		return
	}

	fmt.Println("✅ Installed pre-commit hook")
	fmt.Printf("📁 Location: %s\n", result.HookPath)
	fmt.Println("\n🔒 Git will now BLOCK commits of:")
	fmt.Println("   - Plaintext .env files (.env.local, .env.production, .env.secrets.*)")
	fmt.Printf("   - Age encryption key (%s)\n", env.DefaultAgeKeyPath)
	fmt.Println("\n✅ Encrypted *.age files are still allowed")
}

// ================================================================
// Export & Format Commands
// ================================================================

func cmdExport() {
	format := "simple"
	if len(os.Args) > 2 {
		format = os.Args[2]
	}

	var exportFormat env.ExportFormat
	switch format {
	case "simple":
		exportFormat = env.FormatSimple
	case "docker":
		exportFormat = env.FormatDocker
	case "systemd":
		exportFormat = env.FormatSystemd
	case "k8s", "kubernetes":
		exportFormat = env.FormatK8s
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s\n", format)
		fmt.Fprintln(os.Stderr, "Available formats: simple, docker, systemd, k8s")
		os.Exit(1)
	}

	output := AppRegistry.Export(env.ExportOptions{
		Format: exportFormat,
	})
	fmt.Print(output)
}

func cmdFlySecrets() {
	// Use library function to export secrets for Fly.io
	output, err := deploy.ExportSecretsForFly(AppRegistry, env.Production.FileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to export secrets: %v\n", err)
		fmt.Fprintln(os.Stderr, "\n💡 Make sure you have a .env.production file with secrets")
		os.Exit(1)
	}

	if output == "" {
		fmt.Println("# No secret values found")
		fmt.Println("# Make sure your .env.production file contains secret values")
		return
	}

	fmt.Println("# Fly.io Secrets")
	fmt.Println("# Import with: flyctl secrets import < fly-secrets.txt")
	fmt.Println()
	fmt.Print(output)
}

// ================================================================
// Fly.io Deployment Commands
// ================================================================

func cmdFlyInstall() {
	if err := deploy.Install(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to install flyctl: %v\n", err)
		os.Exit(1)
	}
}

func cmdFlyAuth() {
	fmt.Println("🔐 Checking Fly.io authentication...")
	if err := deploy.AuthWhoami(); err != nil {
		fmt.Println("\n⚠️  Not logged in. Opening browser for login...")
		if err := deploy.AuthLogin(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Login failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Logged in successfully")
	} else {
		fmt.Println("✅ Already logged in")
	}
}

func cmdFlyLaunch() {
	// Parse fly.toml to check if app name is configured
	appName, region, err := deploy.ReadFlyTomlConfig()
	if err != nil {
		// No fly.toml or can't parse - do interactive launch
		fmt.Println("📦 No fly.toml found - launching interactively...")
		if err := deploy.Launch(true); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Launch failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ App created! fly.toml generated")
		fmt.Println("💡 Next steps:")
		fmt.Println("   1. go run . fly-volume")
		fmt.Println("   2. go run . fly-secrets-import")
		fmt.Println("   3. go run . fly-deploy")
		return
	}

	// Check if app already exists
	exists, err := deploy.AppExists(appName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Could not check if app exists: %v\n", err)
	}

	if exists {
		fmt.Printf("✅ App already exists: %s\n", appName)
		fmt.Printf("   Region: %s\n", region)
		fmt.Println("💡 App is ready - skip to: go run . fly-volume")
		return
	}

	// Create app
	fmt.Printf("📦 Creating app: %s\n", appName)
	fmt.Printf("   Region: %s\n", region)

	if err := deploy.AppsCreate(appName, ""); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to create app: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ App created!")
	fmt.Println("💡 Next steps:")
	fmt.Println("   1. go run . fly-volume")
	fmt.Println("   2. go run . fly-secrets-import")
	fmt.Println("   3. go run . fly-deploy")
}

func cmdFlyVolume() {
	appName, region, err := deploy.ReadFlyTomlConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read fly.toml: %v\n", err)
		fmt.Fprintln(os.Stderr, "💡 Run 'go run . fly-launch' first")
		os.Exit(1)
	}

	fmt.Printf("💾 Creating volume for app: %s\n", appName)
	fmt.Printf("   Region: %s\n", region)
	fmt.Println("   Size: 1GB")

	if err := deploy.VolumesCreate("pb_data", appName, region, 1); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to create volume: %v\n", err)
		fmt.Fprintln(os.Stderr, "💡 Volume may already exist - check with: flyctl volumes list")
		os.Exit(1)
	}

	fmt.Println("✅ Volume created!")
	fmt.Println("💡 Next: go run . fly-secrets-import")
}

func cmdFlySecretsImport() {
	appName, _, err := deploy.ReadFlyTomlConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read fly.toml: %v\n", err)
		fmt.Fprintln(os.Stderr, "💡 Run 'go run . fly-launch' first")
		os.Exit(1)
	}

	// Prefer .env.production, fallback to .env.local
	var envFile string
	if env.Production.Exists() {
		envFile = env.Production.FileName
	} else if env.Local.Exists() {
		envFile = env.Local.FileName
	} else {
		fmt.Fprintln(os.Stderr, "❌ No environment file found")
		fmt.Fprintln(os.Stderr, "\n💡 Create one with:")
		fmt.Fprintln(os.Stderr, "   go run . generate-prod  (for production)")
		fmt.Fprintln(os.Stderr, "   go run . setup          (for local/dev)")
		os.Exit(1)
	}

	fmt.Printf("🔐 Importing secrets to app: %s\n", appName)
	fmt.Printf("   Source: %s (or %s.age)\n", envFile, envFile)
	fmt.Println("   Registry: Only variables marked as Secret=true")

	if err := deploy.SecretsImport(AppRegistry, envFile, appName); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to import secrets: %v\n", err)
		fmt.Fprintln(os.Stderr, "\n💡 Make sure:")
		fmt.Fprintf(os.Stderr, "   - %s exists with secret values\n", envFile)
		fmt.Fprintln(os.Stderr, "   - You're logged in: go run . fly-auth")
		os.Exit(1)
	}

	fmt.Println("✅ Secrets imported!")
	fmt.Println("💡 Next: go run . fly-deploy")
}

func cmdFlySecretsList() {
	appName, _, err := deploy.ReadFlyTomlConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read fly.toml: %v\n", err)
		fmt.Fprintln(os.Stderr, "💡 Run 'go run . fly-launch' first")
		os.Exit(1)
	}

	if err := deploy.SecretsList(appName); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to list secrets: %v\n", err)
		os.Exit(1)
	}
}

func cmdFlyDeploy() {
	appName, _, err := deploy.ReadFlyTomlConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read fly.toml: %v\n", err)
		fmt.Fprintln(os.Stderr, "💡 Run 'go run . fly-launch' first")
		os.Exit(1)
	}

	fmt.Printf("🚀 Deploying to Fly.io: %s\n", appName)

	if err := deploy.Deploy(appName); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Deployment failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Deployed!")
	fmt.Printf("🌐 App URL: https://%s.fly.dev\n", appName)
	fmt.Println("💡 Check status: go run . fly-status")
	fmt.Println("💡 View logs: go run . fly-logs")
}

func cmdFlyStatus() {
	appName, _, err := deploy.ReadFlyTomlConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read fly.toml: %v\n", err)
		os.Exit(1)
	}

	if err := deploy.Status(appName); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to get status: %v\n", err)
		os.Exit(1)
	}
}

func cmdFlyLogs() {
	appName, _, err := deploy.ReadFlyTomlConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read fly.toml: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📋 Tailing logs for: %s\n", appName)
	fmt.Println("   Press Ctrl+C to exit")

	if err := deploy.Logs(appName, true); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to tail logs: %v\n", err)
		os.Exit(1)
	}
}

func cmdFlySSH() {
	appName, _, err := deploy.ReadFlyTomlConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read fly.toml: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🔌 Opening SSH console to: %s\n", appName)

	if err := deploy.SSH(appName); err != nil {
		fmt.Fprintf(os.Stderr, "❌ SSH failed: %v\n", err)
		os.Exit(1)
	}
}

func cmdFlyDestroy() {
	appName, _, err := deploy.ReadFlyTomlConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read fly.toml: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("⚠️  WARNING: This will DESTROY the Fly.io app and ALL data!")
	fmt.Printf("   App: %s\n", appName)
	fmt.Print("\nType the app name to confirm: ")

	var confirmation string
	fmt.Scanln(&confirmation)

	if confirmation != appName {
		fmt.Println("❌ Confirmation failed - app name did not match")
		os.Exit(1)
	}

	if err := deploy.AppsDestroy(appName); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to destroy app: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ App destroyed")
}

// ================================================================
// Project Initialization
// ================================================================

func cmdInitProject(appName, packageName string, example, force bool) error {
	fmt.Println("🚀 Initializing new project...")
	fmt.Println()

	// Determine template type
	templateType := "minimal"
	if example {
		templateType = "full"
	}

	// Build generator options
	opts := scaffold.GeneratorOptions{
		Dir:         ".", // Current directory (already changed by --dir flag if specified)
		AppName:     appName,
		PackageName: packageName,
		Template:    templateType,
		Force:       force,
		ImportPath:  "github.com/joeblew999/wellknown",
	}

	// Check if registry already exists
	if scaffold.RegistryExists(opts.Dir) && !opts.Force {
		fmt.Fprintf(os.Stderr, "❌ registry.go already exists\n")
		fmt.Fprintf(os.Stderr, "   Use --force to overwrite\n")
		return fmt.Errorf("registry.go already exists")
	}

	// Generate registry
	fmt.Println("📝 Step 1/2: Generating registry.go")
	if err := scaffold.GenerateRegistry(opts); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to generate registry: %v\n", err)
		return err
	}
	fmt.Printf("   ✅ Created registry.go (%s template)\n", templateType)
	fmt.Println()

	// Step 2: Run sync-registry to generate all files
	fmt.Println("📝 Step 2/2: Running sync-registry")
	fmt.Println()

	// Note: We can't call cmdSyncRegistry() here because it expects AppRegistry to exist
	// The user needs to build/run with the new registry.go first

	fmt.Println("✅ Project initialized successfully!")
	fmt.Println()
	fmt.Println("📝 NEXT STEPS:")
	fmt.Println("   1. Review and customize registry.go")
	fmt.Println("   2. Run: go run . sync-registry")
	fmt.Println("   3. Edit .env.secrets.local and .env.secrets.production")
	fmt.Println("   4. Run: go run . sync-environments")
	fmt.Println("   5. Run: go run . finalize")
	fmt.Println()
	fmt.Println("💡 TIP: See WORKFLOW.md for the complete workflow guide")

	return nil
}
