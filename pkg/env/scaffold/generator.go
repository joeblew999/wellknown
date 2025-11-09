// Package scaffold provides registry generation functionality for new projects
package scaffold

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

// GeneratorOptions configures registry file generation
type GeneratorOptions struct {
	Dir         string // Target directory (default: ".")
	AppName     string // Application name for headers (default: "My Application")
	PackageName string // Go package name (default: "main")
	Template    string // "minimal" or "full" (default: "minimal")
	Force       bool   // Overwrite existing registry.go (default: false)
	ImportPath  string // Module import path (default: "github.com/joeblew999/wellknown")
}

// GenerateRegistry creates a new registry.go file with smart defaults
// Returns an error if the file exists and Force is false
func GenerateRegistry(opts GeneratorOptions) error {
	// Apply defaults
	if opts.Dir == "" {
		opts.Dir = "."
	}
	if opts.AppName == "" {
		opts.AppName = "My Application"
	}
	if opts.PackageName == "" {
		opts.PackageName = "main"
	}
	if opts.Template == "" {
		opts.Template = "minimal"
	}
	if opts.ImportPath == "" {
		opts.ImportPath = "github.com/joeblew999/wellknown"
	}

	// Construct file path
	registryPath := filepath.Join(opts.Dir, "registry.go")

	// Check if file exists
	if _, err := os.Stat(registryPath); err == nil {
		if !opts.Force {
			return fmt.Errorf("registry.go already exists at %s (use --force to overwrite)", registryPath)
		}

		// Create backup before overwriting
		backupPath := fmt.Sprintf("%s.backup.%d", registryPath, time.Now().Unix())
		if err := os.Rename(registryPath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		fmt.Printf("   📦 Created backup: %s\n", filepath.Base(backupPath))
	}

	// Select template
	var templateStr string
	switch opts.Template {
	case "minimal":
		templateStr = minimalTemplate
	case "full":
		templateStr = fullTemplate
	default:
		return fmt.Errorf("unknown template: %s (use 'minimal' or 'full')", opts.Template)
	}

	// Parse and execute template
	tmpl, err := template.New("registry").Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, opts); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Ensure target directory exists
	if err := os.MkdirAll(opts.Dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", opts.Dir, err)
	}

	// Write file
	if err := os.WriteFile(registryPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", registryPath, err)
	}

	return nil
}

// RegistryExists checks if registry.go already exists in the given directory
func RegistryExists(dir string) bool {
	if dir == "" {
		dir = "."
	}
	_, err := os.Stat(filepath.Join(dir, "registry.go"))
	return err == nil
}
