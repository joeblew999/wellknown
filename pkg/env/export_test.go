package env

import (
	"os"
	"strings"
	"testing"
)

func TestRegistry_Export(t *testing.T) {
	vars := []EnvVar{
		{Name: "PUBLIC_VAR", Description: "Public", Default: "public_value", Secret: false},
		{Name: "SECRET_VAR", Description: "Secret", Default: "secret_value", Secret: true, Required: true},
		{Name: "OPTIONAL_SECRET", Description: "Optional secret", Secret: true, Required: false},
		{Name: "REQUIRED_PUBLIC", Description: "Required public", Secret: false, Required: true},
	}
	registry := NewRegistry(vars)

	// Set some environment variables for testing
	os.Setenv("PUBLIC_VAR", "public_value")
	os.Setenv("SECRET_VAR", "secret_value")
	os.Setenv("OPTIONAL_SECRET", "optional_secret_value")
	os.Setenv("REQUIRED_PUBLIC", "required_public_value")
	defer func() {
		os.Unsetenv("PUBLIC_VAR")
		os.Unsetenv("SECRET_VAR")
		os.Unsetenv("OPTIONAL_SECRET")
		os.Unsetenv("REQUIRED_PUBLIC")
	}()

	tests := []struct {
		name       string
		opts       ExportOptions
		wantAll    []string // All strings that should be in output
		wantNone   []string // Strings that should NOT be in output
		wantEmpty  bool     // Should output be empty
	}{
		{
			"simple format all vars",
			ExportOptions{Format: FormatSimple, IncludeEmpty: false},
			[]string{"PUBLIC_VAR=public_value", "SECRET_VAR=secret_value", "OPTIONAL_SECRET=optional_secret_value"},
			[]string{},
			false,
		},
		{
			"secrets only",
			ExportOptions{Format: FormatSimple, SecretsOnly: true, IncludeEmpty: false},
			[]string{"SECRET_VAR=secret_value", "OPTIONAL_SECRET=optional_secret_value"},
			[]string{"PUBLIC_VAR=", "REQUIRED_PUBLIC="},
			false,
		},
		{
			"required only",
			ExportOptions{Format: FormatSimple, RequiredOnly: true, IncludeEmpty: false},
			[]string{"SECRET_VAR=secret_value", "REQUIRED_PUBLIC=required_public_value"},
			[]string{"OPTIONAL_SECRET="},
			false,
		},
		{
			"mask secrets",
			ExportOptions{Format: FormatSimple, MaskSecrets: true, IncludeEmpty: false},
			[]string{"PUBLIC_VAR=public_value", "SECRET_VAR=***", "OPTIONAL_SECRET=***"},
			[]string{"secret_value", "optional_secret_value"},
			false,
		},
		{
			"docker format",
			ExportOptions{Format: FormatDocker, IncludeEmpty: false},
			[]string{"PUBLIC_VAR=public_value", "SECRET_VAR=secret_value"},
			[]string{},
			false,
		},
		{
			"systemd format",
			ExportOptions{Format: FormatSystemd, IncludeEmpty: false},
			[]string{"Environment=\"PUBLIC_VAR=public_value\"", "Environment=\"SECRET_VAR=secret_value\""},
			[]string{},
			false,
		},
		{
			"k8s format",
			ExportOptions{Format: FormatK8s, IncludeEmpty: false},
			[]string{"- name: PUBLIC_VAR", "  value: \"public_value\"", "- name: SECRET_VAR", "  value: \"secret_value\""},
			[]string{},
			false,
		},
		{
			"include empty values",
			ExportOptions{Format: FormatSimple, IncludeEmpty: true, SecretsOnly: true},
			[]string{"SECRET_VAR=secret_value", "OPTIONAL_SECRET=optional_secret_value"},
			[]string{},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.Export(tt.opts)

			if tt.wantEmpty {
				if got != "" {
					t.Errorf("Expected empty output, got: %s", got)
				}
				return
			}

			for _, want := range tt.wantAll {
				if !strings.Contains(got, want) {
					t.Errorf("Export missing %q\nGot:\n%s", want, got)
				}
			}

			for _, unwanted := range tt.wantNone {
				if strings.Contains(got, unwanted) {
					t.Errorf("Export should not contain %q\nGot:\n%s", unwanted, got)
				}
			}
		})
	}
}

