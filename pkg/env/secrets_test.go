package env

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Test ParseSecretsFile with various formats
func TestParseSecretsFile(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]string
	}{
		{
			"simple key=value",
			"KEY1=value1\nKEY2=value2\n",
			map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			"ignores comments",
			"# Comment\nKEY1=value1\n# Another comment\nKEY2=value2",
			map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			"handles empty lines",
			"KEY1=value1\n\nKEY2=value2\n\n",
			map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			"handles values with =",
			"KEY1=value=with=equals\n",
			map[string]string{"KEY1": "value=with=equals"},
		},
		{
			"handles mixed content",
			"# Header comment\nKEY1=value1\n\n# Another section\nKEY2=value2\nKEY3=value3\n",
			map[string]string{"KEY1": "value1", "KEY2": "value2", "KEY3": "value3"},
		},
		{
			"empty file",
			"",
			map[string]string{},
		},
		{
			"only comments",
			"# Comment 1\n# Comment 2\n",
			map[string]string{},
		},
		{
			"trims whitespace",
			"  KEY1  =  value1  \n  KEY2=value2\n",
			map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			"empty value",
			"KEY1=\nKEY2=value2\n",
			map[string]string{"KEY1": "", "KEY2": "value2"},
		},
		{
			"multiline-like values (single line)",
			"KEY1=line1\\nline2\n",
			map[string]string{"KEY1": "line1\\nline2"},
		},
		{
			"special characters in value",
			"KEY1=!@#$%^&*()\nKEY2=value with spaces\n",
			map[string]string{"KEY1": "!@#$%^&*()", "KEY2": "value with spaces"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseSecretsFile([]byte(tt.input))

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSecretsFile() =\n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

// Test LoadSecrets with plaintext file
func TestLoadSecrets_Plaintext(t *testing.T) {
	// Create temp file with secrets
	tmpDir := t.TempDir()
	secretsPath := filepath.Join(tmpDir, ".env.secrets")

	content := "API_KEY=secret123\nDB_PASSWORD=pass456\n"
	if err := os.WriteFile(secretsPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	secrets, err := LoadSecrets(SecretsSource{
		FilePath:     secretsPath,
		PreferEncrypted: false,
	})

	if err != nil {
		t.Fatalf("LoadSecrets failed: %v", err)
	}

	if secrets["API_KEY"] != "secret123" {
		t.Errorf("API_KEY = %v, want secret123", secrets["API_KEY"])
	}
	if secrets["DB_PASSWORD"] != "pass456" {
		t.Errorf("DB_PASSWORD = %v, want pass456", secrets["DB_PASSWORD"])
	}
}

// Test LoadSecrets with nonexistent file
func TestLoadSecrets_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	secretsPath := filepath.Join(tmpDir, "nonexistent.secrets")

	_, err := LoadSecrets(SecretsSource{
		FilePath:     secretsPath,
		PreferEncrypted: false,
	})

	if err == nil {
		t.Error("Expected error for nonexistent file")
	}

	if !os.IsNotExist(err) && err.Error() == "" {
		t.Errorf("Expected file not found error, got: %v", err)
	}
}

// Test LoadSecrets prefers .age version when PreferEncrypted is true
func TestLoadSecrets_PreferEncrypted(t *testing.T) {
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, ".env.secrets")
	agePath := basePath + ".age"

	// Create both files - .age version should be preferred
	plaintextContent := "KEY=plaintext_value\n"
	// For this test, we create a plaintext .age file to verify file selection
	// (real encryption would require Age keys setup)
	ageContent := "KEY=from_age_file\n"

	os.WriteFile(basePath, []byte(plaintextContent), 0600)
	os.WriteFile(agePath, []byte(ageContent), 0600)

	// Note: This will try to decrypt the .age file and fail
	// But we can verify it attempted to use the .age file
	_, err := LoadSecrets(SecretsSource{
		FilePath:     basePath,
		PreferEncrypted: true,
	})

	// We expect an error because our fake .age file isn't actually encrypted
	// But the error should indicate it tried to decrypt
	if err == nil {
		t.Error("Expected decryption error for fake .age file")
	}

	// The important thing is it tried the .age file, not the plaintext
	// If we remove .age and try again, it should work
	os.Remove(agePath)

	secrets, err := LoadSecrets(SecretsSource{
		FilePath:     basePath,
		PreferEncrypted: true,
	})

	if err != nil {
		t.Fatalf("Failed to load plaintext when .age doesn't exist: %v", err)
	}

	if secrets["KEY"] != "plaintext_value" {
		t.Errorf("Got KEY=%s, expected plaintext_value", secrets["KEY"])
	}
}

// Test LoadSecrets with empty file
func TestLoadSecrets_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	secretsPath := filepath.Join(tmpDir, ".env.secrets")

	// Create empty file
	os.WriteFile(secretsPath, []byte(""), 0600)

	secrets, err := LoadSecrets(SecretsSource{
		FilePath:     secretsPath,
		PreferEncrypted: false,
	})

	if err != nil {
		t.Fatalf("LoadSecrets failed on empty file: %v", err)
	}

	if len(secrets) != 0 {
		t.Errorf("Expected empty map, got %d entries", len(secrets))
	}
}

// Test MergeIntoTemplate preserves structure
func TestMergeIntoTemplate(t *testing.T) {
	template := `# Comment line
API_KEY=placeholder
# Another comment
DB_PASSWORD=
OTHER_VAR=keep_this
`

	secrets := map[string]string{
		"API_KEY":     "secret123",
		"DB_PASSWORD": "pass456",
	}

	result := MergeIntoTemplate(template, secrets)

	// Should contain merged secrets
	if !contains(result, "API_KEY=secret123") {
		t.Error("Expected API_KEY=secret123")
	}
	if !contains(result, "DB_PASSWORD=pass456") {
		t.Error("Expected DB_PASSWORD=pass456")
	}
	// Should preserve OTHER_VAR
	if !contains(result, "OTHER_VAR=keep_this") {
		t.Error("Expected OTHER_VAR preserved")
	}
	// Should preserve comments
	if !contains(result, "# Comment line") {
		t.Error("Expected comments preserved")
	}
	if !contains(result, "# Another comment") {
		t.Error("Expected second comment preserved")
	}
}

// Test MergeIntoTemplate with complex template
func TestMergeIntoTemplate_Complex(t *testing.T) {
	template := `# Header section
# Multiple comment lines

# Group 1
VAR1=default1
VAR2=default2

# Group 2
VAR3=
VAR4=keep_this

# Footer comment
`

	secrets := map[string]string{
		"VAR1": "secret1",
		"VAR3": "secret3",
	}

	result := MergeIntoTemplate(template, secrets)

	// Check merged values
	if !contains(result, "VAR1=secret1") {
		t.Error("Expected VAR1=secret1")
	}
	if !contains(result, "VAR3=secret3") {
		t.Error("Expected VAR3=secret3")
	}

	// Check preserved values
	if !contains(result, "VAR2=default2") {
		t.Error("Expected VAR2 preserved")
	}
	if !contains(result, "VAR4=keep_this") {
		t.Error("Expected VAR4 preserved")
	}

	// Check structure preserved
	if !contains(result, "# Header section") {
		t.Error("Expected header comment preserved")
	}
	if !contains(result, "# Group 1") {
		t.Error("Expected group comment preserved")
	}
}

// Test MergeIntoTemplate with no secrets
func TestMergeIntoTemplate_NoSecrets(t *testing.T) {
	template := "VAR1=value1\nVAR2=value2\n"
	secrets := map[string]string{}

	result := MergeIntoTemplate(template, secrets)

	// Template should be unchanged
	if result != template+"\n" { // Extra newline from split/join
		t.Errorf("Expected template unchanged, got:\n%s", result)
	}
}

// Test MergeIntoTemplate handles commented assignments
func TestMergeIntoTemplate_CommentedAssignments(t *testing.T) {
	template := `# This is a comment
VAR1=value1
# VAR2=commented_out
VAR3=value3
`

	secrets := map[string]string{
		"VAR1": "secret1",
		"VAR2": "secret2", // This shouldn't replace the commented line
		"VAR3": "secret3",
	}

	result := MergeIntoTemplate(template, secrets)

	// VAR1 and VAR3 should be replaced
	if !contains(result, "VAR1=secret1") {
		t.Error("Expected VAR1=secret1")
	}
	if !contains(result, "VAR3=secret3") {
		t.Error("Expected VAR3=secret3")
	}

	// Commented line should remain commented
	if !contains(result, "# VAR2=commented_out") {
		t.Error("Expected commented VAR2 to remain unchanged")
	}
}

// Test MergeIntoTemplate with values containing special characters
func TestMergeIntoTemplate_SpecialChars(t *testing.T) {
	template := "API_KEY=\nDATABASE_URL=\n"

	secrets := map[string]string{
		"API_KEY":      "key_with_!@#$%",
		"DATABASE_URL": "postgres://user:pass@host:5432/db",
	}

	result := MergeIntoTemplate(template, secrets)

	if !contains(result, "API_KEY=key_with_!@#$%") {
		t.Error("Expected special characters preserved in API_KEY")
	}
	if !contains(result, "DATABASE_URL=postgres://user:pass@host:5432/db") {
		t.Error("Expected URL preserved in DATABASE_URL")
	}
}

// Test DecryptAgeFile error handling when no identities exist
func TestDecryptAgeFile_NoIdentities(t *testing.T) {
	// This test will fail if user has Age keys, so we skip in that case
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	// Check if any Age identity files exist
	agePaths := []string{
		filepath.Join(homeDir, ".ssh", "age"),
		filepath.Join(homeDir, ".config", "age", "keys.txt"),
	}

	hasKeys := false
	for _, path := range agePaths {
		if _, err := os.Stat(path); err == nil {
			hasKeys = true
			break
		}
	}

	// Also check AGE_IDENTITY env var
	if os.Getenv("AGE_IDENTITY") != "" {
		hasKeys = true
	}

	if hasKeys {
		t.Skip("Age keys exist, cannot test no-identity case")
	}

	// Try to decrypt some fake data
	fakeEncrypted := []byte("fake encrypted data")
	_, err = DecryptAgeFile(fakeEncrypted)

	if err == nil {
		t.Error("Expected error when no identities available")
	}

	// Error should mention creating keys
	if !contains(err.Error(), "age-keygen") {
		t.Errorf("Error should mention age-keygen, got: %v", err)
	}
}

// Test ParseSecretsFile handles malformed lines
func TestParseSecretsFile_Malformed(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]string
	}{
		{
			"no equals sign",
			"JUST_TEXT\nKEY=value\n",
			map[string]string{"KEY": "value"},
		},
		{
			"multiple equals signs",
			"KEY1=value1=extra\nKEY2=value2\n",
			map[string]string{"KEY1": "value1=extra", "KEY2": "value2"},
		},
		{
			"equals with no value",
			"KEY1=\nKEY2=value\n",
			map[string]string{"KEY1": "", "KEY2": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseSecretsFile([]byte(tt.input))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSecretsFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test LoadSecrets with read permission error
func TestLoadSecrets_ReadError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Cannot test permission errors as root")
	}

	tmpDir := t.TempDir()
	secretsPath := filepath.Join(tmpDir, ".env.secrets")

	// Create file and remove read permissions
	os.WriteFile(secretsPath, []byte("KEY=value"), 0000)
	defer os.Chmod(secretsPath, 0600) // Cleanup

	_, err := LoadSecrets(SecretsSource{
		FilePath:     secretsPath,
		PreferEncrypted: false,
	})

	if err == nil {
		t.Error("Expected error when file is not readable")
	}
}

// Test MergeIntoTemplate preserves exact formatting
func TestMergeIntoTemplate_PreservesFormatting(t *testing.T) {
	// Test that indentation, spacing, etc. are preserved
	template := `  # Indented comment
VAR1=value1
  VAR2=value2
    VAR3=value3
`

	secrets := map[string]string{
		"VAR1": "new1",
		"VAR2": "new2",
		"VAR3": "new3",
	}

	result := MergeIntoTemplate(template, secrets)

	// Check indentation is preserved
	if !contains(result, "  # Indented comment") {
		t.Error("Indentation in comment not preserved")
	}
	// Variable values should be replaced, formatting may change slightly
	// This is expected behavior - we replace the whole line
	if !contains(result, "VAR1=new1") {
		t.Error("VAR1 not updated")
	}
}
