package pocketbase

import (
	"fmt"
	"log"
	"os"
	"strconv"

	wellknown "github.com/joeblew999/wellknown/pkg/pb"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterBootstrapHooks registers hooks that apply environment variable
// configuration at server startup. This is the proper PocketBase pattern
// for runtime configuration, separate from schema migrations.
//
// These hooks are idempotent - they only apply env var overrides if the
// settings haven't been customized via the Admin UI or API.
func RegisterBootstrapHooks(app core.App, cfg *wellknown.Config) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		log.Println("🔧 Running bootstrap hooks...")

		// Bootstrap settings from environment variables
		if err := bootstrapSettingsFromEnv(app); err != nil {
			log.Printf("⚠️  Bootstrap settings warning: %v", err)
		}

		// Log AI configuration status
		bootstrapAILogging(cfg)

		// Admin user creation can be done via PocketBase Admin UI
		// No need for programmatic creation - users can:
		// 1. Use the web UI at /_/ on first startup
		// 2. Use the command: ./wellknown-pb pb admin create

		log.Println("✅ Bootstrap hooks completed")
		return e.Next()
	})
}

// bootstrapSettingsFromEnv applies environment variable overrides to settings.
// This only updates settings if they haven't been customized yet.
func bootstrapSettingsFromEnv(app core.App) error {
	settings := app.Settings()
	modified := false

	// SMTP configuration (only if env vars present and not already configured)
	if smtpHost := os.Getenv("SMTP_HOST"); smtpHost != "" && !settings.SMTP.Enabled {
		log.Println("   📧 Configuring SMTP from environment variables")
		settings.SMTP.Enabled = true
		settings.SMTP.Host = smtpHost
		settings.SMTP.Port = getEnvIntOrDefault("SMTP_PORT", 587)
		settings.SMTP.Username = os.Getenv("SMTP_USERNAME")
		settings.SMTP.Password = os.Getenv("SMTP_PASSWORD")
		settings.SMTP.TLS = true
		settings.SMTP.AuthMethod = "PLAIN"

		if fromEmail := os.Getenv("SMTP_FROM_EMAIL"); fromEmail != "" {
			settings.Meta.SenderAddress = fromEmail
		}
		if fromName := os.Getenv("SMTP_FROM_NAME"); fromName != "" {
			settings.Meta.SenderName = fromName
		}

		modified = true
	}

	// S3 storage configuration (only if env vars present and not already configured)
	if s3Bucket := os.Getenv("S3_BUCKET"); s3Bucket != "" && !settings.S3.Enabled {
		log.Println("   ☁️  Configuring S3 storage from environment variables")
		settings.S3.Enabled = true
		settings.S3.Bucket = s3Bucket
		settings.S3.Region = getEnvOrDefault("S3_REGION", "us-east-1")
		settings.S3.Endpoint = os.Getenv("S3_ENDPOINT")
		settings.S3.AccessKey = os.Getenv("S3_ACCESS_KEY")
		settings.S3.Secret = os.Getenv("S3_SECRET")
		settings.S3.ForcePathStyle = getBoolEnv("S3_FORCE_PATH_STYLE", false)
		modified = true
	}

	// S3 backup storage (only if S3 enabled and backup bucket specified)
	if s3BackupBucket := os.Getenv("S3_BACKUP_BUCKET"); s3BackupBucket != "" && !settings.Backups.S3.Enabled {
		log.Println("   💾 Configuring S3 backup storage from environment variables")
		settings.Backups.S3.Enabled = true
		settings.Backups.S3.Bucket = s3BackupBucket
		settings.Backups.S3.Region = getEnvOrDefault("S3_REGION", "us-east-1")
		settings.Backups.S3.Endpoint = os.Getenv("S3_ENDPOINT")
		settings.Backups.S3.AccessKey = os.Getenv("S3_ACCESS_KEY")
		settings.Backups.S3.Secret = os.Getenv("S3_SECRET")
		settings.Backups.S3.ForcePathStyle = getBoolEnv("S3_FORCE_PATH_STYLE", false)
		modified = true
	}

	// Trusted proxy configuration (only if behind proxy and not already configured)
	if getBoolEnv("BEHIND_PROXY", false) && len(settings.TrustedProxy.Headers) == 0 {
		log.Println("   🔒 Configuring trusted proxy from environment variables")
		settings.TrustedProxy.Headers = []string{"X-Forwarded-For", "X-Real-IP"}
		settings.TrustedProxy.UseLeftmostIP = true
		modified = true
	}

	// App URL override (useful for production deployments)
	if appURL := os.Getenv("APP_URL"); appURL != "" && settings.Meta.AppURL == "http://localhost:8090" {
		log.Printf("   🌐 Setting app URL from environment: %s", appURL)
		settings.Meta.AppURL = appURL
		modified = true
	}

	// Save settings if modified
	if modified {
		if err := app.Save(settings); err != nil {
			return fmt.Errorf("failed to save settings: %w", err)
		}
		log.Println("   ✅ Settings updated from environment variables")
	}

	return nil
}

// Helper functions for environment variable parsing

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if result, err := strconv.Atoi(value); err == nil {
			return result
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

// bootstrapAILogging logs the AI configuration status for visibility
func bootstrapAILogging(cfg *wellknown.Config) {
	log.Println("   🤖 AI Configuration (Anthropic Claude):")

	if cfg.AI.Anthropic.APIKey != "" {
		log.Println("      • Mode: API Key (RECOMMENDED)")
		log.Printf("      • Model: %s", cfg.AI.Anthropic.Model)
		log.Println("      • Ready for Claude API calls")
	} else if cfg.AI.Anthropic.UseOAuth {
		log.Println("      • Mode: Anthropic OAuth (NOT IMPLEMENTED YET)")
		log.Printf("      • Model: %s", cfg.AI.Anthropic.Model)
		log.Println("      • ⚠️  OAuth mode requires 'anthropic_tokens' collection (not created)")
		log.Println("      • For now, use ANTHROPIC_API_KEY instead")
	} else {
		log.Println("      • Status: Not configured")
		log.Println("      • To enable: Set ANTHROPIC_API_KEY in .env or fly secrets")
		log.Println("      • Get your key from: https://console.anthropic.com/")
	}
}
