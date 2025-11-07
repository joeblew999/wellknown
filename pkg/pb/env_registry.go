package wellknown

import (
	"github.com/joeblew999/wellknown/pkg/env"
)

// AllEnvVars is the single source of truth for all environment variables in this PocketBase application.
// This registry is used for:
// - Auto-generating .env.example
// - Validating configuration
// - Exporting secrets to Fly.io
// - Self-documenting the application
var AllEnvVars = []env.EnvVar{
	// ================================================================
	// Server Configuration (non-secret, in fly.toml [env] section)
	// ================================================================
	{
		Name:        "SERVER_HOST",
		Description: "Server bind address (currently used for logging only - actual bind controlled by --http/--https flags)",
		Default:     "127.0.0.1",
		Group:       "Server",
	},
	{
		Name:        "SERVER_PORT",
		Description: "Server port",
		Default:     "8090",
		Group:       "Server",
	},
	{
		Name:        "PB_DATA_DIR",
		Description: "PocketBase data directory",
		Default:     ".data/pb",
		Group:       "Server",
	},

	// ================================================================
	// Google OAuth (REQUIRED secrets)
	// ================================================================
	{
		Name:        "GOOGLE_CLIENT_ID",
		Description: "Google OAuth client ID",
		Required:    true,
		Secret:      true,
		Group:       "Google OAuth",
	},
	{
		Name:        "GOOGLE_CLIENT_SECRET",
		Description: "Google OAuth client secret",
		Required:    true,
		Secret:      true,
		Group:       "Google OAuth",
	},
	{
		Name:        "GOOGLE_REDIRECT_URL",
		Description: "Google OAuth callback URL",
		Required:    true,
		Secret:      true,
		Group:       "Google OAuth",
	},

	// ================================================================
	// Apple OAuth (OPTIONAL secrets)
	// ================================================================
	{
		Name:        "APPLE_TEAM_ID",
		Description: "Apple Developer Team ID",
		Secret:      true,
		Group:       "Apple OAuth",
	},
	{
		Name:        "APPLE_CLIENT_ID",
		Description: "Apple OAuth client ID (Service ID)",
		Secret:      true,
		Group:       "Apple OAuth",
	},
	{
		Name:        "APPLE_KEY_ID",
		Description: "Apple private key ID",
		Secret:      true,
		Group:       "Apple OAuth",
	},
	{
		Name:        "APPLE_PRIVATE_KEY",
		Description: "Apple private key (inline PEM content)",
		Secret:      true,
		Group:       "Apple OAuth",
	},
	{
		Name:        "APPLE_PRIVATE_KEY_PATH",
		Description: "Path to Apple private key file (alternative to inline)",
		Secret:      true,
		Group:       "Apple OAuth",
	},
	{
		Name:        "APPLE_REDIRECT_URL",
		Description: "Apple OAuth callback URL",
		Secret:      true,
		Group:       "Apple OAuth",
	},

	// ================================================================
	// AI Configuration (Anthropic Claude API)
	// ================================================================
	{
		Name:        "ANTHROPIC_API_KEY",
		Description: "Anthropic Claude API key (get from https://console.anthropic.com/)",
		Secret:      true,
		Group:       "AI",
	},
	{
		Name:        "ANTHROPIC_MODEL",
		Description: "Claude model name",
		Default:     "claude-sonnet-4-5-20250929",
		Secret:      true,
		Group:       "AI",
	},
	{
		Name:        "AI_USE_OAUTH",
		Description: "Use Anthropic OAuth (NOT IMPLEMENTED YET, keep false)",
		Default:     "false",
		Secret:      true,
		Group:       "AI",
	},

	// ================================================================
	// PocketBase Admin (OPTIONAL secrets)
	// ================================================================
	{
		Name:        "PB_ADMIN_EMAIL",
		Description: "PocketBase admin email",
		Secret:      true,
		Group:       "PocketBase Admin",
	},
	{
		Name:        "PB_ADMIN_PASSWORD",
		Description: "PocketBase admin password",
		Secret:      true,
		Group:       "PocketBase Admin",
	},

	// ================================================================
	// SMTP Configuration (OPTIONAL secrets)
	// ================================================================
	{
		Name:        "SMTP_HOST",
		Description: "SMTP server hostname",
		Secret:      true,
		Group:       "SMTP",
	},
	{
		Name:        "SMTP_PORT",
		Description: "SMTP server port",
		Default:     "587",
		Secret:      true,
		Group:       "SMTP",
	},
	{
		Name:        "SMTP_USERNAME",
		Description: "SMTP username",
		Secret:      true,
		Group:       "SMTP",
	},
	{
		Name:        "SMTP_PASSWORD",
		Description: "SMTP password",
		Secret:      true,
		Group:       "SMTP",
	},
	{
		Name:        "SMTP_FROM_EMAIL",
		Description: "From email address",
		Secret:      true,
		Group:       "SMTP",
	},
	{
		Name:        "SMTP_FROM_NAME",
		Description: "From name",
		Secret:      true,
		Group:       "SMTP",
	},

	// ================================================================
	// S3 Storage (OPTIONAL secrets)
	// ================================================================
	{
		Name:        "S3_ENDPOINT",
		Description: "S3 endpoint URL",
		Secret:      true,
		Group:       "S3",
	},
	{
		Name:        "S3_REGION",
		Description: "S3 region",
		Default:     "us-east-1",
		Secret:      true,
		Group:       "S3",
	},
	{
		Name:        "S3_BUCKET",
		Description: "S3 bucket name",
		Secret:      true,
		Group:       "S3",
	},
	{
		Name:        "S3_ACCESS_KEY",
		Description: "S3 access key",
		Secret:      true,
		Group:       "S3",
	},
	{
		Name:        "S3_SECRET_KEY",
		Description: "S3 secret key",
		Secret:      true,
		Group:       "S3",
	},
	{
		Name:        "S3_FORCE_PATH_STYLE",
		Description: "S3 path style (true/false)",
		Secret:      true,
		Group:       "S3",
	},
	{
		Name:        "S3_BACKUP_BUCKET",
		Description: "S3 bucket for PocketBase backups",
		Secret:      true,
		Group:       "S3",
	},

	// ================================================================
	// Deployment Configuration (OPTIONAL)
	// ================================================================
	{
		Name:        "APP_URL",
		Description: "Application URL (for production deployments)",
		Group:       "Deployment",
	},
	{
		Name:        "BEHIND_PROXY",
		Description: "Set to true if behind a reverse proxy",
		Default:     "false",
		Group:       "Deployment",
	},

	// ================================================================
	// HTTPS/TLS Configuration (Development only - DO NOT use in production)
	// ================================================================
	{
		Name:        "HTTPS_ENABLED",
		Description: "Enable HTTPS with custom certificates (development only, use mkcert)",
		Default:     "false",
		Group:       "HTTPS (Development)",
	},
	{
		Name:        "CERT_FILE",
		Description: "Path to SSL certificate file",
		Default:     ".data/certs/cert.pem",
		Group:       "HTTPS (Development)",
	},
	{
		Name:        "KEY_FILE",
		Description: "Path to SSL private key file",
		Default:     ".data/certs/key.pem",
		Group:       "HTTPS (Development)",
	},
	{
		Name:        "HTTPS_PORT",
		Description: "HTTPS port (development only)",
		Default:     "8443",
		Group:       "HTTPS (Development)",
	},

	// ================================================================
	// Binary Update Configuration (Development/Testing)
	// ================================================================
	{
		Name:        "UPDATE_SOURCE",
		Description: "Update source: 'github' for production or 'local' for testing",
		Default:     "github",
		Group:       "Binary Update",
	},
	{
		Name:        "UPDATE_LOCAL_DIR",
		Description: "Local directory for binary updates when UPDATE_SOURCE=local",
		Default:     ".dist",
		Group:       "Binary Update",
	},
}

// EnvRegistry is the global environment variable registry for this application.
// It provides O(1) lookups and filtering capabilities.
var EnvRegistry = env.NewRegistry(AllEnvVars)
