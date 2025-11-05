package pb_migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Migration: Configure default settings for PocketBase
// This migration sets ONLY sensible defaults - no environment variables!
// Environment variable overrides are handled by bootstrap hooks at runtime.
func init() {
	m.Register(func(app core.App) error {
		settings := app.Settings()

		// Application metadata defaults
		settings.Meta.AppName = "Wellknown OAuth"
		settings.Meta.AppURL = "http://localhost:8090"
		settings.Meta.SenderName = "Wellknown Support"
		settings.Meta.SenderAddress = "noreply@example.com"
		settings.Meta.HideControls = false

		// SMTP disabled by default (will be configured via bootstrap if env vars present)
		settings.SMTP.Enabled = false

		// Logs configuration
		settings.Logs.MaxDays = 7
		settings.Logs.LogIP = true

		// S3 storage disabled by default (will be configured via bootstrap if env vars present)
		settings.S3.Enabled = false

		// Automated backups defaults
		settings.Backups.Cron = "0 2 * * *" // 2 AM daily
		settings.Backups.CronMaxKeep = 5

		// S3 backup storage disabled by default
		settings.Backups.S3.Enabled = false

		// API rate limits - sensible defaults
		settings.RateLimits.Enabled = true
		settings.RateLimits.Rules = []core.RateLimitRule{
			{Label: "*:auth", MaxRequests: 10, Duration: 60},    // 10 auth requests per minute
			{Label: "*:create", MaxRequests: 50, Duration: 60},  // 50 creates per minute
			{Label: "*:update", MaxRequests: 100, Duration: 60}, // 100 updates per minute
		}

		// Batch request handling defaults
		settings.Batch.Enabled = true
		settings.Batch.MaxRequests = 50
		settings.Batch.Timeout = 120                 // 2 minutes
		settings.Batch.MaxBodySize = 10 * 1024 * 1024 // 10MB

		// Trusted proxy disabled by default (will be configured via bootstrap if needed)
		settings.TrustedProxy.Headers = nil
		settings.TrustedProxy.UseLeftmostIP = false

		// Save settings
		if err := app.Save(settings); err != nil {
			return err
		}

		return nil
	}, nil)
}
