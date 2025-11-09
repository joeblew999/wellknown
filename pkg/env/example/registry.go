package main

import "github.com/joeblew999/wellknown/pkg/env"

// AppEnvVars - Single source of truth for ALL environments
// Edit this ONCE, use everywhere (local + production)
var AppEnvVars = []env.EnvVar{
	// Server
	{Name: "SERVER_PORT", Default: "8080", Group: "Server"},
	{Name: "LOG_LEVEL", Default: "info", Group: "Server"},

	// Database (secret)
	{Name: "DATABASE_URL", Required: true, Secret: true, Group: "Database"},

	// External APIs (secrets)
	{Name: "STRIPE_API_KEY", Required: true, Secret: true, Group: "APIs"},
	{Name: "SENDGRID_API_KEY", Secret: true, Group: "APIs"},
	{Name: "OPENAI_API_KEY", Secret: true, Group: "APIs"},

	// Feature Flags
	{Name: "FEATURE_BETA", Default: "false", Group: "Features"},
}

var AppRegistry = env.NewRegistry(AppEnvVars)
