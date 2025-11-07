package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/ghupdate"
	"github.com/pocketbase/pocketbase/plugins/jsvm"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/archive"
	"github.com/pocketbase/pocketbase/tools/osutils"
	"github.com/spf13/cobra"

	_ "github.com/joeblew999/wellknown/pkg/cmd/pocketbase/pb_migrations" // Import migrations
	"github.com/joeblew999/wellknown/pkg/cmd/mcp"
	testdatagen "github.com/joeblew999/wellknown/pkg/cmd/testdata-gen"
	wellknown "github.com/joeblew999/wellknown/pkg/pb"
)

func main() {
	// Load .env.local if it exists (for local development)
	// Silently ignore if file doesn't exist (production uses real env vars)
	_ = godotenv.Load(".env.local")

	// Check if this is a utility command that doesn't need validation
	// (env list/validate/generate commands should work even without credentials)
	isUtilityCommand := len(os.Args) >= 2 &&
		(os.Args[1] == "env" || os.Args[1] == "mcp" || os.Args[1] == "testdata-gen" ||
		 os.Args[1] == "help" || os.Args[1] == "--help" || os.Args[1] == "-h")

	// Only validate for serve/migrate commands (not utility commands)
	if !isUtilityCommand {
		if err := wellknown.ValidateEnv(); err != nil {
			log.Fatalf("❌ Environment validation failed: %v\n\n"+
				"💡 Run 'go run . env list' to see all required variables\n"+
				"💡 Run 'go run . env validate' for detailed validation", err)
		}
	}

	// Load configuration
	cfg, err := wellknown.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create Wellknown app (wraps PocketBase with custom routes)
	wk, err := wellknown.NewWithConfig(cfg)
	if err != nil {
		log.Fatalf("Failed to create wellknown app: %v", err)
	}

	// Get the underlying PocketBase app
	app := wk.PocketBase

	// ---------------------------------------------------------------
	// Optional plugin flags (following PocketBase examples/base pattern):
	// ---------------------------------------------------------------

	var hooksDir string
	app.RootCmd.PersistentFlags().StringVar(
		&hooksDir,
		"hooksDir",
		"",
		"the directory with the JS app hooks",
	)

	var hooksWatch bool
	app.RootCmd.PersistentFlags().BoolVar(
		&hooksWatch,
		"hooksWatch",
		true,
		"auto restart the app on pb_hooks file change",
	)

	var hooksPool int
	app.RootCmd.PersistentFlags().IntVar(
		&hooksPool,
		"hooksPool",
		25,
		"the total prewarm goja.Runtime instances for the JS app hooks execution",
	)

	var migrationsDir string
	if osutils.IsProbablyGoRun() {
		migrationsDir = filepath.Join(app.DataDir(), "../migrations")
	}
	app.RootCmd.PersistentFlags().StringVar(
		&migrationsDir,
		"migrationsDir",
		migrationsDir,
		"the directory with the user defined migrations (default to pb_data/../pb_migrations)",
	)

	// CRITICAL: Parse flags BEFORE registering plugins (following examples/base pattern)
	app.RootCmd.ParseFlags(os.Args[1:])

	// ---------------------------------------------------------------
	// Plugins and hooks (ORDER MATTERS!):
	// ---------------------------------------------------------------
	// Plugin Load Order Documentation:
	//
	// 1. jsvm - MUST be first
	//    - Registers JavaScript VM for pb_hooks and pb_migrations
	//    - Provides runtime for user-defined hooks and migrations
	//    - Other plugins may depend on hook system being available
	//
	// 2. migratecmd - MUST be after jsvm
	//    - Registers 'migrate' command for database migrations
	//    - Depends on jsvm for JavaScript migration support
	//    - Runs before server starts (via 'migrate up' command)
	//
	// 3. ghupdate/localupdate - Can be anywhere after jsvm
	//    - Registers 'update' command for binary updates
	//    - Independent of other plugins
	//    - Never runs during normal server operation
	//
	// 4. Custom commands (env, mcp, testdata-gen) - Can be anywhere
	//    - Independent utility commands
	//    - No dependencies on other plugins
	//
	// 5. TLS configuration - MUST be after all commands registered
	//    - Configures HTTPS via OnServe hook
	//    - Hook execution order matters (runs before route registration)
	//
	// NOTE: Wellknown route registration happens in bindAppHooks()
	//       which is called in NewWithConfig(). Routes are registered
	//       via OnServe() hooks when the server starts.

	// 1. Load jsvm (hooks and migrations)
	jsvm.MustRegister(app, jsvm.Config{
		MigrationsDir: migrationsDir,
		HooksDir:      hooksDir,
		HooksWatch:    hooksWatch,
		HooksPoolSize: hooksPool,
	})

	// 2. Register migrate command
	migrateConfig := migratecmd.Config{Dir: migrationsDir}
	if !osutils.IsProbablyGoRun() {
		migrateConfig.TemplateLang = migratecmd.TemplateLangJS
	}
	migratecmd.MustRegister(app, app.RootCmd, migrateConfig)

	// 3. Register update command (local or GitHub based on env var)
	updateSource := wellknown.EnvRegistry.ByName("UPDATE_SOURCE").GetString()
	if updateSource == "local" {
		// Local update for testing (sources from .dist folder)
		app.RootCmd.AddCommand(newLocalUpdateCommand(app))
	} else {
		// GitHub selfupdate plugin (production)
		ghupdate.MustRegister(app, app.RootCmd, ghupdate.Config{
			Owner:             "joeblew999",
			Repo:              "wellknown",
			ArchiveExecutable: "wellknown-pb",
		})
	}

	// 4. Register custom utility commands (order-independent)
	app.RootCmd.AddCommand(newEnvCommand(app))    // Environment variable management
	app.RootCmd.AddCommand(mcp.NewCommand())      // MCP server for Claude Desktop
	app.RootCmd.AddCommand(testdatagen.NewCommand()) // Test data generation

	// 5. Configure TLS if HTTPS is enabled (development only with mkcert)
	// Production uses Fly.io's native Let's Encrypt HTTPS
	// This registers an OnServe hook, so it must come after all command registration
	if wellknown.EnvRegistry.ByName("HTTPS_ENABLED").GetBool() {
		if err := configureTLS(app, cfg); err != nil {
			log.Fatalf("Failed to configure TLS: %v", err)
		}
	}

	// Start the app (this is blocking)
	// Cobra will dispatch to the appropriate command (serve, migrate, etc.)
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// configureTLS sets up HTTPS with custom certificates for development
func configureTLS(app core.App, cfg *wellknown.Config) error {
	certFile := wellknown.EnvRegistry.ByName("CERT_FILE").GetString()
	keyFile := wellknown.EnvRegistry.ByName("KEY_FILE").GetString()

	// Load certificate and private key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificates (run: make certs-generate): %w", err)
	}

	log.Println("🔐 TLS Configuration:")
	log.Printf("   • Certificate: %s", certFile)
	log.Printf("   • Private Key: %s", keyFile)
	log.Println("   • Mode: Development (mkcert)")
	log.Println("   ⚠️  DO NOT USE IN PRODUCTION")

	// Register OnServe hook to configure TLS and show custom banner
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// Configure TLS
		e.Server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}

		// Show custom banner with correct URLs
		showCustomBanner(e.Server.Addr, true)

		return e.Next()
	})

	return nil
}

