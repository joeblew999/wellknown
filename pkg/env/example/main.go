package main

import (
	"bytes"
	"fmt"
	"os"

	"filippo.io/age"
	"github.com/joeblew999/wellknown/pkg/env"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "env-demo",
		Usage:                "Registry-driven environment variable management",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "dir",
				Aliases: []string{"C"},
				Usage:   "Change to `DIR` before running command",
				EnvVars: []string{"ENV_WORK_DIR"},
			},
		},
		Before: func(c *cli.Context) error {
			// Change directory if --dir specified
			if dir := c.String("dir"); dir != "" {
				if err := os.Chdir(dir); err != nil {
					return fmt.Errorf("failed to change directory to %s: %w", dir, err)
				}
			}
			// Load .env.local if it exists (silently ignore if missing)
			_ = loadEnvFile(".env.local")
			return nil
		},
		Commands: []*cli.Command{
			// Essential Commands
			{
				Name:     "clean",
				Usage:    "Remove all generated files (.env.local, .env.example, etc.)",
				Category: "1. Essential",
				Action:   func(c *cli.Context) error { cmdClean(); return nil },
			},
			{
				Name:     "setup",
				Usage:    "Create .env.local from registry",
				Category: "1. Essential",
				Action:   func(c *cli.Context) error { cmdSetup(); return nil },
			},
			{
				Name:     "list",
				Usage:    "Show all environment variables with current values",
				Category: "1. Essential",
				Action:   func(c *cli.Context) error { cmdList(); return nil },
			},
			{
				Name:     "validate",
				Usage:    "Validate that all required variables are set",
				Category: "1. Essential",
				Action:   func(c *cli.Context) error { cmdValidate(); return nil },
			},

			// Template Generation
			{
				Name:     "generate-example",
				Usage:    "Generate .env.example template",
				Category: "2. Template Generation",
				Action:   func(c *cli.Context) error { cmdGenerateExample(); return nil },
			},
			{
				Name:     "generate-local",
				Usage:    "Generate .env.local for development",
				Category: "2. Template Generation",
				Action:   func(c *cli.Context) error { cmdGenerateLocal(); return nil },
			},
			{
				Name:     "generate-prod",
				Usage:    "Generate .env.production for deployment",
				Category: "2. Template Generation",
				Action:   func(c *cli.Context) error { cmdGenerateProd(); return nil },
			},
			{
				Name:     "generate-secrets",
				Usage:    "Generate .env.secrets template (secrets only)",
				Category: "2. Template Generation",
				Action:   func(c *cli.Context) error { cmdGenerateSecrets(); return nil },
			},
			{
				Name:     "sync-secrets",
				Usage:    "Merge .env.secrets into .env.local",
				Category: "2. Template Generation",
				Action:   func(c *cli.Context) error { cmdSyncSecrets(); return nil },
			},

			// Age Encryption
			{
				Name:     "age-keygen",
				Usage:    "Generate Age encryption keypair at .age/key.txt",
				Category: "3. Age Encryption (git-safe secrets)",
				Action:   func(c *cli.Context) error { cmdAgeKeygen(); return nil },
			},
			{
				Name:     "age-encrypt",
				Usage:    "Encrypt .env.local → .env.local.age (safe to commit)",
				Category: "3. Age Encryption (git-safe secrets)",
				Action:   func(c *cli.Context) error { cmdAgeEncrypt(); return nil },
			},
			{
				Name:     "age-decrypt",
				Usage:    "Decrypt .env.local.age → .env.local",
				Category: "3. Age Encryption (git-safe secrets)",
				Action:   func(c *cli.Context) error { cmdAgeDecrypt(); return nil },
			},

			// File Sync
			{
				Name:     "dockerfile-docs",
				Usage:    "Generate Dockerfile environment documentation",
				Category: "4. File Sync (Registry → Config Files)",
				Action:   func(c *cli.Context) error { cmdDockerfileDocs(); return nil },
			},
			{
				Name:     "dockerfile-sync",
				Usage:    "Sync Dockerfile environment docs from registry",
				Category: "4. File Sync (Registry → Config Files)",
				Action:   func(c *cli.Context) error { cmdDockerfileSync(); return nil },
			},
			{
				Name:     "fly-sync",
				Usage:    "Sync fly.toml [env] section from registry",
				Category: "4. File Sync (Registry → Config Files)",
				Action:   func(c *cli.Context) error { cmdFlySync(); return nil },
			},
			{
				Name:     "compose-sync",
				Usage:    "Sync docker-compose.yml environment from registry",
				Category: "4. File Sync (Registry → Config Files)",
				Action:   func(c *cli.Context) error { cmdComposeSync(); return nil },
			},

			// Fly.io Deployment
			{
				Name:     "fly-install",
				Usage:    "Install flyctl via go install",
				Category: "5. Fly.io Deployment (No Makefile Required)",
				Action:   func(c *cli.Context) error { cmdFlyInstall(); return nil },
			},
			{
				Name:     "fly-auth",
				Usage:    "Check/login to Fly.io",
				Category: "5. Fly.io Deployment (No Makefile Required)",
				Action:   func(c *cli.Context) error { cmdFlyAuth(); return nil },
			},
			{
				Name:     "fly-launch",
				Usage:    "Create app (reads fly.toml for name/region)",
				Category: "5. Fly.io Deployment (No Makefile Required)",
				Action:   func(c *cli.Context) error { cmdFlyLaunch(); return nil },
			},
			{
				Name:     "fly-volume",
				Usage:    "Create persistent volume",
				Category: "5. Fly.io Deployment (No Makefile Required)",
				Action:   func(c *cli.Context) error { cmdFlyVolume(); return nil },
			},
			{
				Name:     "fly-secrets-import",
				Usage:    "Import secrets from .env.local (registry-driven)",
				Category: "5. Fly.io Deployment (No Makefile Required)",
				Action:   func(c *cli.Context) error { cmdFlySecretsImport(); return nil },
			},
			{
				Name:     "fly-secrets-list",
				Usage:    "List all secrets",
				Category: "5. Fly.io Deployment (No Makefile Required)",
				Action:   func(c *cli.Context) error { cmdFlySecretsList(); return nil },
			},
			{
				Name:     "fly-deploy",
				Usage:    "Deploy to Fly.io",
				Category: "5. Fly.io Deployment (No Makefile Required)",
				Action:   func(c *cli.Context) error { cmdFlyDeploy(); return nil },
			},
			{
				Name:     "fly-status",
				Usage:    "Show deployment status",
				Category: "5. Fly.io Deployment (No Makefile Required)",
				Action:   func(c *cli.Context) error { cmdFlyStatus(); return nil },
			},
			{
				Name:     "fly-logs",
				Usage:    "Tail application logs",
				Category: "5. Fly.io Deployment (No Makefile Required)",
				Action:   func(c *cli.Context) error { cmdFlyLogs(); return nil },
			},
			{
				Name:     "fly-ssh",
				Usage:    "SSH into Fly.io machine",
				Category: "5. Fly.io Deployment (No Makefile Required)",
				Action:   func(c *cli.Context) error { cmdFlySSH(); return nil },
			},
			{
				Name:     "fly-destroy",
				Usage:    "Destroy app (WARNING: destructive)",
				Category: "5. Fly.io Deployment (No Makefile Required)",
				Action:   func(c *cli.Context) error { cmdFlyDestroy(); return nil },
			},

			// Export Formats
			{
				Name:      "export",
				Usage:     "Export variables (formats: simple, docker, systemd, k8s)",
				Category:  "6. Export Formats",
				ArgsUsage: "[format]",
				Action:    func(c *cli.Context) error { cmdExport(); return nil },
			},
			{
				Name:     "fly-secrets",
				Usage:    "Export secrets for Fly.io deployment",
				Category: "6. Export Formats",
				Action:   func(c *cli.Context) error { cmdFlySecrets(); return nil },
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// printUsage is no longer needed - urfave/cli provides built-in help
// Use: go run . --help or go run . help

func cmdList() {
	output := ListEnvVars()
	fmt.Print(output)
}

func cmdValidate() {
	if err := ValidateEnv(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Validation failed: %v\n", err)
		fmt.Fprintln(os.Stderr, "\n💡 Tip: Set missing variables or use 'go run . list' to see all variables")
		os.Exit(1)
	}
	fmt.Println("✅ All required environment variables are set!")
}

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

func cmdGenerateExample() {
	output := GenerateEnvExample()
	fmt.Print(output)
}

func cmdGenerateLocal() {
	output := GenerateEnvLocal()
	fmt.Print(output)
}

func cmdGenerateProd() {
	output := GenerateEnvProduction()
	fmt.Print(output)
}

func cmdGenerateSecrets() {
	output := GenerateEnvSecrets()
	fmt.Print(output)
}

func cmdSyncSecrets() {
	// Use the helper function to merge secrets into .env.local
	if err := MergeSecretsIntoEnv(".env.secrets", "local", ".env.local"); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to sync secrets: %v\n", err)
		fmt.Fprintln(os.Stderr, "\n💡 Make sure .env.secrets exists with your secret values")
		os.Exit(1)
	}

	// Count how many secrets were loaded
	secrets, _ := env.LoadSecrets(env.SecretsSource{
		FilePath:     ".env.secrets",
		TryEncrypted: true,
	})

	fmt.Println("✅ Successfully synced secrets to .env.local")
	fmt.Printf("📝 Merged %d secret values\n", len(secrets))
}

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
	content := GenerateDockerfileEnvDocs()
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
}

