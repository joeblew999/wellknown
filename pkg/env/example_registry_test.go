package env_test

import (
	"fmt"
	"os"

	"github.com/joeblew999/wellknown/pkg/env"
)

// Example demonstrates basic registry usage
func Example() {
	// Define environment variables
	vars := []env.EnvVar{
		{
			Name:        "SERVER_PORT",
			Description: "HTTP server port",
			Default:     "8080",
			Required:    true,
			Group:       "Server",
		},
		{
			Name:        "LOG_LEVEL",
			Description: "Logging level",
			Default:     "info",
			Required:    false,
			Group:       "Logging",
		},
	}

	// Create registry
	registry := env.NewRegistry(vars)

	// Get individual variables
	serverPort := registry.ByName("SERVER_PORT")
	if serverPort != nil {
		fmt.Printf("Port: %s\n", serverPort.GetString())
	}

	// Output:
	// Port: 8080
}

// ExampleNewRegistry shows how to create and use a registry
func ExampleNewRegistry() {
	vars := []env.EnvVar{
		{Name: "API_KEY", Secret: true, Required: true},
		{Name: "DEBUG", Default: "false"},
	}

	registry := env.NewRegistry(vars)

	// Look up variables by name
	apiKey := registry.ByName("API_KEY")
	if apiKey != nil {
		fmt.Printf("API_KEY is required: %v\n", apiKey.Required)
		fmt.Printf("API_KEY is secret: %v\n", apiKey.Secret)
	}

	// Output:
	// API_KEY is required: true
	// API_KEY is secret: true
}

// ExampleRegistry_ByName demonstrates variable lookup
func ExampleRegistry_ByName() {
	vars := []env.EnvVar{
		{Name: "DATABASE_URL", Description: "Database connection string"},
		{Name: "CACHE_TTL", Description: "Cache time-to-live", Default: "3600"},
	}

	registry := env.NewRegistry(vars)

	// Look up existing variable
	dbURL := registry.ByName("DATABASE_URL")
	if dbURL != nil {
		fmt.Println("Found DATABASE_URL")
	}

	// Look up non-existent variable
	missing := registry.ByName("NONEXISTENT")
	if missing == nil {
		fmt.Println("NONEXISTENT not found")
	}

	// Output:
	// Found DATABASE_URL
	// NONEXISTENT not found
}

// ExampleRegistry_GetRequired shows filtering required variables
func ExampleRegistry_GetRequired() {
	vars := []env.EnvVar{
		{Name: "API_KEY", Required: true},
		{Name: "DB_PASSWORD", Required: true},
		{Name: "LOG_LEVEL", Required: false},
	}

	registry := env.NewRegistry(vars)
	required := registry.GetRequired()

	fmt.Printf("Required variables: %d\n", len(required))
	for _, v := range required {
		fmt.Printf("- %s\n", v.Name)
	}

	// Output:
	// Required variables: 2
	// - API_KEY
	// - DB_PASSWORD
}

// ExampleRegistry_GetSecrets shows filtering secret variables
func ExampleRegistry_GetSecrets() {
	vars := []env.EnvVar{
		{Name: "PUBLIC_KEY", Secret: false},
		{Name: "PRIVATE_KEY", Secret: true},
		{Name: "API_SECRET", Secret: true},
	}

	registry := env.NewRegistry(vars)
	secrets := registry.GetSecrets()

	fmt.Printf("Secret variables: %d\n", len(secrets))
	for _, v := range secrets {
		fmt.Printf("- %s\n", v.Name)
	}

	// Output:
	// Secret variables: 2
	// - PRIVATE_KEY
	// - API_SECRET
}

// ExampleRegistry_GetByGroup shows grouping variables
func ExampleRegistry_GetByGroup() {
	vars := []env.EnvVar{
		{Name: "SERVER_PORT", Group: "Server"},
		{Name: "SERVER_HOST", Group: "Server"},
		{Name: "DB_URL", Group: "Database"},
	}

	registry := env.NewRegistry(vars)
	groups := registry.GetByGroup()

	for groupName, groupVars := range groups {
		fmt.Printf("%s: %d vars\n", groupName, len(groupVars))
	}

	// Unordered output:
	// Server: 2 vars
	// Database: 1 vars
}

// ExampleRegistry_ValidateRequired demonstrates validation
func ExampleRegistry_ValidateRequired() {
	vars := []env.EnvVar{
		{Name: "REQUIRED_VAR", Required: true},
		{Name: "OPTIONAL_VAR", Required: false},
	}

	registry := env.NewRegistry(vars)

	// Set required variable
	os.Setenv("REQUIRED_VAR", "value")
	defer os.Unsetenv("REQUIRED_VAR")

	// Validate should pass
	err := registry.ValidateRequired()
	if err != nil {
		fmt.Println("Validation failed:", err)
	} else {
		fmt.Println("All required variables are set")
	}

	// Output:
	// All required variables are set
}

// ExampleEnvVar_GetString shows string variable retrieval
func ExampleEnvVar_GetString() {
	envVar := env.EnvVar{
		Name:    "MESSAGE",
		Default: "Hello, World!",
	}

	// Environment variable not set, uses default
	value := envVar.GetString()
	fmt.Println(value)

	// Set environment variable
	os.Setenv("MESSAGE", "Custom message")
	defer os.Unsetenv("MESSAGE")

	value = envVar.GetString()
	fmt.Println(value)

	// Output:
	// Hello, World!
	// Custom message
}

// ExampleEnvVar_GetInt shows integer variable retrieval
func ExampleEnvVar_GetInt() {
	envVar := env.EnvVar{
		Name:    "PORT",
		Default: "8080",
	}

	port := envVar.GetInt()
	fmt.Printf("Port: %d\n", port)

	// Set environment variable
	os.Setenv("PORT", "3000")
	defer os.Unsetenv("PORT")

	port = envVar.GetInt()
	fmt.Printf("Port: %d\n", port)

	// Output:
	// Port: 8080
	// Port: 3000
}

// ExampleEnvVar_GetBool shows boolean variable retrieval
func ExampleEnvVar_GetBool() {
	envVar := env.EnvVar{
		Name:    "DEBUG",
		Default: "false",
	}

	debug := envVar.GetBool()
	fmt.Printf("Debug: %v\n", debug)

	// Set environment variable (various formats work)
	os.Setenv("DEBUG", "true")
	fmt.Printf("Debug: %v\n", envVar.GetBool())

	os.Setenv("DEBUG", "1")
	fmt.Printf("Debug: %v\n", envVar.GetBool())

	os.Setenv("DEBUG", "yes")
	fmt.Printf("Debug: %v\n", envVar.GetBool())

	os.Unsetenv("DEBUG")

	// Output:
	// Debug: false
	// Debug: true
	// Debug: true
	// Debug: true
}

// ExampleRegistry_AllSorted shows sorted variable listing
func ExampleRegistry_AllSorted() {
	vars := []env.EnvVar{
		{Name: "Z_VAR", Group: "Z"},
		{Name: "A_VAR", Group: "A"},
		{Name: "B_VAR", Group: "A"},
	}

	registry := env.NewRegistry(vars)
	sorted := registry.AllSorted()

	for _, v := range sorted {
		fmt.Printf("%s/%s\n", v.Group, v.Name)
	}

	// Output:
	// A/A_VAR
	// A/B_VAR
	// Z/Z_VAR
}