// Test export with empty environment (no vars set)
func TestRegistry_Export_EmptyEnvironment(t *testing.T) {
	vars := []EnvVar{
		{Name: "UNSET_VAR1", Secret: false},
		{Name: "UNSET_VAR2", Secret: true},
	}
	registry := NewRegistry(vars)

	// Make sure vars are not set
	os.Unsetenv("UNSET_VAR1")
	os.Unsetenv("UNSET_VAR2")

	// Without IncludeEmpty, output should be empty
	output := registry.Export(ExportOptions{
		Format:       FormatSimple,
		IncludeEmpty: false,
	})

	if output != "" {
		t.Errorf("Expected empty output when no vars set, got: %s", output)
	}

	// With IncludeEmpty, should include vars with empty values
	output = registry.Export(ExportOptions{
		Format:       FormatSimple,
		IncludeEmpty: true,
	})

	if !strings.Contains(output, "UNSET_VAR1=") {
		t.Error("Expected UNSET_VAR1 with empty value")
	}
}

// Test ExportSimple convenience method
func TestRegistry_ExportSimple(t *testing.T) {
	vars := []EnvVar{
		{Name: "VAR1"},
		{Name: "VAR2"},
	}
	registry := NewRegistry(vars)

	os.Setenv("VAR1", "value1")
	os.Setenv("VAR2", "value2")
	defer func() {
		os.Unsetenv("VAR1")
		os.Unsetenv("VAR2")
	}()

	output := registry.ExportSimple()

	if !strings.Contains(output, "VAR1=value1") {
		t.Error("ExportSimple missing VAR1")
	}
	if !strings.Contains(output, "VAR2=value2") {
		t.Error("ExportSimple missing VAR2")
	}
}

// Test ExportSecrets convenience method
func TestRegistry_ExportSecrets(t *testing.T) {
	vars := []EnvVar{
		{Name: "PUBLIC", Secret: false},
		{Name: "SECRET1", Secret: true},
		{Name: "SECRET2", Secret: true},
	}
	registry := NewRegistry(vars)

	os.Setenv("PUBLIC", "public_value")
	os.Setenv("SECRET1", "secret1_value")
	os.Setenv("SECRET2", "secret2_value")
	defer func() {
		os.Unsetenv("PUBLIC")
		os.Unsetenv("SECRET1")
		os.Unsetenv("SECRET2")
	}()

	output := registry.ExportSecrets()

	// Should only have secrets
	if !strings.Contains(output, "SECRET1=secret1_value") {
		t.Error("ExportSecrets missing SECRET1")
	}
	if !strings.Contains(output, "SECRET2=secret2_value") {
		t.Error("ExportSecrets missing SECRET2")
	}
	// Should not have public vars
	if strings.Contains(output, "PUBLIC=") {
		t.Error("ExportSecrets should not include public vars")
	}
}