func cmdFlySync() {
	content := GenerateFlyTomlEnv()
	fullContent := fmt.Sprintf("# === START AUTO-GENERATED [env] ===\n%s\n# === END AUTO-GENERATED [env] ===", content)

	err := env.SyncFileSection(env.SyncOptions{
		FilePath:     "fly.toml",
		StartMarker:  "# === START AUTO-GENERATED [env] ===",
		EndMarker:    "# === END AUTO-GENERATED [env] ===",
		Content:      fullContent,
		CreateBackup: true,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to sync fly.toml: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Synced fly.toml [env] section from registry")
}

func cmdComposeSync() {
	content := GenerateDockerComposeEnv()
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
}

func cmdClean() {
	files := []string{".env.local", ".env.example", ".env.production", ".env.local.age"}
	removed := 0

	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			if err := os.Remove(file); err != nil {
				fmt.Fprintf(os.Stderr, "⚠️  Failed to remove %s: %v\n", file, err)
			} else {
				removed++
			}
		}
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
	// Generate .env.local directly from registry template
	content := GenerateEnvLocal()
	if err := os.WriteFile(".env.local", []byte(content), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to write .env.local: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Created .env.local from registry")
	fmt.Println("\n📝 Next steps:")
	fmt.Println("  1. Edit .env.local with your real values")
	fmt.Println("  2. Run: go run . validate")
	fmt.Println("\n🔐 Optional - Encrypt for git:")
	fmt.Println("  go run . age-keygen")
	fmt.Println("  go run . age-encrypt")
}

func cmdFlySecrets() {
	// Export only secret variables in simple format for Fly.io
	secrets := GetSecretVars()
	if len(secrets) == 0 {
		fmt.Println("# No secret variables found")
		return
	}

	fmt.Println("# Fly.io Secrets")
	fmt.Println("# Import with: flyctl secrets import < fly-secrets.txt")
	fmt.Println()

	for _, v := range secrets {
		value := os.Getenv(v.Name)
		if value == "" {
			value = v.Default
		}
		if value != "" {
			fmt.Printf("%s=%s\n", v.Name, value)
		}
	}
}

// loadEnvFile loads environment variables from a file (basic implementation)
func loadEnvFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	lines := string(data)
	for _, line := range splitLines(lines) {
		line = trimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}

		parts := splitOnce(line, "=")
		if len(parts) == 2 {
			key := trimSpace(parts[0])
			value := trimSpace(parts[1])
			os.Setenv(key, value)
		}
	}

	return nil
}