// showCustomBanner displays server info with functional URLs
func showCustomBanner(addr string, isHTTPS bool) {
	protocol := "http"
	if isHTTPS {
		protocol = "https"
	}

	// Extract port from address
	_, port, _ := net.SplitHostPort(addr)
	if port == "" {
		port = "8090" // fallback
	}

	// Calculate display URLs
	localURL := fmt.Sprintf("%s://localhost:%s", protocol, port)
	localIP := getLocalIP()
	mobileURL := fmt.Sprintf("%s://%s:%s", protocol, localIP, port)
	mobileURLTrust := fmt.Sprintf("%s://%s", protocol, localIP) // For iOS trust (no port)

	log.Println("")
	log.Println("🎉 Wellknown Server Started")
	log.Printf("   Local:  %s", localURL)
	log.Printf("   Mobile: %s", mobileURL)

	// Show iOS-specific instructions for HTTPS with non-standard port
	if isHTTPS && port != "443" {
		log.Println("")
		log.Println("   📱 iOS First-Time Setup:")
		log.Printf("      1. Visit: %s (to trust certificate)", mobileURLTrust)
		log.Printf("      2. Then:  %s (to access server)", mobileURL)
	}

	log.Println("")
	log.Printf("   API:    %s/api/", localURL)
	log.Printf("   Admin:  %s/_/", localURL)
	log.Println("")
}