// Test ExportRequired convenience method
func TestRegistry_ExportRequired(t *testing.T) {
	vars := []EnvVar{
		{Name: "OPTIONAL", Required: false},
		{Name: "REQUIRED1", Required: true},
		{Name: "REQUIRED2", Required: true},
	}
	registry := NewRegistry(vars)

	os.Setenv("OPTIONAL", "optional_value")
	os.Setenv("REQUIRED1", "required1_value")
	// REQUIRED2 intentionally not set to test IncludeEmpty
	defer func() {
		os.Unsetenv("OPTIONAL")
		os.Unsetenv("REQUIRED1")
	}()

	output := registry.ExportRequired()

	// Should only have required vars
	if !strings.Contains(output, "REQUIRED1=required1_value") {
		t.Error("ExportRequired missing REQUIRED1")
	}
	// Should include empty required vars (to show what's missing)
	if !strings.Contains(output, "REQUIRED2=") {
		t.Error("ExportRequired should include empty REQUIRED2")
	}
	// Should not have optional vars
	if strings.Contains(output, "OPTIONAL=") {
		t.Error("ExportRequired should not include optional vars")
	}
}

// Test ExportSystemd convenience method
func TestRegistry_ExportSystemd(t *testing.T) {
	vars := []EnvVar{
		{Name: "VAR1"},
		{Name: "VAR2"},
	}
	registry := NewRegistry(vars)

	os.Setenv("VAR1", "value1")
	os.Setenv("VAR2", "value2")
	defer func() {
		os.Unsetenv("VAR1")
		os.Unsetenv("VAR2")
	}()

	output := registry.ExportSystemd()

	if !strings.Contains(output, "Environment=\"VAR1=value1\"") {
		t.Error("ExportSystemd format incorrect for VAR1")
	}
	if !strings.Contains(output, "Environment=\"VAR2=value2\"") {
		t.Error("ExportSystemd format incorrect for VAR2")
	}
}

// Test ExportK8s convenience method
func TestRegistry_ExportK8s(t *testing.T) {
	vars := []EnvVar{
		{Name: "VAR1"},
		{Name: "VAR2"},
	}
	registry := NewRegistry(vars)

	os.Setenv("VAR1", "value1")
	os.Setenv("VAR2", "value2")
	defer func() {
		os.Unsetenv("VAR1")
		os.Unsetenv("VAR2")
	}()

	output := registry.ExportK8s()

	// Check K8s YAML format
	if !strings.Contains(output, "- name: VAR1") {
		t.Error("ExportK8s missing VAR1 name")
	}
	if !strings.Contains(output, "  value: \"value1\"") {
		t.Error("ExportK8s missing VAR1 value")
	}
	if !strings.Contains(output, "- name: VAR2") {
		t.Error("ExportK8s missing VAR2 name")
	}
	if !strings.Contains(output, "  value: \"value2\"") {
		t.Error("ExportK8s missing VAR2 value")
	}
}

// Test format variations
func TestExport_Formats(t *testing.T) {
	vars := []EnvVar{
		{Name: "TEST_VAR"},
	}
	registry := NewRegistry(vars)

	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	tests := []struct {
		format ExportFormat
		want   string
	}{
		{FormatSimple, "TEST_VAR=test_value"},
		{FormatDocker, "TEST_VAR=test_value"},
		{FormatSystemd, "Environment=\"TEST_VAR=test_value\""},
		{FormatK8s, "- name: TEST_VAR"},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			output := registry.Export(ExportOptions{
				Format:       tt.format,
				IncludeEmpty: false,
			})

			if !strings.Contains(output, tt.want) {
				t.Errorf("Format %s missing %q\nGot: %s", tt.format, tt.want, output)
			}
		})
	}
}

// Test masking with empty secrets
func TestRegistry_Export_MaskEmptySecrets(t *testing.T) {
	vars := []EnvVar{
		{Name: "EMPTY_SECRET", Secret: true},
		{Name: "SET_SECRET", Secret: true},
	}
	registry := NewRegistry(vars)

	os.Setenv("SET_SECRET", "secret_value")
	os.Unsetenv("EMPTY_SECRET")
	defer os.Unsetenv("SET_SECRET")

	output := registry.Export(ExportOptions{
		Format:       FormatSimple,
		MaskSecrets:  true,
		IncludeEmpty: true,
	})

	// Empty secrets should remain empty, not masked
	if strings.Contains(output, "EMPTY_SECRET=***") {
		t.Error("Empty secrets should not be masked")
	}

	// Set secrets should be masked
	if !strings.Contains(output, "SET_SECRET=***") {
		t.Error("Set secrets should be masked")
	}
	if strings.Contains(output, "secret_value") {
		t.Error("Secret value should not appear in output when masked")
	}
}

