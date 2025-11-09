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
)

// Server runs the HTTP server demonstrating environment variable usage
func runServer() error {
	// Get port from environment (uses registry default if not set)
	port := getRegistryDefault("SERVER_PORT")

	// Get log level (uses registry default if not set)
	logLevel := getRegistryDefault("LOG_LEVEL")

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/env", handleEnv)
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
		log.Printf("   GET %s/health    - Health check\n", baseURL)
		log.Printf("   GET %s/env       - Environment variables\n", baseURL)
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
	environment := detectEnvironment()

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
            <li><a href="/health">/health</a> - Health check endpoint (JSON)</li>
            <li><a href="/env">/env</a> - Environment variables showcase (JSON)</li>
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

// handleHealth returns health check status
func handleHealth(w http.ResponseWriter, r *http.Request) {
	environment := detectEnvironment()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "healthy",
		"environment": environment,
		"timestamp":   time.Now().Format(time.RFC3339),
	})
}

// handleEnv shows environment variables (hides secret values)
func handleEnv(w http.ResponseWriter, r *http.Request) {
	// Get all variables from registry
	vars := AppRegistry.All()

	// Build response
	response := map[string]interface{}{
		"server": map[string]string{
			"port":      getRegistryDefault("SERVER_PORT"),
			"log_level": getRegistryDefault("LOG_LEVEL"),
		},
		"features": map[string]interface{}{
			"beta": getRegistryDefault("FEATURE_BETA") == "true",
		},
		"secrets": map[string]bool{},
	}

	// Check which secrets are configured
	secrets := response["secrets"].(map[string]bool)
	for _, v := range vars {
		if v.Secret {
			key := strings.ToLower(v.Name)
			secrets[key+"_configured"] = os.Getenv(v.Name) != ""
		}
	}

	// Check if client wants JSON (via Accept header or query param)
	acceptHeader := r.Header.Get("Accept")
	format := r.URL.Query().Get("format")

	if format == "json" || strings.Contains(acceptHeader, "application/json") {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Default: HTML output
	port := getRegistryDefault("SERVER_PORT")
	logLevel := getRegistryDefault("LOG_LEVEL")
	featureBeta := getRegistryDefault("FEATURE_BETA")
	environment := detectEnvironment()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Environment Variables - env-demo</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            max-width: 1000px;
            margin: 40px auto;
            padding: 20px;
            line-height: 1.6;
        }
        h1 { color: #2c3e50; }
        h2 { color: #34495e; border-bottom: 2px solid #3498db; padding-bottom: 5px; margin-top: 30px; }
        .env-section { background: #f8f9fa; padding: 20px; border-radius: 8px; margin: 20px 0; }
        .env-group { margin: 20px 0; }
        .env-group h3 { color: #555; margin-bottom: 10px; font-size: 1.1em; }
        .env-var {
            display: flex;
            padding: 10px;
            margin: 5px 0;
            background: white;
            border-radius: 5px;
            border-left: 4px solid #3498db;
        }
        .env-var.secret { border-left-color: #e74c3c; }
        .env-var.required { border-left-color: #f39c12; }
        .env-name {
            font-weight: bold;
            font-family: monospace;
            min-width: 200px;
            color: #2c3e50;
        }
        .env-value {
            flex: 1;
            font-family: monospace;
            color: #27ae60;
        }
        .env-value.secret {
            color: #999;
            font-style: italic;
        }
        .env-value.empty {
            color: #e74c3c;
        }
        .badge {
            display: inline-block;
            padding: 2px 8px;
            border-radius: 3px;
            font-size: 0.8em;
            margin-left: 10px;
            font-weight: normal;
        }
        .badge.secret { background: #ffe6e6; color: #c0392b; }
        .badge.required { background: #fff3cd; color: #856404; }
        .badge.configured { background: #d4edda; color: #155724; }
        .badge.missing { background: #f8d7da; color: #721c24; }
        .nav { margin: 20px 0; }
        .nav a {
            display: inline-block;
            padding: 8px 16px;
            background: #3498db;
            color: white;
            text-decoration: none;
            border-radius: 5px;
            margin-right: 10px;
        }
        .nav a:hover { background: #2980b9; }
        code {
            background: #f4f4f4;
            padding: 2px 6px;
            border-radius: 3px;
            font-size: 0.9em;
        }
        .status-box {
            background: #e8f5e9;
            padding: 15px;
            border-radius: 5px;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <h1>🔧 Environment Variables</h1>

    <div class="nav">
        <a href="/">← Home</a>
        <a href="/env?format=json">View as JSON</a>
    </div>

    <div class="status-box">
        <strong>Environment:</strong> %s<br>
        <strong>Registry Variables:</strong> %d total (%d secrets, %d required)
    </div>

    <h2>Configuration Values</h2>

    <div class="env-section">
        <div class="env-group">
            <h3>Server Configuration</h3>
            <div class="env-var">
                <span class="env-name">SERVER_PORT</span>
                <span class="env-value">%s</span>
            </div>
            <div class="env-var">
                <span class="env-name">LOG_LEVEL</span>
                <span class="env-value">%s</span>
            </div>
        </div>

        <div class="env-group">
            <h3>Features</h3>
            <div class="env-var">
                <span class="env-name">FEATURE_BETA</span>
                <span class="env-value">%s</span>
            </div>
        </div>

        <div class="env-group">
            <h3>Secrets Status</h3>`,
		environment,
		len(vars),
		len(AppRegistry.GetSecrets()),
		len(AppRegistry.GetRequired()),
		port,
		logLevel,
		featureBeta)

	// Add secret status for each secret variable
	for _, v := range vars {
		if v.Secret {
			configured := os.Getenv(v.Name) != ""
			requiredBadge := ""
			if v.Required {
				requiredBadge = `<span class="badge required">REQUIRED</span>`
			}

			statusBadge := `<span class="badge missing">NOT SET</span>`
			valueDisplay := `<span class="env-value empty">(not configured)</span>`
			if configured {
				statusBadge = `<span class="badge configured">CONFIGURED</span>`
				valueDisplay = `<span class="env-value secret">••••••••</span>`
			}

			fmt.Fprintf(w, `
            <div class="env-var secret">
                <span class="env-name">%s<span class="badge secret">SECRET</span>%s</span>
                %s
                %s
            </div>`, v.Name, requiredBadge, valueDisplay, statusBadge)
		}
	}

	fmt.Fprintf(w, `
        </div>
    </div>

    <h2>All Registry Variables</h2>
    <div class="env-section">`)

	// Group variables by group
	groups := make(map[string][]env.EnvVar)
	for _, v := range vars {
		group := v.Group
		if group == "" {
			group = "Other"
		}
		groups[group] = append(groups[group], v)
	}

	// Display each group
	for groupName, groupVars := range groups {
		fmt.Fprintf(w, `
        <div class="env-group">
            <h3>%s</h3>`, groupName)

		for _, v := range groupVars {
			value := os.Getenv(v.Name)
			if value == "" {
				value = v.Default
			}

			badges := ""
			varClass := "env-var"
			valueClass := "env-value"

			if v.Secret {
				badges += `<span class="badge secret">SECRET</span>`
				varClass += " secret"
				if value != "" {
					value = "••••••••"
					valueClass += " secret"
				} else {
					valueClass += " empty"
					value = "(not set)"
				}
			} else if value == "" {
				valueClass += " empty"
				value = "(empty)"
			}

			if v.Required {
				badges += `<span class="badge required">REQUIRED</span>`
				varClass += " required"
			}

			fmt.Fprintf(w, `
            <div class="%s">
                <span class="env-name">%s%s</span>
                <span class="%s">%s</span>
            </div>`, varClass, v.Name, badges, valueClass, value)
		}

		fmt.Fprintf(w, `
        </div>`)
	}

	fmt.Fprintf(w, `
    </div>

    <h2>How to Use</h2>
    <div class="env-section">
        <p>
            <strong>View as JSON:</strong> Add <code>?format=json</code> to the URL or set
            <code>Accept: application/json</code> header.
        </p>
        <p>
            <strong>Registry-driven:</strong> All values come from <code>registry.go</code>
            with environment variable overrides.
        </p>
        <p>
            <strong>Secrets:</strong> Secret values are hidden in HTML view but their
            configured status is shown.
        </p>
    </div>

    <p style="margin-top: 40px; color: #999; font-size: 0.9em;">
        Powered by <strong>github.com/joeblew999/wellknown/pkg/env</strong>
    </p>
</body>
</html>`)
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
                color: white; padding: 40px; border-radius: 10px; }
        h1 { font-size: 3em; margin: 0; }
        p { font-size: 1.2em; }
    </style>
</head>
<body>
    <div class="beta">
        <h1>🎉 Beta Features Enabled!</h1>
        <p>You're seeing this because FEATURE_BETA=true</p>
        <p>Welcome to the cutting edge! 🚀</p>
    </div>
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
        .standard { background: #f5f5f5; padding: 40px; border-radius: 10px; }
        h1 { font-size: 2.5em; margin: 0; color: #333; }
        p { font-size: 1.1em; color: #666; }
    </style>
</head>
<body>
    <div class="standard">
        <h1>📦 Standard Version</h1>
        <p>You're seeing this because FEATURE_BETA=false</p>
        <p>Set FEATURE_BETA=true to enable beta features</p>
    </div>
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

func detectEnvironment() string {
	// Check if we're running on Fly.io
	if os.Getenv("FLY_APP_NAME") != "" {
		return "production (Fly.io)"
	}

	// Check if we're in a container
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return "docker"
	}

	// Check if DATABASE_URL looks like production
	if dbURL := os.Getenv("DATABASE_URL"); strings.Contains(dbURL, "production") ||
		strings.Contains(dbURL, "prod") {
		return "production"
	}

	return "local"
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