// getLocalIP returns the local network IP address
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}

// ---------------------------------------------------------------
// Environment Variable Management Command
// ---------------------------------------------------------------

func newEnvCommand(app core.App) *cobra.Command {
	envCmd := &cobra.Command{
		Use:   "env",
		Short: "Environment variable management",
		Long:  "Manage and export environment variables for deployment",
	}

	// Sub-command: env export-secrets
	exportCmd := &cobra.Command{
		Use:   "export-secrets",
		Short: "Export secrets for flyctl secrets import",
		Long: `Export environment variables marked as secrets in NAME=VALUE format.
This output can be piped directly to 'flyctl secrets import'.

Example:
  . ./.env && ./wellknown env export-secrets | flyctl secrets import`,
		RunE: func(cmd *cobra.Command, args []string) error {
			output := wellknown.ExportSecretsFormat()
			if output == "" {
				fmt.Fprintln(os.Stderr, "⚠️  No secrets found in environment")
				return fmt.Errorf("no secrets found in environment")
			}
			fmt.Print(output)
			return nil
		},
	}

	// Sub-command: env list
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all environment variables and their status",
		Long: `Display all registered environment variables with their current values.
Secret values are masked for security.

This shows the complete environment variable registry from pkg/pb/env.go.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print(wellknown.ListEnvVars())
			return nil
		},
	}

	// Sub-command: env validate
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate required environment variables",
		Long: `Check if all required environment variables are set.
Returns an error if any required variables are missing.

Required variables are marked with Required: true in pkg/pb/env.go.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := wellknown.ValidateEnv(); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Validation failed: %v\n", err)
				return err
			}
			fmt.Println("✅ All required environment variables are set")
			return nil
		},
	}

	// Sub-command: env sync-dockerfile
	var dockerfileDryRun bool
	syncDockerfileCmd := &cobra.Command{
		Use:   "sync-dockerfile",
		Short: "Sync environment variable documentation to Dockerfile",
		Long: `Updates the Dockerfile environment variable section with current registry.
Preserves Dockerfile structure, only updates the env vars comment block.

The Dockerfile must contain the marker comments for this to work.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := wellknown.SyncDockerfileEnvDocs("Dockerfile", dockerfileDryRun); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to sync Dockerfile: %v\n", err)
				return err
			}
			if dockerfileDryRun {
				fmt.Println("✅ Dry run complete (no changes made)")
			} else {
				fmt.Println("✅ Dockerfile environment documentation updated")
			}
			return nil
		},
	}
	syncDockerfileCmd.Flags().BoolVarP(&dockerfileDryRun, "dry-run", "n", false, "Preview changes without writing")

	// Sub-command: env sync-flytoml
	var flytomlDryRun bool
	syncFlyTomlCmd := &cobra.Command{
		Use:   "sync-flytoml",
		Short: "Sync non-secret environment variables to fly.toml",
		Long: `Updates fly.toml [env] section with non-secret environment variables.
Only includes variables where Secret=false.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := wellknown.SyncFlyTomlEnv("fly.toml", flytomlDryRun); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to sync fly.toml: %v\n", err)
				return err
			}
			if flytomlDryRun {
				fmt.Println("✅ Dry run complete (no changes made)")
			} else {
				fmt.Println("✅ fly.toml [env] section updated")
			}
			return nil
		},
	}
	syncFlyTomlCmd.Flags().BoolVarP(&flytomlDryRun, "dry-run", "n", false, "Preview changes without writing")

	// Sub-command: env generate-local
	generateLocalCmd := &cobra.Command{
		Use:   "generate-local",
		Short: "Generate .env.local template",
		Long: `Generates .env.local template with development-specific defaults.
Includes HTTPS_ENABLED=true and localhost OAuth URLs.

This will overwrite any existing .env.local file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			content := wellknown.GenerateEnvLocal()
			if err := os.WriteFile(".env.local", []byte(content), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to write .env.local: %v\n", err)
				return err
			}
			fmt.Println("✅ .env.local generated")
			fmt.Println("💡 Configure your OAuth credentials before running the server")
			return nil
		},
	}

	// Sub-command: env generate-production
	generateProductionCmd := &cobra.Command{
		Use:   "generate-production",
		Short: "Generate .env.production template",
		Long: `Generates .env.production template with production-specific defaults.
Includes HTTPS_ENABLED=false (Fly.io handles TLS) and production OAuth URLs.

This will overwrite any existing .env.production file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			content := wellknown.GenerateEnvProduction()
			if err := os.WriteFile(".env.production", []byte(content), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to write .env.production: %v\n", err)
				return err
			}
			fmt.Println("✅ .env.production generated")
			fmt.Println("💡 Configure your production OAuth credentials before deploying")
			return nil
		},
	}

	// Sub-command: env generate-example
	generateExampleCmd := &cobra.Command{
		Use:   "generate-example",
		Short: "Generate .env.example template",
		Long: `Generates .env.example template with placeholder values (safe to commit).
This file shows all available environment variables without real credentials.

This will overwrite any existing .env.example file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			content := wellknown.GenerateEnvExample()
			if err := os.WriteFile(".env.example", []byte(content), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to write .env.example: %v\n", err)
				return err
			}
			fmt.Println("✅ .env.example generated")
			fmt.Println("💡 This file is safe to commit to version control")
			return nil
		},
	}

	// Sub-command: env sync-secrets
	syncSecretsCmd := &cobra.Command{
		Use:   "sync-secrets",
		Short: "Decrypt and merge .env.secrets[.age] into .env.local",
		Long: `Merges git-tracked secrets (.env.secrets or .env.secrets.age) into .env.local for local development.

This command:
1. Reads your credentials from .env.secrets or .env.secrets.age (auto-decrypts if .age)
2. Generates .env.local template with localhost URLs
3. Merges your secrets into the template
4. Writes to .env.local (ready for local development)

Workflow:
  cp .env.secrets.example .env.secrets
  # Edit .env.secrets with real credentials
  # Optional: Encrypt with Age: age -e -r YOUR_PUBLIC_KEY .env.secrets > .env.secrets.age
  make env-sync-secrets
  make run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := wellknown.MergeSecretsIntoEnv(".env.secrets", "local", ".env.local"); err != nil {
				return err
			}
			fmt.Println("✅ .env.local generated from .env.secrets")
			fmt.Println("💡 Your local development environment is ready!")
			fmt.Println("   Run: make run")
			return nil
		},
	}

	// Sub-command: env sync-secrets-production
	syncSecretsProductionCmd := &cobra.Command{
		Use:   "sync-secrets-production",
		Short: "Decrypt and merge .env.secrets[.age] into .env.production",
		Long: `Merges git-tracked secrets (.env.secrets or .env.secrets.age) into .env.production for Fly.io deployment.

This command:
1. Reads your credentials from .env.secrets or .env.secrets.age (auto-decrypts if .age)
2. Generates .env.production template with fly.dev URLs
3. Merges your secrets into the template
4. Writes to .env.production (ready for Fly.io deployment)

Workflow:
  make env-sync-secrets-production
  make fly-secrets  # Push to Fly.io
  make fly-deploy`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := wellknown.MergeSecretsIntoEnv(".env.secrets", "production", ".env.production"); err != nil {
				return err
			}
			fmt.Println("✅ .env.production generated from .env.secrets")
			fmt.Println("💡 Ready to deploy to Fly.io!")
			fmt.Println("   Next: make fly-secrets")
			return nil
		},
	}

	envCmd.AddCommand(
		exportCmd,
		listCmd,
		validateCmd,
		syncDockerfileCmd,
		syncFlyTomlCmd,
		generateLocalCmd,
		generateProductionCmd,
		generateExampleCmd,
		syncSecretsCmd,
		syncSecretsProductionCmd,
	)
	return envCmd
}

