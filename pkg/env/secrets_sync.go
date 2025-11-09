package env

import (
	"fmt"
	"os"
)

// SecretsSyncOptions configures secrets synchronization to environment files.
type SecretsSyncOptions struct {
	Registry    *Registry    // Registry containing environment variable definitions
	TargetEnv   *Environment // Target environment file to write (.env.local or .env.production)
	SecretsEnv  *Environment // Secrets environment to read from (or nil to auto-resolve)
	AppName     string       // Application name for template header
	AutoResolve bool         // If true and SecretsEnv is nil, use ResolveSecretsFile()
}

// SecretsSyncResult contains the result of secrets synchronization.
type SecretsSyncResult struct {
	TargetFile    string // Target file that was written
	SecretsFile   string // Secrets file that was read
	SecretsCount  int    // Number of secrets merged
	UsedFallback  bool   // Whether fallback secrets file was used
	FallbackFile  string // Fallback file that was used (if any)
}

// SyncSecretsToEnvironment merges secrets from a secrets file into an environment template.
//
// This function:
//  1. Resolves which secrets file to use (with fallback logic if AutoResolve is true)
//  2. Loads secrets from the file (prefers encrypted .age versions)
//  3. Generates environment template from registry
//  4. Merges secrets into template
//  5. Writes merged content to target environment file
//
// Example:
//
//	result, err := env.SyncSecretsToEnvironment(env.SecretsSyncOptions{
//	    Registry:    AppRegistry,
//	    TargetEnv:   env.Local,
//	    AppName:     "My Application",
//	    AutoResolve: true, // Will use ResolveSecretsFile() to find best secrets source
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Synced %d secrets to %s\n", result.SecretsCount, result.TargetFile)
func SyncSecretsToEnvironment(opts SecretsSyncOptions) (*SecretsSyncResult, error) {
	result := &SecretsSyncResult{
		TargetFile: opts.TargetEnv.FileName,
	}

	// Set defaults
	if opts.AppName == "" {
		opts.AppName = "Application"
	}

	// Resolve secrets file if needed
	secretsEnv := opts.SecretsEnv
	if secretsEnv == nil && opts.AutoResolve {
		var usedFallback bool
		secretsEnv, usedFallback = ResolveSecretsFile(opts.TargetEnv)
		if secretsEnv == nil {
			return nil, fmt.Errorf("no secrets file found for %s", opts.TargetEnv.FileName)
		}
		result.UsedFallback = usedFallback
		if usedFallback {
			result.FallbackFile = secretsEnv.FileName
		}
	}

	if secretsEnv == nil {
		return nil, fmt.Errorf("no secrets environment specified (set SecretsEnv or enable AutoResolve)")
	}

	result.SecretsFile = secretsEnv.FileName

	// Load secrets (with automatic .age detection and decryption)
	secrets, err := LoadSecrets(SecretsSource{
		FilePath:        secretsEnv.FileName,
		PreferEncrypted: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load secrets from %s: %w", secretsEnv.FileName, err)
	}

	result.SecretsCount = len(secrets)

	// Generate template and merge secrets
	template := opts.TargetEnv.Generate(opts.Registry, opts.AppName)
	mergedContent := MergeIntoTemplate(template, secrets)

	// Write merged content to target environment file
	if err := os.WriteFile(opts.TargetEnv.FileName, []byte(mergedContent), 0600); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", opts.TargetEnv.FileName, err)
	}

	return result, nil
}
