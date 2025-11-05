package pb_migrations

import (
	"fmt"
	"os"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Configure application metadata
		settings := app.Settings()
		settings.Meta.AppName = "Wellknown OAuth"
		settings.Meta.AppURL = getEnvOrDefault("APP_URL", "http://localhost:8090")
		settings.Meta.SenderName = "Wellknown Support"
		settings.Meta.SenderAddress = getEnvOrDefault("SENDER_EMAIL", "noreply@example.com")
		settings.Meta.HideControls = false

		// Configure SMTP (only if credentials provided)
		if os.Getenv("SMTP_HOST") != "" {
			settings.SMTP.Enabled = true
			settings.SMTP.Host = os.Getenv("SMTP_HOST")
			settings.SMTP.Port = getEnvIntOrDefault("SMTP_PORT", 587)
			settings.SMTP.Username = os.Getenv("SMTP_USERNAME")
			settings.SMTP.Password = os.Getenv("SMTP_PASSWORD")
			settings.SMTP.TLS = true
			settings.SMTP.AuthMethod = "PLAIN"
		}

		// Configure logs
		settings.Logs.MaxDays = 7
		settings.Logs.LogIP = true

		// Configure S3 storage (optional, only if enabled)
		if getBoolEnv("S3_ENABLED", false) {
			settings.S3.Enabled = true
			settings.S3.Bucket = os.Getenv("S3_BUCKET")
			settings.S3.Region = os.Getenv("S3_REGION")
			settings.S3.Endpoint = os.Getenv("S3_ENDPOINT")
			settings.S3.AccessKey = os.Getenv("S3_ACCESS_KEY")
			settings.S3.Secret = os.Getenv("S3_SECRET")
			settings.S3.ForcePathStyle = getBoolEnv("S3_FORCE_PATH_STYLE", false)
		}

		// Configure automated backups
		settings.Backups.Cron = getEnvOrDefault("BACKUP_CRON", "0 2 * * *") // 2 AM daily by default
		settings.Backups.CronMaxKeep = getEnvIntOrDefault("BACKUP_MAX_KEEP", 5)

		// S3 backup storage (if S3 enabled and backup bucket specified)
		if getBoolEnv("S3_ENABLED", false) && os.Getenv("S3_BACKUP_BUCKET") != "" {
			settings.Backups.S3.Enabled = true
			settings.Backups.S3.Bucket = os.Getenv("S3_BACKUP_BUCKET")
			settings.Backups.S3.Region = os.Getenv("S3_REGION")
			settings.Backups.S3.Endpoint = os.Getenv("S3_ENDPOINT")
			settings.Backups.S3.AccessKey = os.Getenv("S3_ACCESS_KEY")
			settings.Backups.S3.Secret = os.Getenv("S3_SECRET")
			settings.Backups.S3.ForcePathStyle = getBoolEnv("S3_FORCE_PATH_STYLE", false)
		}

		// Configure API rate limits
		settings.RateLimits.Enabled = getBoolEnv("RATE_LIMIT_ENABLED", true)
		settings.RateLimits.Rules = []core.RateLimitRule{
			{Label: "*:auth", MaxRequests: 10, Duration: 60},      // 10 auth requests per minute
			{Label: "*:create", MaxRequests: 50, Duration: 60},    // 50 creates per minute
			{Label: "*:update", MaxRequests: 100, Duration: 60},   // 100 updates per minute
			{Label: "/api/*", MaxRequests: 200, Duration: 60},     // 200 general API calls per minute
		}

		// Configure batch request handling
		settings.Batch.Enabled = getBoolEnv("BATCH_ENABLED", true)
		settings.Batch.MaxRequests = getEnvIntOrDefault("BATCH_MAX_REQUESTS", 50)
		settings.Batch.Timeout = getEnvInt64OrDefault("BATCH_TIMEOUT", 120)                       // 2 minutes
		settings.Batch.MaxBodySize = getEnvInt64OrDefault("BATCH_MAX_BODY_SIZE", 10*1024*1024) // 10MB

		// Configure trusted proxy (for reverse proxy setups)
		if getBoolEnv("BEHIND_PROXY", false) {
			settings.TrustedProxy.Headers = []string{"X-Forwarded-For", "X-Real-IP"}
			settings.TrustedProxy.UseLeftmostIP = true
		}

		// Save settings
		if err := app.Save(settings); err != nil {
			return err
		}

		return nil
	}, nil)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		// Simple conversion - in production you'd want proper error handling
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

func getEnvInt64OrDefault(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		var result int64
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
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
