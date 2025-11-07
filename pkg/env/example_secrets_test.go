package env_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joeblew999/wellknown/pkg/env"
)

// ExampleParseSecretsFile demonstrates parsing key=value format
func ExampleParseSecretsFile() {
	content := `# Database credentials
DB_HOST=localhost
DB_PASSWORD=secret123

# API keys
API_KEY=abc123xyz
`

	secrets := env.ParseSecretsFile([]byte(content))

	fmt.Printf("DB_HOST: %s\n", secrets["DB_HOST"])
	fmt.Printf("DB_PASSWORD: %s\n", secrets["DB_PASSWORD"])
	fmt.Printf("API_KEY: %s\n", secrets["API_KEY"])

	// Output:
	// DB_HOST: localhost
	// DB_PASSWORD: secret123
	// API_KEY: abc123xyz
}

// ExampleLoadSecrets shows loading secrets from a file
func ExampleLoadSecrets() {
	// Create temporary secrets file
	tmpDir := os.TempDir()
	secretsFile := filepath.Join(tmpDir, "example.secrets")

	content := "API_KEY=secret123\nDB_PASSWORD=pass456\n"
	os.WriteFile(secretsFile, []byte(content), 0600)
	defer os.Remove(secretsFile)

	// Load secrets
	secrets, err := env.LoadSecrets(env.SecretsSource{
		FilePath:     secretsFile,
		TryEncrypted: false,
	})

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Loaded %d secrets\n", len(secrets))
	fmt.Printf("API_KEY: %s\n", secrets["API_KEY"])

	// Output:
	// Loaded 2 secrets
	// API_KEY: secret123
}

// ExampleLoadSecrets_encrypted demonstrates encrypted secrets handling
func ExampleLoadSecrets_encrypted() {
	// Create temporary plaintext file (in real use, you'd have .age file)
	tmpDir := os.TempDir()
	secretsFile := filepath.Join(tmpDir, "example2.secrets")

	content := "SECRET_KEY=encrypted_value\n"
	os.WriteFile(secretsFile, []byte(content), 0600)
	defer os.Remove(secretsFile)

	// TryEncrypted will look for .age version first
	secrets, err := env.LoadSecrets(env.SecretsSource{
		FilePath:     secretsFile,
		TryEncrypted: true,
	})

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Loaded %d secrets\n", len(secrets))

	// Output:
	// Loaded 1 secrets
}

// ExampleMergeIntoTemplate shows merging secrets into a template
func ExampleMergeIntoTemplate() {
	template := `# Configuration file
# Database settings
DB_HOST=localhost
DB_PASSWORD=

# API settings
API_KEY=placeholder
`

	secrets := map[string]string{
		"DB_PASSWORD": "secret_password",
		"API_KEY":     "real_api_key",
	}

	result := env.MergeIntoTemplate(template, secrets)
	fmt.Print(result)

	// Output:
	// # Configuration file
	// # Database settings
	// DB_HOST=localhost
	// DB_PASSWORD=secret_password
	//
	// # API settings
	// API_KEY=real_api_key
}

// ExampleMergeIntoTemplate_preserveStructure shows structure preservation
func ExampleMergeIntoTemplate_preserveStructure() {
	template := `# Header comment

# Section 1
VAR1=default1

# Section 2
VAR2=default2
VAR3=keep_this
`

	// Only provide VAR1, others should be preserved
	secrets := map[string]string{
		"VAR1": "new_value",
	}

	result := env.MergeIntoTemplate(template, secrets)
	fmt.Print(result)

	// Output:
	// # Header comment
	//
	// # Section 1
	// VAR1=new_value
	//
	// # Section 2
	// VAR2=default2
	// VAR3=keep_this
}

// ExampleParseSecretsFile_comments shows comment handling
func ExampleParseSecretsFile_comments() {
	content := `# This is a comment
KEY1=value1
# Another comment
KEY2=value2

# Empty lines are ignored
KEY3=value3
`

	secrets := env.ParseSecretsFile([]byte(content))

	fmt.Printf("Parsed %d keys\n", len(secrets))
	for key := range secrets {
		fmt.Printf("- %s\n", key)
	}

	// Unordered output:
	// Parsed 3 keys
	// - KEY1
	// - KEY2
	// - KEY3
}

// ExampleParseSecretsFile_specialChars shows special character handling
func ExampleParseSecretsFile_specialChars() {
	content := `URL=https://example.com/path?key=value&other=123
PASSWORD=p@ssw0rd!with#special$chars
BASE64=SGVsbG8gV29ybGQ=
`

	secrets := env.ParseSecretsFile([]byte(content))

	fmt.Println(secrets["URL"])
	fmt.Println(secrets["PASSWORD"])
	fmt.Println(secrets["BASE64"])

	// Output:
	// https://example.com/path?key=value&other=123
	// p@ssw0rd!with#special$chars
	// SGVsbG8gV29ybGQ=
}

// ExampleSecretsSource shows source configuration
func ExampleSecretsSource() {
	// Example of configuring secrets source
	source := env.SecretsSource{
		FilePath:     ".env.secrets",
		TryEncrypted: true,
	}

	fmt.Printf("File path: %s\n", source.FilePath)
	fmt.Printf("Try encrypted: %v\n", source.TryEncrypted)

	// This will:
	// 1. First try .env.secrets.age (if TryEncrypted is true)
	// 2. Fall back to .env.secrets if .age doesn't exist
	// 3. Decrypt automatically if Age keys are available

	// Output:
	// File path: .env.secrets
	// Try encrypted: true
}

// ExampleMergeIntoTemplate_emptyValues shows empty value handling
func ExampleMergeIntoTemplate_emptyValues() {
	template := `KEY1=
KEY2=default
KEY3=
`

	secrets := map[string]string{
		"KEY1": "now_filled",
		"KEY3": "also_filled",
	}

	result := env.MergeIntoTemplate(template, secrets)
	fmt.Print(result)

	// Output:
	// KEY1=now_filled
	// KEY2=default
	// KEY3=also_filled
}
