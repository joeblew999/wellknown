package scaffold

// minimalTemplate is a starter registry with basic common variables
const minimalTemplate = `package {{.PackageName}}

import "{{.ImportPath}}/pkg/env"

// ================================================================
// {{.AppName}} - Environment Registry
// ================================================================
// This is the SINGLE SOURCE OF TRUTH for all environment variables.
// Edit this file ONCE, use everywhere (local + production).
//
// After editing this file, run:
//   go run . sync-registry
//
// Learn more: See WORKFLOW.md for the complete workflow

// AppEnvVars defines all environment variables for this application
var AppEnvVars = []env.EnvVar{
	// ----------------------------------------------------------------
	// Server
	// ----------------------------------------------------------------
	{
		Name:    "SERVER_PORT",
		Default: "8080",
		Group:   "Server",
		Comment: "HTTP server port",
	},
	{
		Name:    "LOG_LEVEL",
		Default: "info",
		Group:   "Server",
		Comment: "Logging level (debug, info, warn, error)",
	},

	// ----------------------------------------------------------------
	// Database
	// ----------------------------------------------------------------
	{
		Name:     "DATABASE_URL",
		Required: true,
		Secret:   true,
		Group:    "Database",
		Comment:  "PostgreSQL connection string",
	},
}

// AppRegistry is the initialized registry instance
var AppRegistry = env.NewRegistry(AppEnvVars)
`

// fullTemplate is a comprehensive example showing all features
const fullTemplate = `package {{.PackageName}}

import "{{.ImportPath}}/pkg/env"

// ================================================================
// {{.AppName}} - Environment Registry
// ================================================================
// This is the SINGLE SOURCE OF TRUTH for all environment variables.
// Edit this file ONCE, use everywhere (local + production).
//
// After editing this file, run:
//   go run . sync-registry
//
// Learn more: See WORKFLOW.md for the complete workflow

// AppEnvVars defines all environment variables for this application
var AppEnvVars = []env.EnvVar{
	// ----------------------------------------------------------------
	// Server
	// ----------------------------------------------------------------
	{
		Name:    "SERVER_PORT",
		Default: "8080",
		Group:   "Server",
		Comment: "HTTP server port",
	},
	{
		Name:    "SERVER_HOST",
		Default: "0.0.0.0",
		Group:   "Server",
		Comment: "HTTP server bind address",
	},
	{
		Name:    "LOG_LEVEL",
		Default: "info",
		Group:   "Server",
		Comment: "Logging level (debug, info, warn, error)",
	},
	{
		Name:    "LOG_FORMAT",
		Default: "json",
		Group:   "Server",
		Comment: "Log format (json, text)",
	},

	// ----------------------------------------------------------------
	// Database
	// ----------------------------------------------------------------
	{
		Name:     "DATABASE_URL",
		Required: true,
		Secret:   true,
		Group:    "Database",
		Comment:  "PostgreSQL connection string",
	},
	{
		Name:    "DATABASE_POOL_SIZE",
		Default: "10",
		Group:   "Database",
		Comment: "Maximum number of database connections",
	},
	{
		Name:    "DATABASE_POOL_TIMEOUT",
		Default: "30s",
		Group:   "Database",
		Comment: "Connection pool timeout",
	},

	// ----------------------------------------------------------------
	// APIs (Secrets)
	// ----------------------------------------------------------------
	{
		Name:     "STRIPE_API_KEY",
		Required: true,
		Secret:   true,
		Group:    "APIs",
		Comment:  "Stripe API key for payments",
	},
	{
		Name:   "SENDGRID_API_KEY",
		Secret: true,
		Group:  "APIs",
		Comment: "SendGrid API key for email (optional)",
	},
	{
		Name:   "AWS_ACCESS_KEY_ID",
		Secret: true,
		Group:  "APIs",
		Comment: "AWS access key for S3 storage (optional)",
	},
	{
		Name:   "AWS_SECRET_ACCESS_KEY",
		Secret: true,
		Group:  "APIs",
		Comment: "AWS secret key for S3 storage (optional)",
	},

	// ----------------------------------------------------------------
	// Feature Flags
	// ----------------------------------------------------------------
	{
		Name:    "FEATURE_BETA_UI",
		Default: "false",
		Group:   "Features",
		Comment: "Enable beta UI features",
	},
	{
		Name:    "FEATURE_ANALYTICS",
		Default: "true",
		Group:   "Features",
		Comment: "Enable analytics tracking",
	},

	// ----------------------------------------------------------------
	// External Services
	// ----------------------------------------------------------------
	{
		Name:    "REDIS_URL",
		Default: "redis://localhost:6379",
		Group:   "External Services",
		Comment: "Redis connection URL for caching",
	},
	{
		Name:    "SENTRY_DSN",
		Secret:  true,
		Group:   "External Services",
		Comment: "Sentry DSN for error tracking (optional)",
	},
}

// AppRegistry is the initialized registry instance
var AppRegistry = env.NewRegistry(AppEnvVars)
`