// Helper functions (simple implementations to avoid external dependencies)
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func splitOnce(s, sep string) []string {
	idx := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			idx = i
			break
		}
	}
	if idx == 0 {
		return []string{s}
	}
	return []string{s[:idx], s[idx+1:]}
}

func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

// ================================================================
// Age Encryption Commands
// ================================================================

func cmdAgeKeygen() {
	keyPath := ".age/key.txt"

	// Check if key already exists
	if _, err := os.Stat(keyPath); err == nil {
		fmt.Printf("⚠️  Key already exists at %s\n", keyPath)
		fmt.Print("Overwrite? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborted")
			return
		}
	}

	// Create .age directory if it doesn't exist
	ageDir := ".age"
	if err := os.MkdirAll(ageDir, 0700); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to create %s: %v\n", ageDir, err)
		os.Exit(1)
	}

	// Generate new identity
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to generate identity: %v\n", err)
		os.Exit(1)
	}

	// Write identity to file
	identityStr := fmt.Sprintf("# created: %s\n# public key: %s\n%s\n",
		identity.Recipient().String(),
		identity.Recipient().String(),
		identity.String())

	if err := os.WriteFile(keyPath, []byte(identityStr), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to write key: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Generated age keypair")
	fmt.Printf("📁 Saved to: %s\n", keyPath)
	fmt.Printf("\n🔑 Public key: %s\n", identity.Recipient().String())
}

