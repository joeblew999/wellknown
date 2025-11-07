package main

import "github.com/joeblew999/wellknown/pkg/env"

// AppEnvVars defines the environment variables for our sample application.
// This demonstrates how to use the env package in a real application.
var AppEnvVars = []env.EnvVar{
	// ================================================================
	// Server Configuration
	// ================================================================
	{
		Name:        "SERVER_HOST",
		Description: "Server bind address",
		Default:     "0.0.0.0",
		Group:       "Server",
	},
	{
		Name:        "SERVER_PORT",
		Description: "Server port",
		Default:     "8080",
		Group:       "Server",
	},
	{
		Name:        "SERVER_TIMEOUT",
		Description: "Server request timeout in seconds",
		Default:     "30",
		Group:       "Server",
	},
	{
		Name:        "LOG_LEVEL",
		Description: "Logging level (debug, info, warn, error)",
		Default:     "info",
		Group:       "Server",
	},

	// ================================================================
	// Database Configuration (with secrets)
	// ================================================================
	{
		Name:        "DATABASE_URL",
		Description: "Database connection URL",
		Required:    true,
		Secret:      true,
		Group:       "Database",
	},
	{
		Name:        "DATABASE_MAX_CONNECTIONS",
		Description: "Maximum database connections",
		Default:     "25",
		Group:       "Database",
	},
	{
		Name:        "DATABASE_SSL_MODE",
		Description: "Database SSL mode (disable, require, verify-full)",
		Default:     "require",
		Group:       "Database",
	},

	// ================================================================
	// External APIs (secrets)
	// ================================================================
	{
		Name:        "STRIPE_API_KEY",
		Description: "Stripe API secret key",
		Required:    true,
		Secret:      true,
		Group:       "External APIs",
	},
	{
		Name:        "STRIPE_WEBHOOK_SECRET",
		Description: "Stripe webhook signing secret",
		Required:    true,
		Secret:      true,
		Group:       "External APIs",
	},
	{
		Name:        "SENDGRID_API_KEY",
		Description: "SendGrid API key for email",
		Secret:      true,
		Group:       "External APIs",
	},

	// ================================================================
	// OAuth Configuration (secrets)
	// ================================================================
	{
		Name:        "OAUTH_GOOGLE_CLIENT_ID",
		Description: "Google OAuth client ID",
		Required:    true,
		Secret:      true,
		Group:       "OAuth",
	},
	{
		Name:        "OAUTH_GOOGLE_CLIENT_SECRET",
		Description: "Google OAuth client secret",
		Required:    true,
		Secret:      true,
		Group:       "OAuth",
	},
	{
		Name:        "OAUTH_GOOGLE_REDIRECT_URL",
		Description: "Google OAuth redirect URL",
		Required:    true,
		Secret:      true,
		Group:       "OAuth",
	},
	{
		Name:        "OAUTH_GITHUB_CLIENT_ID",
		Description: "GitHub OAuth client ID",
		Secret:      true,
		Group:       "OAuth",
	},
	{
		Name:        "OAUTH_GITHUB_CLIENT_SECRET",
		Description: "GitHub OAuth client secret",
		Secret:      true,
		Group:       "OAuth",
	},

	// ================================================================
	// Feature Flags
	// ================================================================
	{
		Name:        "FEATURE_NEW_UI",
		Description: "Enable new UI (true/false)",
		Default:     "false",
		Group:       "Features",
	},
	{
		Name:        "FEATURE_BETA_API",
		Description: "Enable beta API endpoints (true/false)",
		Default:     "false",
		Group:       "Features",
	},
	{
		Name:        "FEATURE_ANALYTICS",
		Description: "Enable analytics tracking (true/false)",
		Default:     "true",
		Group:       "Features",
	},

	// ================================================================
	// Redis Cache (optional)
	// ================================================================
	{
		Name:        "REDIS_URL",
		Description: "Redis connection URL",
		Secret:      true,
		Group:       "Cache",
	},
	{
		Name:        "REDIS_PASSWORD",
		Description: "Redis password",
		Secret:      true,
		Group:       "Cache",
	},
	{
		Name:        "CACHE_TTL",
		Description: "Cache TTL in seconds",
		Default:     "3600",
		Group:       "Cache",
	},
}

// AppRegistry is the global registry instance for this application
var AppRegistry = env.NewRegistry(AppEnvVars)
