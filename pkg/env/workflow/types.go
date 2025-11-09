// Package workflow provides high-level workflow orchestration for environment management
package workflow

import (
	"io"

	"github.com/joeblew999/wellknown/pkg/env"
)

// ================================================================
// Options Structures
// ================================================================

// RegistrySyncOptions configures the registry synchronization workflow
type RegistrySyncOptions struct {
	Registry           *env.Registry       // The registry to sync from
	AppName            string              // Application name for headers
	DeploymentConfigs  []DeploymentConfig  // Optional deployment configs to sync
	CreateSecretsFiles bool                // Create .env.secrets.* templates if missing
	OutputWriter       io.Writer           // Where to write progress messages (nil = discard)
	SyncOnlyConfigs    []string            // Optional: only sync these config files (nil = sync all)
	SkipEnvironments   bool                // Skip .env.local/.env.production generation
}

// DeploymentConfig defines a deployment configuration file to sync
type DeploymentConfig struct {
	FilePath    string                             // Path to the config file
	StartMarker string                             // Start marker for auto-generated section
	EndMarker   string                             // End marker for auto-generated section
	Generator   func(*env.Registry) (string, error) // Function to generate content
}

// EnvironmentsSyncOptions configures the environments synchronization workflow
type EnvironmentsSyncOptions struct {
	Registry          *env.Registry     // The registry to validate against
	AppName           string            // Application name for headers
	LocalEnv          *env.Environment  // Local environment to sync
	ProductionEnv     *env.Environment  // Production environment to sync
	LocalSecrets      *env.Environment  // Local secrets file
	ProductionSecrets *env.Environment  // Production secrets file
	ValidateRequired  bool              // Whether to validate required variables
	OutputWriter      io.Writer         // Where to write progress messages (nil = discard)
}

// FinalizeOptions configures the finalization workflow (encryption + git)
type FinalizeOptions struct {
	Environments      []*env.Environment // Environments to encrypt
	EncryptionKeyPath string             // Path to age encryption key
	GitAdd            bool               // Whether to add encrypted files to git
	OutputWriter      io.Writer          // Where to write progress messages (nil = discard)
}

// ================================================================
// Result Structures
// ================================================================

// WorkflowResult contains structured results from workflow execution
type WorkflowResult struct {
	GeneratedFiles []string // Files that were created
	UpdatedFiles   []string // Files that were updated
	SkippedFiles   []string // Files that were skipped
	Warnings       []string // Non-fatal warnings
	Errors         []error  // Errors encountered (workflow may continue despite some errors)
}

// AddGenerated adds a file to the generated files list
func (r *WorkflowResult) AddGenerated(file string) {
	r.GeneratedFiles = append(r.GeneratedFiles, file)
}

// AddUpdated adds a file to the updated files list
func (r *WorkflowResult) AddUpdated(file string) {
	r.UpdatedFiles = append(r.UpdatedFiles, file)
}

// AddSkipped adds a file to the skipped files list
func (r *WorkflowResult) AddSkipped(file string) {
	r.SkippedFiles = append(r.SkippedFiles, file)
}

// AddWarning adds a warning message
func (r *WorkflowResult) AddWarning(msg string) {
	r.Warnings = append(r.Warnings, msg)
}

// AddError adds an error
func (r *WorkflowResult) AddError(err error) {
	if err != nil {
		r.Errors = append(r.Errors, err)
	}
}

// HasErrors returns true if any errors were encountered
func (r *WorkflowResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns true if any warnings were generated
func (r *WorkflowResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}