func cmdAgeEncrypt() {
	keyPath := ".age/key.txt"

	// Check if key exists
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "❌ No age key found. Run: go run . age-keygen")
		os.Exit(1)
	}

	// Check if .env.local exists
	if _, err := os.Stat(".env.local"); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "❌ .env.local not found. Run: go run . setup")
		os.Exit(1)
	}

	// Read identity file
	identityFile, err := os.ReadFile(keyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read key: %v\n", err)
		os.Exit(1)
	}

	identities, err := age.ParseIdentities(bytes.NewReader(identityFile))
	if err != nil || len(identities) == 0 {
		fmt.Fprintf(os.Stderr, "❌ Failed to parse identity: %v\n", err)
		os.Exit(1)
	}

	// Get recipient (public key) from identity
	recipient := identities[0].(*age.X25519Identity).Recipient()

	// Read .env.local
	plaintext, err := os.ReadFile(".env.local")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read .env.local: %v\n", err)
		os.Exit(1)
	}

	// Encrypt
	var buf bytes.Buffer
	w, err := age.Encrypt(&buf, recipient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to create encryptor: %v\n", err)
		os.Exit(1)
	}

	if _, err := w.Write(plaintext); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to encrypt: %v\n", err)
		os.Exit(1)
	}

	if err := w.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to finalize encryption: %v\n", err)
		os.Exit(1)
	}

	// Write encrypted file
	if err := os.WriteFile(".env.local.age", buf.Bytes(), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to write .env.local.age: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Encrypted .env.local → .env.local.age")
	fmt.Println("\n💡 Next steps:")
	fmt.Println("  git add .env.local.age")
	fmt.Println("  git commit -m \"Add encrypted secrets\"")
	fmt.Println("\n🔒 Your secrets are now safe to commit!")
}

func cmdAgeDecrypt() {
	// Check if .env.local.age exists
	if _, err := os.Stat(".env.local.age"); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "❌ .env.local.age not found")
		os.Exit(1)
	}

	// Read encrypted file
	encrypted, err := os.ReadFile(".env.local.age")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read .env.local.age: %v\n", err)
		os.Exit(1)
	}

	// Set AGE_IDENTITY to local key path for decryption
	os.Setenv("AGE_IDENTITY", ".age/key.txt")

	// Use env package's DecryptAgeFile (handles key discovery automatically)
	decrypted, err := env.DecryptAgeFile(encrypted)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}

	// Write decrypted file
	if err := os.WriteFile(".env.local", decrypted, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to write .env.local: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Decrypted .env.local.age → .env.local")
	fmt.Println("💡 You can now run: go run . validate")
}

// ================================================================
// Fly.io Commands
// ================================================================

func cmdFlyInstall() {
	if err := FlyInstall(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to install flyctl: %v\n", err)
		os.Exit(1)
	}
}

