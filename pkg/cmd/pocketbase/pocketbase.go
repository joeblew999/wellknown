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
	// Check environment variables (warn if missing, don't fail)
	requiredEnvVars := []string{
		"GOOGLE_CLIENT_ID",
		"GOOGLE_CLIENT_SECRET",
		"GOOGLE_REDIRECT_URL",
	}

	missingVars := []string{}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			missingVars = append(missingVars, envVar)
		}
	}

	if len(missingVars) > 0 {
		log.Printf("⚠️  Warning: Missing environment variables: %v", missingVars)
		log.Println("   Google OAuth will not work until these are set.")
		log.Println("   Copy pb/base/.env.example to pb/base/.env and configure.")
	}

	// Create wellknown PocketBase app
	app := wellknown.New()

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

	log.Println("🚀 Starting Wellknown PocketBase server...")
	log.Println("")
	log.Println("📍 Available endpoints:")
	log.Println("   Root (HTML):     http://localhost:8090/")
	log.Println("   API Index (JSON): http://localhost:8090/api/")
	log.Println("   Admin UI:        http://localhost:8090/_/")
	log.Println("")
	log.Println("🔐 OAuth:")
	log.Println("   Google Login:    http://localhost:8090/auth/google")
	log.Println("   OAuth Status:    http://localhost:8090/auth/status")
	log.Println("   Logout:          http://localhost:8090/auth/logout")
	log.Println("")
	log.Println("📅 Calendar API (authenticated):")
	log.Println("   List Events:     GET  http://localhost:8090/api/calendar/events")
	log.Println("   Create Event:    POST http://localhost:8090/api/calendar/events")
	log.Println("")
	log.Println("🏦 Banking API (example):")
	log.Println("   List Accounts:   GET  http://localhost:8090/api/banking/accounts?user_id=<id>")
	log.Println("   Get Account:     GET  http://localhost:8090/api/banking/accounts/:id")
	log.Println("   Transactions:    GET  http://localhost:8090/api/banking/accounts/:id/transactions")
	log.Println("   Create Account:  POST http://localhost:8090/api/banking/accounts")
	log.Println("   Create TX:       POST http://localhost:8090/api/banking/transactions")

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
