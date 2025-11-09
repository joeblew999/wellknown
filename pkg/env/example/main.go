package main

// main.go - Simplified CLI for Fly.io deployment demo
// This version includes ONLY:
//   - HTTP server (serve, health)
//   - Workflow commands (from workflow.go)
//
// For low-level commands, see commands.go (preserved as reference)

import (
	"flag"
	"fmt"
	"os"
)

const appName = "env-demo"
const appUsage = "Environment management demo with HTTP server"

func main() {
	// Parse global flags
	workDir := flag.String("dir", "", "Change to `DIR` before running command")
	flag.StringVar(workDir, "C", "", "Change to `DIR` before running command (shorthand)")

	// Custom usage function
	flag.Usage = printUsage

	flag.Parse()

	// Get command (first non-flag argument)
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(0)
	}
	command := args[0]

	// Before hook: change directory if --dir specified
	if *workDir != "" {
		if err := os.Chdir(*workDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to change directory to %s: %v\n", *workDir, err)
			os.Exit(1)
		}
	}

	// Check ENV_WORK_DIR environment variable (if flag not set)
	if *workDir == "" {
		if envDir := os.Getenv("ENV_WORK_DIR"); envDir != "" {
			if err := os.Chdir(envDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to change directory to %s: %v\n", envDir, err)
				os.Exit(1)
			}
		}
	}

	// Command routing
	switch command {
	// HTTP Server
	case "serve":
		cmdServe()
	case "health":
		cmdHealth()
	case "killport":
		cmdKillPort()

	// Workflow Automation (from workflow.go)
	case "sync-registry":
		cmdSyncRegistry()
	case "sync-environments":
		cmdSyncEnvironments()
	case "finalize":
		cmdFinalize()
	case "ko-build":
		cmdKoBuild()

	// Help
	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command: %s\n", command)
		fmt.Fprintf(os.Stderr, "Run '%s help' for usage.\n", appName)
		os.Exit(1)
	}
}

// printUsage prints the main help text
func printUsage() {
	fmt.Printf("%s - %s\n\n", appName, appUsage)
	fmt.Printf("USAGE:\n")
	fmt.Printf("  %s [global options] command\n\n", appName)

	fmt.Printf("GLOBAL OPTIONS:\n")
	fmt.Printf("  --dir DIR, -C DIR  Change to DIR before running command [$ENV_WORK_DIR]\n")
	fmt.Printf("  --help, -h         Show help\n\n")

	fmt.Printf("COMMANDS:\n\n")

	// HTTP Server
	fmt.Printf("  HTTP Server:\n")
	fmt.Printf("    serve          Start HTTP server on $SERVER_PORT (default: 8080)\n")
	fmt.Printf("    health         Perform CLI health check\n")
	fmt.Printf("    killport       Kill any process using $SERVER_PORT\n\n")

	// Workflow Automation
	fmt.Printf("  Workflow Automation:\n")
	fmt.Printf("    sync-registry      Sync deployment configs and environment templates\n")
	fmt.Printf("    sync-environments  Merge secrets into environments and validate\n")
	fmt.Printf("    finalize           Encrypt files and prepare for deployment\n")
	fmt.Printf("    ko-build           Build with ko (fast 12MB Docker image)\n\n")

	fmt.Printf("WORKFLOW:\n")
	fmt.Printf("  1. Edit registry.go to define your environment variables\n")
	fmt.Printf("  2. Run: %s sync-registry\n", appName)
	fmt.Printf("  3. Edit .env.secrets.local and .env.secrets.production with actual values\n")
	fmt.Printf("  4. Run: %s sync-environments\n", appName)
	fmt.Printf("  5. Run: %s finalize\n", appName)
	fmt.Printf("  6. Deploy to Fly.io: flyctl deploy\n\n")

	fmt.Printf("ENDPOINTS:\n")
	fmt.Printf("  Once running with 'serve', the following endpoints are available:\n")
	fmt.Printf("    GET /               Homepage\n")
	fmt.Printf("    GET /health         Health check (JSON)\n")
	fmt.Printf("    GET /env            Environment variables showcase (JSON)\n")
	fmt.Printf("    GET /feature-demo   Feature flag demonstration\n")
	fmt.Printf("    GET /database       Database connection status (JSON)\n\n")

	fmt.Printf("See WORKFLOW.md for detailed usage guide.\n")
}
