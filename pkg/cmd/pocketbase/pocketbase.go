package pocketbase

// PocketBase service for wellknown
// Follows the proper PocketBase pattern.
// Reference: .src/presentator/base/main.go
//
// Key PocketBase plugins enabled:
// - jsvm: JS hooks support (pb_hooks/*.pb.js files)
// - migratecmd: Database migration commands
// - ghupdate: GitHub self-update command
//
// See: https://pocketbase.io/docs/go-overview/

import (
	"log"
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase/plugins/ghupdate"
	"github.com/pocketbase/pocketbase/plugins/jsvm"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/osutils"

	wellknown "github.com/joeblew999/wellknown/pkg/pb"
	_ "github.com/joeblew999/wellknown/pkg/cmd/pocketbase/pb_migrations" // Import migrations
)

// Main is the entry point for the PocketBase service
// args contains the command-line arguments after the service name
func Main(args []string) {
	// Load configuration
	cfg, err := wellknown.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Warn if OAuth is not configured
	if !cfg.OAuth.Google.Enabled {
		log.Println("⚠️  Google OAuth not configured")
		log.Println("   Set GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, and GOOGLE_REDIRECT_URL")
	}

	// Create wellknown PocketBase app with config
	app, err := wellknown.NewWithConfig(cfg)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}

	// ---------------------------------------------------------------
	// Plugins and hooks:
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

	// Set os.Args for PocketBase's cobra commands
	// If no command specified, default to "serve"
	if len(args) == 0 {
		os.Args = []string{os.Args[0], "serve"}
	} else {
		os.Args = append([]string{os.Args[0]}, args...)
	}

	app.RootCmd.ParseFlags(args)

	// Load jsvm plugin (supports pb_hooks/*.pb.js)
	jsvm.MustRegister(app, jsvm.Config{
		MigrationsDir: migrationsDir,
		HooksDir:      hooksDir,
		HooksWatch:    hooksWatch,
		HooksPoolSize: hooksPool,
	})

	// Migrate command plugin
	migrateConfig := migratecmd.Config{Dir: migrationsDir}
	if !osutils.IsProbablyGoRun() {
		migrateConfig.TemplateLang = migratecmd.TemplateLangJS
	}
	migratecmd.MustRegister(app, app.RootCmd, migrateConfig)

	// GitHub selfupdate plugin
	ghupdate.MustRegister(app, app.RootCmd, ghupdate.Config{
		Owner:             "joeblew999",
		Repo:              "wellknown",
		ArchiveExecutable: "wellknown-pb",
	})

	// Register bootstrap hooks for environment variable configuration
	// This applies env vars at runtime (proper PocketBase pattern)
	RegisterBootstrapHooks(app, cfg)

	log.Println("🚀 Starting Wellknown PocketBase server...")
	log.Printf("📍 Server will be available at: %s", cfg.Server.ServerURL())
	log.Println("")

	// Start the app (this is blocking)
	// Route logging happens in the OnServe hook in wellknown package
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
