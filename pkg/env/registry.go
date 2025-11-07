// Package env provides generic environment variable management for Go applications.
// This package handles env var registration, validation, type conversion, and secrets management.
package env

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// EnvVar represents an environment variable definition with metadata.
// This is the core type for registering and managing environment variables.
type EnvVar struct {
	Name        string // Environment variable name (e.g., "SERVER_PORT")
	Description string // Human-readable description
	Required    bool   // Is this variable required?
	Secret      bool   // Should this be treated as a secret (masked in logs, etc.)?
	Default     string // Default value (empty string if no default)
	Group       string // Logical grouping for organization (e.g., "Server", "OAuth")
}

// Registry holds a collection of environment variables and provides lookup/filtering operations.
type Registry struct {
	vars  []EnvVar
	index map[string]*EnvVar // Fast lookup by name
}

// NewRegistry creates a new environment variable registry from a slice of EnvVar.
func NewRegistry(vars []EnvVar) *Registry {
	r := &Registry{
		vars:  vars,
		index: make(map[string]*EnvVar, len(vars)),
	}

	// Build index for O(1) lookup
	for i := range r.vars {
		r.index[r.vars[i].Name] = &r.vars[i]
	}

	return r
}

// ByName returns the environment variable with the given name, or nil if not found.
func (r *Registry) ByName(name string) *EnvVar {
	return r.index[name]
}

// GetRequired returns all required environment variables.
func (r *Registry) GetRequired() []EnvVar {
	var required []EnvVar
	for _, v := range r.vars {
		if v.Required {
			required = append(required, v)
		}
	}
	return required
}

// GetSecrets returns all environment variables marked as secrets.
func (r *Registry) GetSecrets() []EnvVar {
	var secrets []EnvVar
	for _, v := range r.vars {
		if v.Secret {
			secrets = append(secrets, v)
		}
	}
	return secrets
}

// GetByGroup returns a map of environment variables grouped by their Group field.
func (r *Registry) GetByGroup() map[string][]EnvVar {
	groups := make(map[string][]EnvVar)
	for _, v := range r.vars {
		groups[v.Group] = append(groups[v.Group], v)
	}
	return groups
}

// All returns all environment variables in the registry.
func (r *Registry) All() []EnvVar {
	return r.vars
}

// AllSorted returns all environment variables sorted by group and name.
func (r *Registry) AllSorted() []EnvVar {
	sorted := make([]EnvVar, len(r.vars))
	copy(sorted, r.vars)

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Group != sorted[j].Group {
			return sorted[i].Group < sorted[j].Group
		}
		return sorted[i].Name < sorted[j].Name
	})

	return sorted
}

// ValidateRequired checks if all required environment variables are set.
// Returns an error listing any missing required variables.
func (r *Registry) ValidateRequired() error {
	var missing []string
	for _, v := range r.GetRequired() {
		if os.Getenv(v.Name) == "" {
			missing = append(missing, v.Name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}

	return nil
}

// GetString returns the value of the environment variable as a string.
// If the variable is not set, returns the default value.
func (e *EnvVar) GetString() string {
	if value := os.Getenv(e.Name); value != "" {
		return value
	}
	return e.Default
}

// GetInt returns the value of the environment variable as an integer.
// If the variable is not set or cannot be parsed, returns the default value as an int.
// If the default cannot be parsed, returns 0.
func (e *EnvVar) GetInt() int {
	if value := os.Getenv(e.Name); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		// TODO: Add warning logging for parse failures
	}

	// Parse default value
	if e.Default != "" {
		if defaultInt, err := strconv.Atoi(e.Default); err == nil {
			return defaultInt
		}
	}

	return 0
}

// GetBool returns the value of the environment variable as a boolean.
// Accepts "true", "1", "yes" as true; "false", "0", "no" as false.
// If the variable is not set or cannot be parsed, returns the default value as a bool.
// If the default cannot be parsed, returns false.
func (e *EnvVar) GetBool() bool {
	if value := os.Getenv(e.Name); value != "" {
		switch strings.ToLower(value) {
		case "true", "1", "yes":
			return true
		case "false", "0", "no":
			return false
		default:
			// TODO: Add warning logging for parse failures
		}
	}

	// Parse default value
	if e.Default != "" {
		switch strings.ToLower(e.Default) {
		case "true", "1", "yes":
			return true
		case "false", "0", "no":
			return false
		}
	}

	return false
}
