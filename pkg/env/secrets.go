// Package env provides Age encryption utilities for secure secrets management.
package env

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
)

// DecryptAgeFile decrypts an Age-encrypted file using identities from standard locations.
// It looks for Age identities in:
//  1. AGE_IDENTITY environment variable (path to identity file) - highest priority
//  2. ~/.ssh/age (SSH-style Age key)
//  3. ~/.config/age/keys.txt (Age native keys)
//
// Returns the decrypted data or an error with helpful guidance.
func DecryptAgeFile(encryptedData []byte) ([]byte, error) {
	// Find identity files
	var identities []age.Identity
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Try standard Age identity locations
	identityPaths := []string{
		filepath.Join(homeDir, ".ssh", "age"),
		filepath.Join(homeDir, ".config", "age", "keys.txt"),
	}

	// Check for AGE_IDENTITY environment variable (highest priority)
	if envIdentity := os.Getenv("AGE_IDENTITY"); envIdentity != "" {
		identityPaths = append([]string{envIdentity}, identityPaths...)
	}

	// Load identities from files
	for _, path := range identityPaths {
		if _, err := os.Stat(path); err == nil {
			identityFile, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			parsedIdentities, err := age.ParseIdentities(bytes.NewReader(identityFile))
			if err != nil {
				continue
			}

			identities = append(identities, parsedIdentities...)
		}
	}

	if len(identities) == 0 {
		return nil, fmt.Errorf("no Age identities found. Create one with:\n  age-keygen -o ~/.ssh/age\n\nOr set AGE_IDENTITY environment variable to your identity file path")
	}

	// Decrypt the file
	r, err := age.Decrypt(bytes.NewReader(encryptedData), identities...)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w\n\nMake sure you have the correct Age identity key", err)
	}

	decrypted, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read decrypted data: %w", err)
	}

	return decrypted, nil
}

// ParseSecretsFile parses a key=value formatted secrets file (like .env.secrets).
// It ignores comments (lines starting with #) and empty lines.
// Returns a map of environment variable names to their values.
func ParseSecretsFile(data []byte) map[string]string {
	secrets := make(map[string]string)
	lines := bytes.Split(data, []byte("\n"))

	for _, line := range lines {
		// Trim whitespace
		lineStr := string(bytes.TrimSpace(line))

		// Skip empty lines and comments
		if lineStr == "" || lineStr[0] == '#' {
			continue
		}

		// Split on first '=' only
		parts := bytes.SplitN(line, []byte("="), 2)
		if len(parts) == 2 {
			key := string(bytes.TrimSpace(parts[0]))
			value := string(bytes.TrimSpace(parts[1]))
			secrets[key] = value
		}
	}

	return secrets
}

// SecretsSource specifies where to load secrets from.
// This supports automatic fallback from encrypted to plaintext versions.
type SecretsSource struct {
	FilePath        string // Path to secrets file (e.g., ".env.secrets")
	PreferEncrypted bool   // Prefer .age version first before plaintext
}

// LoadSecrets loads and optionally decrypts secrets from a file.
//
// Behavior:
//  1. If PreferEncrypted is true and FilePath.age exists, use that (decrypt)
//  2. Otherwise use FilePath directly
//  3. Parse as key=value format
//  4. Return map of secrets
//
// Example:
//
//	secrets, err := LoadSecrets(SecretsSource{
//	  FilePath: ".env.secrets",
//	  PreferEncrypted: true,
//	})
//	// Will try .env.secrets.age first, then .env.secrets
func LoadSecrets(src SecretsSource) (map[string]string, error) {
	// Determine which file to load
	actualPath := src.FilePath
	needsDecryption := false

	if src.PreferEncrypted {
		ageVersion := src.FilePath + ".age"
		if _, err := os.Stat(ageVersion); err == nil {
			actualPath = ageVersion
			needsDecryption = true
			fmt.Printf("Using encrypted secrets from %s\n", ageVersion)
		}
	}

	// Check if file exists
	if _, err := os.Stat(actualPath); os.IsNotExist(err) {
		if src.PreferEncrypted {
			return nil, fmt.Errorf("secrets file not found: %s or %s.age\n\nPlease create it from .env.secrets.example\nOptional: Encrypt with Age:\n  age -e -r YOUR_PUBLIC_KEY %s > %s.age",
				src.FilePath, src.FilePath, src.FilePath, src.FilePath)
		}
		return nil, fmt.Errorf("secrets file not found: %s", actualPath)
	}

	// Read file
	data, err := os.ReadFile(actualPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets file %s: %w", actualPath, err)
	}

	// Decrypt if needed
	if needsDecryption {
		decrypted, err := DecryptAgeFile(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt Age file: %w", err)
		}
		data = decrypted
	}

	// Parse and return
	return ParseSecretsFile(data), nil
}

