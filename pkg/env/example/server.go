package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joeblew999/wellknown/pkg/env"
	"github.com/joeblew999/wellknown/pkg/env/webui"
)

// Server runs the HTTP server demonstrating environment variable usage
func runServer() error {
	// Get port from environment (uses registry default if not set)
	port := getRegistryDefault("SERVER_PORT")

	// Get log level (uses registry default if not set)
	logLevel := getRegistryDefault("LOG_LEVEL")

	// Setup routes
	mux := http.NewServeMux()

	// Register webui routes for env management
	webuiHandler := webui.NewHandler(AppRegistry)
	webuiHandler.RegisterRoutes(mux)

	// App-specific routes
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/feature-demo", handleFeatureDemo)
	mux.HandleFunc("/database", handleDatabase)

	// Create server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		baseURL := fmt.Sprintf("http://localhost:%s", port)
		log.Printf("🚀 Server starting at %s (log level: %s)\n", baseURL, logLevel)
		log.Printf("📍 Endpoints:\n")
		log.Printf("   GET %s/          - Homepage\n", baseURL)
		log.Printf("   GET %s/env       - Environment variables (webui)\n", baseURL)
		log.Printf("   GET %s/health    - Health check (webui)\n", baseURL)
		log.Printf("   GET %s/feature-demo - Feature flag demo\n", baseURL)
		log.Printf("   GET %s/database  - Database status\n", baseURL)
		log.Println()

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed: %v\n", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\n🛑 Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("✅ Server stopped")
	return nil
}

