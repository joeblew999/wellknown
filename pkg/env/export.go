// Package env provides environment variable export utilities.
package env

import (
	"fmt"
	"os"
	"strings"
)

// ExportFormat specifies output format for environment variables.
type ExportFormat string

const (
	FormatSimple  ExportFormat = "simple"  // KEY=VALUE
	FormatDocker  ExportFormat = "docker"  // Same as simple (for backward compatibility)
	FormatSystemd ExportFormat = "systemd" // Environment="KEY=VALUE"
	FormatK8s     ExportFormat = "k8s"     // - name: KEY\n  value: VALUE
)

// ExportOptions controls environment variable export behavior.
// Use this to filter which variables are exported and how they're formatted.
type ExportOptions struct {
	Format       ExportFormat // Output format
	SecretsOnly  bool         // Export only secret vars
	RequiredOnly bool         // Export only required vars
	IncludeEmpty bool         // Include vars with empty values
	MaskSecrets  bool         // Replace secret values with ***
}

// Export formats environment variables according to options.
//
// This is the main export function that applies filtering and formatting.
// It reads actual values from the environment (os.Getenv).
//
// Example:
//
//	output := registry.Export(ExportOptions{
//	  Format: FormatSimple,
//	  SecretsOnly: true,
//	  IncludeEmpty: false,
//	})
func (r *Registry) Export(opts ExportOptions) string {
	// Get variables to export based on filters
	var varsToExport []EnvVar

	for _, v := range r.vars {
		// Apply filters
		if opts.SecretsOnly && !v.Secret {
			continue
		}
		if opts.RequiredOnly && !v.Required {
			continue
		}

		// Get actual value from environment
		value := os.Getenv(v.Name)

		// Skip empty values unless explicitly included
		if !opts.IncludeEmpty && value == "" {
			continue
		}

		varsToExport = append(varsToExport, v)
	}

	// Format output
	return formatVars(varsToExport, opts)
}

// formatVars formats a list of variables according to the specified format.
func formatVars(vars []EnvVar, opts ExportOptions) string {
	var lines []string

	for _, v := range vars {
		value := os.Getenv(v.Name)

		// Mask secrets if requested
		if opts.MaskSecrets && v.Secret && value != "" {
			value = "***"
		}

		// Format based on type
		switch opts.Format {
		case FormatSimple, FormatDocker:
			lines = append(lines, fmt.Sprintf("%s=%s", v.Name, value))

		case FormatSystemd:
			lines = append(lines, fmt.Sprintf("Environment=\"%s=%s\"", v.Name, value))

		case FormatK8s:
			lines = append(lines, fmt.Sprintf("- name: %s", v.Name))
			lines = append(lines, fmt.Sprintf("  value: \"%s\"", value))

		default:
			// Default to simple format
			lines = append(lines, fmt.Sprintf("%s=%s", v.Name, value))
		}
	}

	return strings.Join(lines, "\n")
}

// ExportSimple is a convenience method for simple KEY=VALUE format.
// Exports all variables with non-empty values.
func (r *Registry) ExportSimple() string {
	return r.Export(ExportOptions{
		Format:       FormatSimple,
		IncludeEmpty: false,
	})
}

// ExportSecrets is a convenience method for exporting only secret variables.
// Uses simple KEY=VALUE format, excludes empty values.
//
// This is useful for generating files for tools like flyctl secrets import.
func (r *Registry) ExportSecrets() string {
	return r.Export(ExportOptions{
		Format:       FormatSimple,
		SecretsOnly:  true,
		IncludeEmpty: false,
	})
}

// ExportRequired is a convenience method for exporting only required variables.
// Uses simple KEY=VALUE format, includes empty values (to show what's missing).
func (r *Registry) ExportRequired() string {
	return r.Export(ExportOptions{
		Format:       FormatSimple,
		RequiredOnly: true,
		IncludeEmpty: true,
	})
}

// ExportSystemd exports variables in systemd Environment= format.
func (r *Registry) ExportSystemd() string {
	return r.Export(ExportOptions{
		Format:       FormatSystemd,
		IncludeEmpty: false,
	})
}

// ExportK8s exports variables in Kubernetes YAML format.
func (r *Registry) ExportK8s() string {
	return r.Export(ExportOptions{
		Format:       FormatK8s,
		IncludeEmpty: false,
	})
}