// MergeIntoTemplate merges secrets map into a template string.
//
// This preserves the template structure (comments, headers, blank lines)
// while replacing variable values where secrets exist.
//
// Template format:
//
//	# Comment lines are preserved
//	KEY_NAME=template_value
//
// If secrets map contains KEY_NAME, the line becomes:
//
//	KEY_NAME=secret_value
//
// Otherwise the template line is kept as-is.
func MergeIntoTemplate(template string, secrets map[string]string) string {
	var sb strings.Builder
	lines := strings.Split(template, "\n")

	for _, line := range lines {
		// Check if this is a variable assignment (not a comment or empty line)
		trimmedLine := strings.TrimSpace(line)
		if strings.Contains(line, "=") && !strings.HasPrefix(trimmedLine, "#") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])

				// If we have a secret value for this key, use it
				if secretValue, exists := secrets[key]; exists {
					sb.WriteString(fmt.Sprintf("%s=%s\n", key, secretValue))
					continue
				}
			}
		}

		// Otherwise, keep the line as-is (comments, headers, empty lines, or vars without secrets)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return sb.String()
}

// ================================================================
// Age Key Generation
// ================================================================

// KeygenOptions configures Age key generation.
type KeygenOptions struct {
	KeyPath         string      // Path to save the key (e.g., ".age/key.txt")
	OverwritePrompt func() bool // Optional: callback to prompt for overwrite confirmation
}

// KeygenResult contains the result of key generation.
type KeygenResult struct {
	KeyPath   string // Path where the key was saved
	PublicKey string // Public key string (for sharing)
	Created   bool   // Whether a new key was created (false if aborted)
}

// GenerateAgeKey generates a new Age encryption key pair.
//
// This function:
//  1. Checks if key already exists
//  2. Optionally prompts for overwrite confirmation
//  3. Creates the key directory if needed
//  4. Generates an X25519 identity
//  5. Writes the identity to the key file
//
// Example:
//
//	result, err := env.GenerateAgeKey(env.KeygenOptions{
//	    KeyPath: ".age/key.txt",
//	    OverwritePrompt: func() bool {
//	        fmt.Print("Overwrite existing key? (y/N): ")
//	        var response string
//	        fmt.Scanln(&response)
//	        return response == "y" || response == "Y"
//	    },
//	})
func GenerateAgeKey(opts KeygenOptions) (*KeygenResult, error) {
	// Set defaults
	if opts.KeyPath == "" {
		opts.KeyPath = DefaultAgeKeyPath
	}

	// Check if key already exists
	if _, err := os.Stat(opts.KeyPath); err == nil {
		// Key exists - check if we should overwrite
		if opts.OverwritePrompt != nil {
			if !opts.OverwritePrompt() {
				return &KeygenResult{
					KeyPath: opts.KeyPath,
					Created: false,
				}, nil
			}
		} else {
			// No prompt provided - don't overwrite
			return nil, fmt.Errorf("key already exists at %s", opts.KeyPath)
		}
	}

	// Create directory if it doesn't exist
	keyDir := filepath.Dir(opts.KeyPath)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", keyDir, err)
	}

	// Generate new identity
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("failed to generate identity: %w", err)
	}

	// Format identity file with metadata
	identityStr := fmt.Sprintf("# created: %s\n# public key: %s\n%s\n",
		identity.Recipient().String(),
		identity.Recipient().String(),
		identity.String())

	// Write identity to file with restricted permissions
	if err := os.WriteFile(opts.KeyPath, []byte(identityStr), 0600); err != nil {
		return nil, fmt.Errorf("failed to write key to %s: %w", opts.KeyPath, err)
	}

	return &KeygenResult{
		KeyPath:   opts.KeyPath,
		PublicKey: identity.Recipient().String(),
		Created:   true,
	}, nil
}

// ================================================================
// Batch Encryption/Decryption
// ================================================================

// EncryptionOptions configures batch encryption or decryption.
type EncryptionOptions struct {
	KeyPath      string         // Path to Age identity file
	Environments []*Environment // Environments to encrypt/decrypt
}

// EncryptionResult contains the result of batch encryption/decryption.
type EncryptionResult struct {
	ProcessedFiles []string // Files that were successfully processed
	SkippedFiles   []string // Files that were skipped (didn't exist)
	Errors         []error  // Errors encountered during processing
}