// ---------------------------------------------------------------
// Local Update Command (for development/testing)
// ---------------------------------------------------------------

func newLocalUpdateCommand(app core.App) *cobra.Command {
	localDir := wellknown.EnvRegistry.ByName("UPDATE_LOCAL_DIR").GetString()
	archiveExec := "wellknown-pb"

	var withBackup bool

	updateCmd := &cobra.Command{
		Use:          "update",
		Short:        "Update executable from local build directory (development/testing)",
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			return updateFromLocal(app, localDir, archiveExec, withBackup)
		},
	}

	updateCmd.PersistentFlags().BoolVar(
		&withBackup,
		"backup",
		true,
		"Creates a pb_data backup at the end of the update process",
	)

	return updateCmd
}

// updateFromLocal performs the update using a local .dist directory
func updateFromLocal(app core.App, localDir string, archiveExec string, withBackup bool) error {
	color.Yellow("Updating from local directory: %s", localDir)

	// Determine platform suffix
	suffix := archiveSuffix(runtime.GOOS, runtime.GOARCH)
	if suffix == "" {
		return errors.New("unsupported platform")
	}

	// Find the appropriate archive in local directory
	archiveName := archiveExec + suffix
	archivePath := filepath.Join(localDir, archiveName)

	if _, err := os.Stat(archivePath); err != nil {
		return fmt.Errorf("local archive not found: %s (did you run 'make release'?)", archivePath)
	}

	color.Yellow("Found local archive: %s", archiveName)

	// Create temporary extraction directory
	releaseDir := filepath.Join(app.DataDir(), core.LocalTempDirName)
	defer os.RemoveAll(releaseDir)

	color.Yellow("Extracting %s...", archiveName)

	extractDir := filepath.Join(releaseDir, "extracted_"+archiveName)
	defer os.RemoveAll(extractDir)

	if err := archive.Extract(archivePath, extractDir); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	color.Yellow("Replacing the executable...")

	oldExec, err := os.Executable()
	if err != nil {
		return err
	}
	renamedOldExec := oldExec + ".old"
	defer os.Remove(renamedOldExec)

	newExec := filepath.Join(extractDir, archiveExec)
	if _, err := os.Stat(newExec); err != nil {
		// try again with an .exe extension
		newExec = newExec + ".exe"
		if _, fallbackErr := os.Stat(newExec); fallbackErr != nil {
			return fmt.Errorf("the executable in the extracted path is missing or it is inaccessible: %v, %v", err, fallbackErr)
		}
	}

	// Rename the current executable
	if err := os.Rename(oldExec, renamedOldExec); err != nil {
		return fmt.Errorf("failed to rename the current executable: %w", err)
	}

	tryToRevertExecChanges := func() {
		if revertErr := os.Rename(renamedOldExec, oldExec); revertErr != nil {
			app.Logger().Error(
				"Failed to revert executable",
				"old", renamedOldExec,
				"new", oldExec,
				"error", revertErr.Error(),
			)
		}
	}

	// Replace with the extracted binary
	if err := os.Rename(newExec, oldExec); err != nil {
		tryToRevertExecChanges()
		return fmt.Errorf("failed replacing the executable: %w", err)
	}

	if withBackup {
		color.Yellow("Creating pb_data backup...")

		backupName := "@local_update.zip"
		if err := app.CreateBackup(nil, backupName); err != nil {
			tryToRevertExecChanges()
			return err
		}
	}

	color.HiBlack("---")
	color.Green("Local update completed successfully! You can start the executable as usual.")
	color.Cyan("(Source: %s)", archivePath)

	return nil
}

// archiveSuffix returns the platform-specific archive suffix
func archiveSuffix(goos, goarch string) string {
	switch goos {
	case "linux":
		switch goarch {
		case "amd64":
			return "_linux_amd64.zip"
		case "arm64":
			return "_linux_arm64.zip"
		case "arm":
			return "_linux_armv7.zip"
		}
	case "darwin":
		switch goarch {
		case "amd64":
			return "_darwin_amd64.zip"
		case "arm64":
			return "_darwin_arm64.zip"
		}
	case "windows":
		switch goarch {
		case "amd64":
			return "_windows_amd64.zip"
		case "arm64":
			return "_windows_arm64.zip"
		}
	}

	return ""
}
