// Package cli provides a lightweight command-line interface framework using only stdlib.
// It offers command routing, help generation, and flag parsing without external dependencies.
package cli

import "flag"

// Command represents a CLI command with its configuration.
type Command struct {
	Name        string           // Command name (e.g., "list", "setup")
	Usage       string           // Short description shown in help
	Category    string           // Category for grouping in help (e.g., "Setup & Validation")
	Action      func() error     // Function to execute (no arguments)
	ActionFlags func([]string)   // Alternative: function that handles its own flags
	Flags       *flag.FlagSet    // Command-specific flags (optional)
	Hidden      bool             // Hide from help output
}

// App represents the CLI application configuration.
type App struct {
	Name     string              // Application name
	Usage    string              // Application description
	Commands []Command           // Registered commands
	Before   func() error        // Hook to run before any command
	Flags    *flag.FlagSet       // Global flags
}

// AppConfig contains configuration for creating a new App.
type AppConfig struct {
	Name  string  // Application name
	Usage string  // Application description
}