// Test combined filters
func TestRegistry_Export_CombinedFilters(t *testing.T) {
	vars := []EnvVar{
		{Name: "PUBLIC_REQUIRED", Secret: false, Required: true},
		{Name: "PUBLIC_OPTIONAL", Secret: false, Required: false},
		{Name: "SECRET_REQUIRED", Secret: true, Required: true},
		{Name: "SECRET_OPTIONAL", Secret: true, Required: false},
	}
	registry := NewRegistry(vars)

	os.Setenv("PUBLIC_REQUIRED", "value1")
	os.Setenv("PUBLIC_OPTIONAL", "value2")
	os.Setenv("SECRET_REQUIRED", "value3")
	os.Setenv("SECRET_OPTIONAL", "value4")
	defer func() {
		os.Unsetenv("PUBLIC_REQUIRED")
		os.Unsetenv("PUBLIC_OPTIONAL")
		os.Unsetenv("SECRET_REQUIRED")
		os.Unsetenv("SECRET_OPTIONAL")
	}()

	// Only required secrets
	output := registry.Export(ExportOptions{
		Format:       FormatSimple,
		SecretsOnly:  true,
		RequiredOnly: true,
		IncludeEmpty: false,
	})

	if !strings.Contains(output, "SECRET_REQUIRED=value3") {
		t.Error("Should contain required secret")
	}
	if strings.Contains(output, "PUBLIC_REQUIRED") ||
		strings.Contains(output, "PUBLIC_OPTIONAL") ||
		strings.Contains(output, "SECRET_OPTIONAL") {
		t.Error("Should only contain required secrets")
	}
}

// Test special characters in values
func TestRegistry_Export_SpecialChars(t *testing.T) {
	vars := []EnvVar{
		{Name: "SPECIAL_VAR"},
	}
	registry := NewRegistry(vars)

	specialValue := "value with spaces and = and \"quotes\""
	os.Setenv("SPECIAL_VAR", specialValue)
	defer os.Unsetenv("SPECIAL_VAR")

	// Simple format should preserve special chars
	output := registry.Export(ExportOptions{
		Format:       FormatSimple,
		IncludeEmpty: false,
	})

	if !strings.Contains(output, specialValue) {
		t.Error("Special characters not preserved in simple format")
	}

	// K8s format quotes the value
	outputK8s := registry.Export(ExportOptions{
		Format:       FormatK8s,
		IncludeEmpty: false,
	})

	if !strings.Contains(outputK8s, "  value: \""+specialValue+"\"") {
		t.Error("K8s format should quote values")
	}
}

// Test empty registry
func TestRegistry_Export_EmptyRegistry(t *testing.T) {
	registry := NewRegistry([]EnvVar{})

	output := registry.Export(ExportOptions{
		Format:       FormatSimple,
		IncludeEmpty: true,
	})

	if output != "" {
		t.Errorf("Expected empty output for empty registry, got: %s", output)
	}
}

// Test default format fallback
func TestRegistry_Export_InvalidFormat(t *testing.T) {
	vars := []EnvVar{
		{Name: "TEST_VAR"},
	}
	registry := NewRegistry(vars)

	os.Setenv("TEST_VAR", "value")
	defer os.Unsetenv("TEST_VAR")

	// Invalid format should fall back to simple
	output := registry.Export(ExportOptions{
		Format:       ExportFormat("invalid"),
		IncludeEmpty: false,
	})

	// Should default to simple format
	if !strings.Contains(output, "TEST_VAR=value") {
		t.Error("Invalid format should default to simple format")
	}
}