func cmdFlyAuth() {
	fmt.Println("🔐 Checking Fly.io authentication...")
	if err := FlyAuthWhoami(); err != nil {
		fmt.Println("\n⚠️  Not logged in. Opening browser for login...")
		if err := FlyAuthLogin(); err != nil {
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
	appName, region, err := ParseFlyToml()
	if err != nil {
		// No fly.toml or can't parse - do interactive launch
		fmt.Println("📦 No fly.toml found - launching interactively...")
		if err := FlyLaunch(true); err != nil {
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
	exists, err := FlyAppExists(appName)
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

	if err := FlyAppsCreate(appName, ""); err != nil {
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
	appName, region, err := ParseFlyToml()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read fly.toml: %v\n", err)
		fmt.Fprintln(os.Stderr, "💡 Run 'go run . fly-launch' first")
		os.Exit(1)
	}

	fmt.Printf("💾 Creating volume for app: %s\n", appName)
	fmt.Printf("   Region: %s\n", region)
	fmt.Println("   Size: 1GB")

	if err := FlyVolumesCreate("pb_data", appName, region, 1); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to create volume: %v\n", err)
		fmt.Fprintln(os.Stderr, "💡 Volume may already exist - check with: flyctl volumes list")
		os.Exit(1)
	}

	fmt.Println("✅ Volume created!")
	fmt.Println("💡 Next: go run . fly-secrets-import")
}

func cmdFlySecretsImport() {
	appName, _, err := ParseFlyToml()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read fly.toml: %v\n", err)
		fmt.Fprintln(os.Stderr, "💡 Run 'go run . fly-launch' first")
		os.Exit(1)
	}

	fmt.Printf("🔐 Importing secrets to app: %s\n", appName)
	fmt.Println("   Source: .env.local (or .env.local.age)")
	fmt.Println("   Registry: Only variables marked as Secret=true")

	if err := FlySecretsImport(appName); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to import secrets: %v\n", err)
		fmt.Fprintln(os.Stderr, "\n💡 Make sure:")
		fmt.Fprintln(os.Stderr, "   - .env.local exists with secret values")
		fmt.Fprintln(os.Stderr, "   - You're logged in: go run . fly-auth")
		os.Exit(1)
	}

	fmt.Println("✅ Secrets imported!")
	fmt.Println("💡 Next: go run . fly-deploy")
}

func cmdFlySecretsList() {
	appName, _, err := ParseFlyToml()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read fly.toml: %v\n", err)
		fmt.Fprintln(os.Stderr, "💡 Run 'go run . fly-launch' first")
		os.Exit(1)
	}

	if err := FlySecretsList(appName); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to list secrets: %v\n", err)
		os.Exit(1)
	}
}

func cmdFlyDeploy() {
	appName, _, err := ParseFlyToml()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read fly.toml: %v\n", err)
		fmt.Fprintln(os.Stderr, "💡 Run 'go run . fly-launch' first")
		os.Exit(1)
	}

	fmt.Printf("🚀 Deploying to Fly.io: %s\n", appName)

	if err := FlyDeploy(appName); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Deployment failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Deployed!")
	fmt.Printf("🌐 App URL: https://%s.fly.dev\n", appName)
	fmt.Println("💡 Check status: go run . fly-status")
	fmt.Println("💡 View logs: go run . fly-logs")
}

func cmdFlyStatus() {
	appName, _, err := ParseFlyToml()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read fly.toml: %v\n", err)
		os.Exit(1)
	}

	if err := FlyStatus(appName); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to get status: %v\n", err)
		os.Exit(1)
	}
}

func cmdFlyLogs() {
	appName, _, err := ParseFlyToml()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read fly.toml: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📋 Tailing logs for: %s\n", appName)
	fmt.Println("   Press Ctrl+C to exit")

	if err := FlyLogs(appName, true); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to tail logs: %v\n", err)
		os.Exit(1)
	}
}

func cmdFlySSH() {
	appName, _, err := ParseFlyToml()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read fly.toml: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🔌 Opening SSH console to: %s\n", appName)

	if err := FlySSH(appName); err != nil {
		fmt.Fprintf(os.Stderr, "❌ SSH failed: %v\n", err)
		os.Exit(1)
	}
}

func cmdFlyDestroy() {
	appName, _, err := ParseFlyToml()
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

	if err := FlyAppsDestroy(appName); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to destroy app: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ App destroyed")
}