// handleHome shows the application homepage
func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	port := getRegistryDefault("SERVER_PORT")
	logLevel := getRegistryDefault("LOG_LEVEL")
	featureBeta := getRegistryDefault("FEATURE_BETA")
	environment := env.DetectEnvironment()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Environment Management Demo</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            max-width: 800px;
            margin: 40px auto;
            padding: 20px;
            line-height: 1.6;
        }
        .status { background: #e8f5e9; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .config { background: #f5f5f5; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .secret { color: #999; font-style: italic; }
        h1 { color: #2c3e50; }
        h2 { color: #34495e; border-bottom: 2px solid #3498db; padding-bottom: 5px; }
        code { background: #f4f4f4; padding: 2px 6px; border-radius: 3px; }
        a { color: #3498db; text-decoration: none; }
        a:hover { text-decoration: underline; }
        .endpoints { margin: 20px 0; }
        .endpoints li { margin: 10px 0; }
    </style>
</head>
<body>
    <h1>🚀 Environment Management Demo</h1>

    <div class="status">
        <strong>Status:</strong> Running<br>
        <strong>Environment:</strong> %s<br>
        <strong>Port:</strong> %s<br>
        <strong>Log Level:</strong> %s<br>
        <strong>Feature Beta:</strong> %s
    </div>

    <h2>About This Demo</h2>
    <p>
        This is a demonstration application for the
        <code>github.com/joeblew999/wellknown/pkg/env</code> package.
        It shows how to manage environment variables across different environments
        (local, production) with built-in secrets management and encryption.
    </p>

    <h2>Available Endpoints</h2>
    <div class="endpoints">
        <ul>
            <li><a href="/">/</a> - This homepage</li>
            <li><a href="/env">/env</a> - Environment variables GUI (webui package)</li>
            <li><a href="/health">/health</a> - Health check (JSON, webui package)</li>
            <li><a href="/feature-demo">/feature-demo</a> - Feature flag demonstration</li>
            <li><a href="/database">/database</a> - Database connection status (JSON)</li>
        </ul>
    </div>

    <h2>Features</h2>
    <ul>
        <li>✅ Registry-driven environment management</li>
        <li>✅ Separate local and production configurations</li>
        <li>✅ Built-in secrets encryption (Age)</li>
        <li>✅ Template-based env file generation</li>
        <li>✅ Workflow automation for deployment</li>
        <li>✅ Fly.io deployment integration</li>
    </ul>

    <h2>Workflow Commands</h2>
    <div class="config">
        <code>sync-registry</code> - Sync deployment configs and templates<br>
        <code>sync-environments</code> - Merge secrets into environments<br>
        <code>finalize</code> - Encrypt files and prepare for deployment
    </div>

    <p style="margin-top: 40px; color: #999; font-size: 0.9em;">
        Powered by <strong>github.com/joeblew999/wellknown/pkg/env</strong>
    </p>
</body>
</html>`, environment, port, logLevel, featureBeta)
}

// handleFeatureDemo demonstrates feature flag usage
func handleFeatureDemo(w http.ResponseWriter, r *http.Request) {
	featureBeta := getRegistryDefault("FEATURE_BETA") == "true"

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if featureBeta {
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Beta Features</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            max-width: 800px;
            margin: 40px auto;
            padding: 20px;
            text-align: center;
        }
        .beta { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
                color: white; padding: 40px; border-radius: 10px; margin-bottom: 20px; }
        h1 { font-size: 3em; margin: 0; }
        p { font-size: 1.2em; }
        a { display: inline-block; margin-top: 20px; padding: 10px 20px;
            background: #667eea; color: white; text-decoration: none;
            border-radius: 5px; }
        a:hover { background: #764ba2; }
    </style>
</head>
<body>
    <div class="beta">
        <h1>🎉 Beta Features Enabled!</h1>
        <p>You're seeing this because FEATURE_BETA=true</p>
        <p>Welcome to the cutting edge! 🚀</p>
    </div>
    <a href="/">← Back to Home</a>
</body>
</html>`)
	} else {
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Standard Features</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            max-width: 800px;
            margin: 40px auto;
            padding: 20px;
            text-align: center;
        }
        .standard { background: #f5f5f5; padding: 40px; border-radius: 10px; margin-bottom: 20px; }
        h1 { font-size: 2.5em; margin: 0; color: #333; }
        p { font-size: 1.1em; color: #666; }
        a { display: inline-block; margin-top: 20px; padding: 10px 20px;
            background: #3498db; color: white; text-decoration: none;
            border-radius: 5px; }
        a:hover { background: #2980b9; }
    </style>
</head>
<body>
    <div class="standard">
        <h1>📦 Standard Version</h1>
        <p>You're seeing this because FEATURE_BETA=false</p>
        <p>Set FEATURE_BETA=true to enable beta features</p>
    </div>
    <a href="/">← Back to Home</a>
</body>
</html>`)
	}
}

// handleDatabase shows database connection status
func handleDatabase(w http.ResponseWriter, r *http.Request) {
	databaseURL := os.Getenv("DATABASE_URL")

	response := map[string]interface{}{
		"configured": databaseURL != "",
		"status":     "not_connected",
	}

	if databaseURL != "" {
		// Validate URL format (don't show actual value)
		if strings.HasPrefix(databaseURL, "postgresql://") ||
			strings.HasPrefix(databaseURL, "postgres://") ||
			strings.HasPrefix(databaseURL, "mysql://") ||
			strings.HasPrefix(databaseURL, "sqlite://") {
			response["status"] = "url_valid"
			response["protocol"] = strings.Split(databaseURL, "://")[0]
		} else {
			response["status"] = "url_invalid"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper functions

// getRegistryDefault gets environment variable with fallback to registry default
func getRegistryDefault(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	// Get default from registry
	if envVar := AppRegistry.ByName(key); envVar != nil {
		return envVar.Default
	}
	return ""
}

// cmdServe starts the HTTP server
func cmdServe() {
	// Load environment variables from .env files if they exist
	if env.Local.Exists() {
		if err := loadEnvFile(env.Local.FullPath()); err != nil {
			log.Printf("⚠️  Could not load %s: %v\n", env.Local.FileName, err)
		}
	} else if env.Production.Exists() {
		if err := loadEnvFile(env.Production.FullPath()); err != nil {
			log.Printf("⚠️  Could not load %s: %v\n", env.Production.FileName, err)
		}
	}

	if err := runServer(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Server error: %v\n", err)
		os.Exit(1)
	}
}

// cmdHealth performs a health check (CLI version)
func cmdHealth() {
	port := getRegistryDefault("SERVER_PORT")
	url := fmt.Sprintf("http://localhost:%s/health", port)

	fmt.Printf("🔍 Checking %s...\n", url)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Health check failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "❌ Health check failed: HTTP %d\n", resp.StatusCode)
		os.Exit(1)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Invalid health response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Health check passed")
	fmt.Printf("   Status: %v\n", result["status"])
	fmt.Printf("   Environment: %v\n", result["environment"])
}

// cmdKillPort kills any process using the configured SERVER_PORT
func cmdKillPort() {
	port := getRegistryDefault("SERVER_PORT")

	fmt.Printf("🔍 Checking port %s...\n", port)

	// Find process using lsof
	findCmd := fmt.Sprintf("lsof -ti:%s", port)
	out, err := exec.Command("sh", "-c", findCmd).Output()

	if err != nil || len(out) == 0 {
		fmt.Printf("✅ No process found using port %s\n", port)
		return
	}

	pid := strings.TrimSpace(string(out))
	fmt.Printf("✅ Found process using port %s: PID %s\n", port, pid)
	fmt.Printf("🔪 Killing process %s...\n", pid)

	// Kill the process
	killCmd := fmt.Sprintf("kill -9 %s", pid)
	exec.Command("sh", "-c", killCmd).Run()

	fmt.Printf("✅ Port %s is now free\n", port)
}

// loadEnvFile loads environment variables from a file
func loadEnvFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			value = strings.Trim(value, `"'`)
			os.Setenv(key, value)
		}
	}

	return nil
}
