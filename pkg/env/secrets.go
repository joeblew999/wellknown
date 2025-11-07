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
	FilePath     string // Path to secrets file (e.g., ".env.secrets")
	TryEncrypted bool   // Try .age version first before plaintext
}

// LoadSecrets loads and optionally decrypts secrets from a file.
//
// Behavior:
//  1. If TryEncrypted is true and FilePath.age exists, use that (decrypt)
//  2. Otherwise use FilePath directly
//  3. Parse as key=value format
//  4. Return map of secrets
//
// Example:
//
//	secrets, err := LoadSecrets(SecretsSource{
//	  FilePath: ".env.secrets",
//	  TryEncrypted: true,
//	})
//	// Will try .env.secrets.age first, then .env.secrets
func LoadSecrets(src SecretsSource) (map[string]string, error) {
	// Determine which file to load
	actualPath := src.FilePath
	needsDecryption := false

	if src.TryEncrypted {
		ageVersion := src.FilePath + ".age"
		if _, err := os.Stat(ageVersion); err == nil {
			actualPath = ageVersion
			needsDecryption = true
			fmt.Printf("Using encrypted secrets from %s\n", ageVersion)
		}
	}

	// Check if file exists
	if _, err := os.Stat(actualPath); os.IsNotExist(err) {
		if src.TryEncrypted {
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
