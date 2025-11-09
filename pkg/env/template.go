package env

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// ================================================================
// Template Generation - Generic Environment File Builders
// ================================================================

// TemplateOptions configures environment file template generation
type TemplateOptions struct {
	// Header lines to prepend (typically comments)
	Header []string

	// Footer lines to append (typically comments)
	Footer []string

	// GroupOrder specifies the order of groups (if empty, alphabetical)
	GroupOrder []string

	// ValueOverrides provides custom values for specific variables
	// Function signature: func(envVar EnvVar) (customValue string, useCustom bool)
	ValueOverrides func(EnvVar) (string, bool)

	// IncludeComments adds description/required comments above each variable
	IncludeComments bool

	// IncludeGroupHeaders adds group section headers
	IncludeGroupHeaders bool

	// GroupHeaderFormat formats group headers (receives group name)
	// Default: "# ----------------------------------------------------------------\n# %s\n# ----------------------------------------------------------------\n"
	GroupHeaderFormat func(groupName string) string
}

// GenerateTemplate creates an environment file template from the registry
// This is the core generic template builder used by all format-specific functions
func (r *Registry) GenerateTemplate(opts TemplateOptions) string {
	var sb strings.Builder

	// Write header
	for _, line := range opts.Header {
		sb.WriteString(line)
		if !strings.HasSuffix(line, "\n") {
			sb.WriteString("\n")
		}
	}

	// Get groups
	groups := r.GetByGroup()

	// Determine group ordering
	var groupNames []string
	if len(opts.GroupOrder) > 0 {
		// Use specified order, but only include groups that exist
		for _, name := range opts.GroupOrder {
			if _, exists := groups[name]; exists {
				groupNames = append(groupNames, name)
			}
		}
	} else {
		// Alphabetical order
		groupNames = make([]string, 0, len(groups))
		for name := range groups {
			groupNames = append(groupNames, name)
		}
		sort.Strings(groupNames)
	}

	// Default group header format
	groupHeaderFmt := opts.GroupHeaderFormat
	if groupHeaderFmt == nil {
		groupHeaderFmt = func(groupName string) string {
			return fmt.Sprintf("# ----------------------------------------------------------------\n# %s\n# ----------------------------------------------------------------\n", groupName)
		}
	}

	// Process each group
	for _, groupName := range groupNames {
		vars := groups[groupName]

		// Sort variables within group alphabetically by name
		sort.Slice(vars, func(i, j int) bool {
			return vars[i].Name < vars[j].Name
		})

		// Write group header
		if opts.IncludeGroupHeaders {
			sb.WriteString(groupHeaderFmt(groupName))
		}

		// Write each variable
		for _, v := range vars {
			// Add description comment
			if opts.IncludeComments && v.Description != "" {
				sb.WriteString(fmt.Sprintf("# %s\n", v.Description))
			}

			// Mark as required
			if opts.IncludeComments && v.Required {
				sb.WriteString("# REQUIRED\n")
			}

			// Determine value (custom override or default)
			var value string
			if opts.ValueOverrides != nil {
				if customValue, useCustom := opts.ValueOverrides(v); useCustom {
					value = customValue
				} else {
					value = v.Default
				}
			} else {
				value = v.Default
			}

			// Write variable line
			if value != "" {
				sb.WriteString(fmt.Sprintf("%s=%s\n", v.Name, value))
			} else {
				sb.WriteString(fmt.Sprintf("%s=\n", v.Name))
			}
			sb.WriteString("\n")
		}
	}

	// Write footer
	for _, line := range opts.Footer {
		sb.WriteString(line)
		if !strings.HasSuffix(line, "\n") {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// ================================================================
// Pre-Built Template Formats
// ================================================================

// GenerateEnvExample creates a .env.example file with all variables and descriptions
func (r *Registry) GenerateEnvExample(appName string) string {
	return r.GenerateTemplate(TemplateOptions{
		Header: []string{
			"# ================================================================",
			fmt.Sprintf("# %s Environment Variables", appName),
			"# ================================================================",
			"# This file is auto-generated - DO NOT EDIT MANUALLY",
			"# Copy this to .env and configure for your environment",
			"# ================================================================\n",
		},
		IncludeComments:     true,
		IncludeGroupHeaders: true,
	})
}

// GenerateEnvList creates a human-readable listing of all environment variables
// Shows current values with secrets masked
func (r *Registry) GenerateEnvList(title string) string {
	var sb strings.Builder

	if title == "" {
		title = "Environment Variables Registry"
	}

	sb.WriteString(title + "\n")
	sb.WriteString(strings.Repeat("=", len(title)) + "\n\n")

	groups := r.GetByGroup()
	groupNames := make([]string, 0, len(groups))
	for name := range groups {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	for _, groupName := range groupNames {
		vars := groups[groupName]
		sb.WriteString(fmt.Sprintf("## %s\n", groupName))

		for _, v := range vars {
			// Status badges
			status := ""
			if v.Required {
				status = " [REQUIRED]"
			}
			if v.Secret {
				status += " [SECRET]"
			}

			// Current value (masked if secret)
			currentValue := os.Getenv(v.Name)
			valueDisplay := "not set"
			if currentValue != "" {
				if v.Secret {
					valueDisplay = "***set***"
				} else {
					valueDisplay = currentValue
				}
			} else if v.Default != "" {
				valueDisplay = fmt.Sprintf("(default: %s)", v.Default)
			}

			sb.WriteString(fmt.Sprintf("  %s%s\n", v.Name, status))
			sb.WriteString(fmt.Sprintf("    %s\n", v.Description))
			sb.WriteString(fmt.Sprintf("    Current: %s\n", valueDisplay))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// ================================================================
// Dockerfile-Style Documentation Generator
// ================================================================

// DockerfileDocsOptions configures Dockerfile environment documentation generation
type DockerfileDocsOptions struct {
	// AppName for header comments
	AppName string

	// UpdateCommand shown in footer (e.g., "make env-sync-dockerfile")
	UpdateCommand string

	// DeploymentPlatform (e.g., "Fly.io", "AWS ECS")
	DeploymentPlatform string

	// NonSecretEnvSource describes where non-secret vars come from (e.g., "fly.toml [env] section")
	NonSecretEnvSource string

	// SecretSource describes where secrets come from (e.g., "Fly.io secrets")
	SecretSource string

	// SyncCommand shown in footer (e.g., "make fly-secrets")
	SyncCommand string
}

// GenerateDockerfileDocs creates Dockerfile-style environment variable documentation
// Categorizes variables by required/optional and secret/non-secret
func (r *Registry) GenerateDockerfileDocs(opts DockerfileDocsOptions) string {
	var sb strings.Builder

	// Header
	sb.WriteString("# ================================================================\n")
	sb.WriteString(fmt.Sprintf("# Environment Variables (injected at runtime by %s)\n", opts.DeploymentPlatform))
	sb.WriteString("# AUTO-GENERATED - DO NOT EDIT MANUALLY\n")
	if opts.UpdateCommand != "" {
		sb.WriteString(fmt.Sprintf("# To update: %s\n", opts.UpdateCommand))
	}
	sb.WriteString("# ================================================================\n")
	sb.WriteString("#\n")

	allVars := r.All()

	// Section 1: Required non-secrets
	sb.WriteString(fmt.Sprintf("# Required (set via %s):\n", opts.NonSecretEnvSource))
	hasNonSecretRequired := false
	for _, v := range allVars {
		if !v.Secret && v.Required {
			hasNonSecretRequired = true
			defaultVal := v.Default
			if defaultVal == "" {
				defaultVal = "<value>"
			}
			sb.WriteString(fmt.Sprintf("#   %s=%s\n", v.Name, defaultVal))
			if v.Description != "" {
				sb.WriteString(fmt.Sprintf("#     %s\n", v.Description))
			}
		}
	}
	if !hasNonSecretRequired {
		sb.WriteString("#   (none)\n")
	}
	sb.WriteString("#\n")

	// Section 2: Required secrets
	sb.WriteString(fmt.Sprintf("# Required (set via %s):\n", opts.SecretSource))
	hasSecretRequired := false
	for _, v := range allVars {
		if v.Secret && v.Required {
			hasSecretRequired = true
			sb.WriteString(fmt.Sprintf("#   %s\n", v.Name))
			if v.Description != "" {
				sb.WriteString(fmt.Sprintf("#     %s\n", v.Description))
			}
		}
	}
	if !hasSecretRequired {
		sb.WriteString("#   (none)\n")
	}
	sb.WriteString("#\n")

	// Section 3: Optional secrets (grouped)
	sb.WriteString(fmt.Sprintf("# Optional (set via %s if needed):\n", opts.SecretSource))
	groups := r.GetByGroup()
	groupNames := make([]string, 0, len(groups))
	for name := range groups {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	hasOptionalSecrets := false
	for _, groupName := range groupNames {
		vars := groups[groupName]
		groupHasSecrets := false
		for _, v := range vars {
			if v.Secret && !v.Required {
				if !groupHasSecrets {
					sb.WriteString(fmt.Sprintf("#   # %s\n", groupName))
					groupHasSecrets = true
					hasOptionalSecrets = true
				}
				sb.WriteString(fmt.Sprintf("#   %s\n", v.Name))
				if v.Description != "" {
					sb.WriteString(fmt.Sprintf("#     %s\n", v.Description))
				}
			}
		}
	}
	if !hasOptionalSecrets {
		sb.WriteString("#   (none)\n")
	}
	sb.WriteString("#\n")

	// Footer
	if opts.SyncCommand != "" {
		sb.WriteString(fmt.Sprintf("# Sync secrets: %s\n", opts.SyncCommand))
	}
	sb.WriteString("# ================================================================\n")

	return sb.String()
}

// ================================================================
// TOML Format Generator (for fly.toml, etc.)
// ================================================================

// GenerateTOMLEnv generates a TOML [env] section with non-secret defaults
func (r *Registry) GenerateTOMLEnv(sectionName string, comments []string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("[%s]\n", sectionName))
	for _, comment := range comments {
		sb.WriteString(fmt.Sprintf("  # %s\n", comment))
	}

	for _, v := range r.All() {
		if !v.Secret && v.Default != "" {
			sb.WriteString(fmt.Sprintf("  %s = \"%s\"\n", v.Name, v.Default))
		}
	}

	return sb.String()
}

// GenerateTOMLSecretsList generates a TOML comment list of secret variables
// for deployment config files like fly.toml. Includes auto-sync markers.
//
// Parameters:
//   - importCommand: Command to import secrets (e.g., "go run . fly-secrets-import")
//
// Returns:
//
//	Auto-sync section with commented list of secret variable names
//
// Example:
//
//	# === START AUTO-GENERATED SECRETS LIST ===
//	# Secrets (set via: go run . fly-secrets-import)
//	# - DATABASE_URL
//	# - API_KEY
//	# === END AUTO-GENERATED SECRETS LIST ===
func (r *Registry) GenerateTOMLSecretsList(importCommand string) string {
	var sb strings.Builder

	sb.WriteString("# === START AUTO-GENERATED SECRETS LIST ===\n")
	sb.WriteString(fmt.Sprintf("# Secrets (set via: %s)\n", importCommand))

	for _, v := range r.All() {
		if v.Secret {
			sb.WriteString(fmt.Sprintf("# - %s\n", v.Name))
		}
	}

	sb.WriteString("# === END AUTO-GENERATED SECRETS LIST ===")

	return sb.String()
}

// GenerateDockerComposeEnv generates a docker-compose.yml environment section
// with non-secret variables that have defaults.
//
// Parameters:
//   - comments: Comment lines to include (e.g., update instructions)
//
// Returns:
//
//	Docker Compose YAML environment section with proper indentation
//
// Example:
//
//	environment:
//	  # AUTO-GENERATED - DO NOT EDIT MANUALLY
//	  SERVER_PORT: "8080"
//	  LOG_LEVEL: "info"
func (r *Registry) GenerateDockerComposeEnv(comments []string) string {
	var sb strings.Builder

	sb.WriteString("    environment:\n")
	for _, comment := range comments {
		sb.WriteString(fmt.Sprintf("      # %s\n", comment))
	}

	for _, v := range r.All() {
		if !v.Secret && v.Default != "" {
			sb.WriteString(fmt.Sprintf("      %s: \"%s\"\n", v.Name, v.Default))
		}
	}

	return sb.String()
}