// HasErrors returns true if any errors occurred.
func (r *EncryptionResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// EncryptEnvironments encrypts multiple environment files using Age encryption.
//
// This function:
//  1. Loads the Age identity from KeyPath
//  2. For each environment file that exists:
//     - Reads the plaintext content
//     - Encrypts it with Age
//     - Writes to .age file
//  3. Returns structured results
//
// Example:
//
//	result, err := env.EncryptEnvironments(env.EncryptionOptions{
//	    KeyPath:      ".age/key.txt",
//	    Environments: env.AllEnvironmentFiles(),
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Encrypted %d files\n", len(result.ProcessedFiles))
func EncryptEnvironments(opts EncryptionOptions) (*EncryptionResult, error) {
	result := &EncryptionResult{}

	// Set defaults
	if opts.KeyPath == "" {
		opts.KeyPath = DefaultAgeKeyPath
	}
	if opts.Environments == nil {
		opts.Environments = AllEnvironmentFiles()
	}

	// Check if key exists
	if _, err := os.Stat(opts.KeyPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no Age key found at %s. Generate one with GenerateAgeKey()", opts.KeyPath)
	}

	// Read and parse identity
	identityFile, err := os.ReadFile(opts.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key from %s: %w", opts.KeyPath, err)
	}

	identities, err := age.ParseIdentities(bytes.NewReader(identityFile))
	if err != nil || len(identities) == 0 {
		return nil, fmt.Errorf("failed to parse identity from %s: %w", opts.KeyPath, err)
	}

	// Get recipient (public key) from identity
	recipient := identities[0].(*age.X25519Identity).Recipient()

	// Encrypt each environment file
	for _, envFile := range opts.Environments {
		// Skip if plaintext doesn't exist
		if !envFile.Exists() {
			result.SkippedFiles = append(result.SkippedFiles, envFile.FileName)
			continue
		}

		// Read plaintext
		plaintext, err := os.ReadFile(envFile.FullPath())
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to read %s: %w", envFile.FileName, err))
			continue
		}

		// Encrypt
		var buf bytes.Buffer
		w, err := age.Encrypt(&buf, recipient)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to encrypt %s: %w", envFile.FileName, err))
			continue
		}

		if _, err := w.Write(plaintext); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to write encrypted %s: %w", envFile.FileName, err))
			continue
		}

		if err := w.Close(); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to finalize %s: %w", envFile.FileName, err))
			continue
		}

		// Write encrypted file
		encryptedPath := envFile.FullEncryptedPath()
		if err := os.WriteFile(encryptedPath, buf.Bytes(), 0600); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to write %s: %w", envFile.EncryptedFileName(), err))
			continue
		}

		result.ProcessedFiles = append(result.ProcessedFiles, envFile.EncryptedFileName())
	}

	// If nothing was processed and we have errors, return error
	if len(result.ProcessedFiles) == 0 && len(result.Errors) > 0 {
		return result, fmt.Errorf("failed to encrypt any files: %v", result.Errors[0])
	}

	return result, nil
}

// DecryptEnvironments decrypts multiple encrypted environment files.
//
// This function:
//  1. Sets AGE_IDENTITY environment variable
//  2. For each .age file that exists:
//     - Decrypts using DecryptAgeFile()
//     - Writes plaintext to environment file
//  3. Returns structured results
//
// Example:
//
//	result, err := env.DecryptEnvironments(env.EncryptionOptions{
//	    KeyPath:      ".age/key.txt",
//	    Environments: env.AllEnvironmentFiles(),
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Decrypted %d files\n", len(result.ProcessedFiles))
func DecryptEnvironments(opts EncryptionOptions) (*EncryptionResult, error) {
	result := &EncryptionResult{}

	// Set defaults
	if opts.KeyPath == "" {
		opts.KeyPath = DefaultAgeKeyPath
	}
	if opts.Environments == nil {
		opts.Environments = AllEnvironmentFiles()
	}

	// Set AGE_IDENTITY for DecryptAgeFile
	os.Setenv("AGE_IDENTITY", opts.KeyPath)

	// Decrypt each environment file
	for _, envFile := range opts.Environments {
		encryptedPath := envFile.FullEncryptedPath()

		// Skip if encrypted file doesn't exist
		if _, err := os.Stat(encryptedPath); os.IsNotExist(err) {
			result.SkippedFiles = append(result.SkippedFiles, envFile.EncryptedFileName())
			continue
		}

		// Read encrypted file
		encrypted, err := os.ReadFile(encryptedPath)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to read %s: %w", envFile.EncryptedFileName(), err))
			continue
		}

		// Decrypt
		decrypted, err := DecryptAgeFile(encrypted)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to decrypt %s: %w", envFile.EncryptedFileName(), err))
			continue
		}

		// Write plaintext
		if err := os.WriteFile(envFile.FullPath(), decrypted, 0600); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to write %s: %w", envFile.FileName, err))
			continue
		}

		result.ProcessedFiles = append(result.ProcessedFiles, envFile.FileName)
	}

	// If nothing was processed and we have errors, return error
	if len(result.ProcessedFiles) == 0 && len(result.Errors) > 0 {
		return result, fmt.Errorf("failed to decrypt any files: %v", result.Errors[0])
	}

	return result, nil
}
