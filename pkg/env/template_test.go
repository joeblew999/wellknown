package env

import (
	"os"
	"strings"
	"testing"
)

// ================================================================
// GenerateTemplate Tests
// ================================================================

func TestRegistry_GenerateTemplate(t *testing.T) {
	tests := []struct {
		name     string
		vars     []EnvVar
		opts     TemplateOptions
		contains []string
		notContains []string
	}{
		{
			name: "basic template with defaults",
			vars: []EnvVar{
				{Name: "VAR1", Description: "First variable", Default: "default1", Group: "Group A"},
				{Name: "VAR2", Description: "Second variable", Default: "default2", Group: "Group A"},
			},
			opts: TemplateOptions{
				IncludeComments:     true,
				IncludeGroupHeaders: true,
			},
			contains: []string{
				"# First variable",
				"# Second variable",
				"VAR1=default1",
				"VAR2=default2",
				"# Group A",
			},
		},
		{
			name: "custom header and footer",
			vars: []EnvVar{
				{Name: "VAR1", Default: "val1", Group: "Test"},
			},
			opts: TemplateOptions{
				Header: []string{"# CUSTOM HEADER", "# Line 2"},
				Footer: []string{"# CUSTOM FOOTER"},
			},
			contains: []string{
				"# CUSTOM HEADER",
				"# Line 2",
				"# CUSTOM FOOTER",
				"VAR1=val1",
			},
		},
		{
			name: "custom group ordering",
			vars: []EnvVar{
				{Name: "VAR_A", Group: "Zebra"},
				{Name: "VAR_B", Group: "Alpha"},
				{Name: "VAR_C", Group: "Midway"},
			},
			opts: TemplateOptions{
				GroupOrder:          []string{"Alpha", "Midway", "Zebra"},
				IncludeGroupHeaders: true,
			},
			contains: []string{
				"Alpha",
				"Midway",
				"Zebra",
			},
		},
		{
			name: "value overrides",
			vars: []EnvVar{
				{Name: "URL", Default: "https://prod.example.com", Group: "Config"},
				{Name: "PORT", Default: "8080", Group: "Config"},
			},
			opts: TemplateOptions{
				ValueOverrides: func(v EnvVar) (string, bool) {
					if v.Name == "URL" {
						return "https://localhost:3000", true
					}
					return "", false
				},
			},
			contains: []string{
				"URL=https://localhost:3000",
				"PORT=8080",
			},
			notContains: []string{
				"https://prod.example.com",
			},
		},
		{
			name: "required variables marked",
			vars: []EnvVar{
				{Name: "REQUIRED_VAR", Required: true, Description: "Must be set", Group: "Test"},
				{Name: "OPTIONAL_VAR", Required: false, Group: "Test"},
			},
			opts: TemplateOptions{
				IncludeComments: true,
			},
			contains: []string{
				"# Must be set",
				"# REQUIRED",
				"REQUIRED_VAR=",
				"OPTIONAL_VAR=",
			},
		},
		{
			name: "empty values",
			vars: []EnvVar{
				{Name: "EMPTY_VAR", Default: "", Group: "Test"},
			},
			opts: TemplateOptions{},
			contains: []string{
				"EMPTY_VAR=",
			},
		},
		{
			name: "without comments",
			vars: []EnvVar{
				{Name: "VAR1", Description: "This should not appear", Group: "Test"},
			},
			opts: TemplateOptions{
				IncludeComments: false,
			},
			contains: []string{
				"VAR1=",
			},
			notContains: []string{
				"This should not appear",
			},
		},
		{
			name: "without group headers",
			vars: []EnvVar{
				{Name: "VAR1", Group: "TestGroup"},
			},
			opts: TemplateOptions{
				IncludeGroupHeaders: false,
			},
			contains: []string{
				"VAR1=",
			},
			notContains: []string{
				"TestGroup",
			},
		},
		{
			name: "custom group header format",
			vars: []EnvVar{
				{Name: "VAR1", Group: "Custom"},
			},
			opts: TemplateOptions{
				IncludeGroupHeaders: true,
				GroupHeaderFormat: func(groupName string) string {
					return ">>> " + groupName + " <<<\n"
				},
			},
			contains: []string{
				">>> Custom <<<",
			},
			notContains: []string{
				"# -----",
			},
		},
		{
			name: "alphabetical ordering without GroupOrder",
			vars: []EnvVar{
				{Name: "VAR1", Group: "Zebra"},
				{Name: "VAR2", Group: "Alpha"},
				{Name: "VAR3", Group: "Midway"},
			},
			opts: TemplateOptions{
				IncludeGroupHeaders: true,
			},
			contains: []string{
				"Alpha",
				"Midway",
				"Zebra",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry(tt.vars)
			result := registry.GenerateTemplate(tt.opts)

			for _, needle := range tt.contains {
				if !strings.Contains(result, needle) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", needle, result)
				}
			}

			for _, needle := range tt.notContains {
				if strings.Contains(result, needle) {
					t.Errorf("Expected output NOT to contain %q, but it did.\nOutput:\n%s", needle, result)
				}
			}
		})
	}
}

func TestRegistry_GenerateTemplate_GroupSorting(t *testing.T) {
	vars := []EnvVar{
		{Name: "Z1", Group: "Zebra"},
		{Name: "A1", Group: "Alpha"},
		{Name: "Z2", Group: "Zebra"},
		{Name: "A2", Group: "Alpha"},
		{Name: "M1", Group: "Midway"},
	}

	registry := NewRegistry(vars)
	result := registry.GenerateTemplate(TemplateOptions{
		IncludeGroupHeaders: true,
	})

	// Check alphabetical group order
	alphaIdx := strings.Index(result, "Alpha")
	midwayIdx := strings.Index(result, "Midway")
	zebraIdx := strings.Index(result, "Zebra")

	if alphaIdx == -1 || midwayIdx == -1 || zebraIdx == -1 {
		t.Fatal("Expected all groups to be present")
	}

	if !(alphaIdx < midwayIdx && midwayIdx < zebraIdx) {
		t.Errorf("Expected alphabetical order: Alpha < Midway < Zebra, got indices: %d, %d, %d", alphaIdx, midwayIdx, zebraIdx)
	}

	// Check variables within groups are sorted
	a1Idx := strings.Index(result, "A1=")
	a2Idx := strings.Index(result, "A2=")
	z1Idx := strings.Index(result, "Z1=")
	z2Idx := strings.Index(result, "Z2=")

	if a1Idx > a2Idx {
		t.Error("Expected A1 before A2 within Alpha group")
	}
	if z1Idx > z2Idx {
		t.Error("Expected Z1 before Z2 within Zebra group")
	}
}

// ================================================================
// GenerateEnvExample Tests
// ================================================================

func TestRegistry_GenerateEnvExample(t *testing.T) {
	vars := []EnvVar{
		{Name: "VAR1", Description: "Test var 1", Required: true, Group: "Test"},
		{Name: "VAR2", Description: "Test var 2", Default: "default", Group: "Test"},
	}

	registry := NewRegistry(vars)
	result := registry.GenerateEnvExample("TestApp")

	expectedContains := []string{
		"# TestApp Environment Variables",
		"# REQUIRED",
		"# Test var 1",
		"# Test var 2",
		"VAR1=",
		"VAR2=default",
	}

	for _, needle := range expectedContains {
		if !strings.Contains(result, needle) {
			t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", needle, result)
		}
	}
}

// ================================================================
// GenerateEnvList Tests
// ================================================================

func TestRegistry_GenerateEnvList(t *testing.T) {
	// Set test environment variable
	os.Setenv("TEST_PUBLIC_VAR", "visible_value")
	os.Setenv("TEST_SECRET_VAR", "secret123")
	defer func() {
		os.Unsetenv("TEST_PUBLIC_VAR")
		os.Unsetenv("TEST_SECRET_VAR")
	}()

	vars := []EnvVar{
		{Name: "TEST_PUBLIC_VAR", Description: "Public variable", Secret: false, Group: "Test"},
		{Name: "TEST_SECRET_VAR", Description: "Secret variable", Secret: true, Group: "Test"},
		{Name: "TEST_UNSET_VAR", Description: "Not set", Default: "default_val", Group: "Test"},
	}

	registry := NewRegistry(vars)
	result := registry.GenerateEnvList("Custom Title")

	expectedContains := []string{
		"Custom Title",
		"## Test",
		"TEST_PUBLIC_VAR",
		"visible_value", // Public var shows actual value
		"TEST_SECRET_VAR [SECRET]",
		"***set***", // Secret var is masked
		"TEST_UNSET_VAR",
		"(default: default_val)",
	}

	for _, needle := range expectedContains {
		if !strings.Contains(result, needle) {
			t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", needle, result)
		}
	}

	// Secret should NOT show actual value
	if strings.Contains(result, "secret123") {
		t.Error("Secret value should be masked, but wasn't")
	}
}

func TestRegistry_GenerateEnvList_EmptyTitle(t *testing.T) {
	vars := []EnvVar{
		{Name: "VAR1", Group: "Test"},
	}

	registry := NewRegistry(vars)
	result := registry.GenerateEnvList("")

	if !strings.Contains(result, "Environment Variables Registry") {
		t.Error("Expected default title when empty string provided")
	}
}

// ================================================================
// GenerateDockerfileDocs Tests
// ================================================================

func TestRegistry_GenerateDockerfileDocs(t *testing.T) {
	vars := []EnvVar{
		{Name: "PUBLIC_REQUIRED", Required: true, Secret: false, Default: "value1", Description: "Public required", Group: "Config"},
		{Name: "SECRET_REQUIRED", Required: true, Secret: true, Description: "Secret required", Group: "Auth"},
		{Name: "SECRET_OPTIONAL", Required: false, Secret: true, Description: "Optional secret", Group: "Auth"},
		{Name: "PUBLIC_OPTIONAL", Required: false, Secret: false, Default: "val2", Group: "Config"},
	}

	registry := NewRegistry(vars)
	result := registry.GenerateDockerfileDocs(DockerfileDocsOptions{
		AppName:            "MyApp",
		UpdateCommand:      "make update-dockerfile",
		DeploymentPlatform: "AWS ECS",
		NonSecretEnvSource: "task definition",
		SecretSource:       "AWS Secrets Manager",
		SyncCommand:        "make sync-secrets",
	})

	expectedContains := []string{
		"AWS ECS",
		"make update-dockerfile",
		"# Required (set via task definition):",
		"#   PUBLIC_REQUIRED=value1",
		"#     Public required",
		"# Required (set via AWS Secrets Manager):",
		"#   SECRET_REQUIRED",
		"#     Secret required",
		"# Optional (set via AWS Secrets Manager if needed):",
		"#   # Auth",
		"#   SECRET_OPTIONAL",
		"make sync-secrets",
	}

	for _, needle := range expectedContains {
		if !strings.Contains(result, needle) {
			t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", needle, result)
		}
	}
}

func TestRegistry_GenerateDockerfileDocs_NoRequiredVars(t *testing.T) {
	vars := []EnvVar{
		{Name: "OPTIONAL_VAR", Required: false, Secret: true, Group: "Test"},
	}

	registry := NewRegistry(vars)
	result := registry.GenerateDockerfileDocs(DockerfileDocsOptions{
		DeploymentPlatform: "Test",
		NonSecretEnvSource: "env file",
		SecretSource:       "vault",
	})

	if !strings.Contains(result, "#   (none)") {
		t.Error("Expected '(none)' marker for empty required sections")
	}
}

// ================================================================
// GenerateTOMLEnv Tests
// ================================================================

func TestRegistry_GenerateTOMLEnv(t *testing.T) {
	vars := []EnvVar{
		{Name: "PUBLIC_VAR", Secret: false, Default: "value1", Group: "Config"},
		{Name: "SECRET_VAR", Secret: true, Default: "secret", Group: "Config"},
		{Name: "NO_DEFAULT", Secret: false, Default: "", Group: "Config"},
	}

	registry := NewRegistry(vars)
	result := registry.GenerateTOMLEnv("env", []string{
		"Configuration values",
		"Secrets are managed separately",
	})

	expectedContains := []string{
		"[env]",
		"# Configuration values",
		"# Secrets are managed separately",
		`PUBLIC_VAR = "value1"`,
	}

	for _, needle := range expectedContains {
		if !strings.Contains(result, needle) {
			t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", needle, result)
		}
	}

	// Secrets should NOT appear in TOML
	notExpected := []string{
		"SECRET_VAR",
		"NO_DEFAULT",
	}

	for _, needle := range notExpected {
		if strings.Contains(result, needle) {
			t.Errorf("Expected output NOT to contain %q, but it did.\nOutput:\n%s", needle, result)
		}
	}
}

func TestRegistry_GenerateTOMLEnv_CustomSection(t *testing.T) {
	vars := []EnvVar{
		{Name: "VAR1", Secret: false, Default: "val1", Group: "Test"},
	}

	registry := NewRegistry(vars)
	result := registry.GenerateTOMLEnv("app.config", []string{"Custom section"})

	if !strings.Contains(result, "[app.config]") {
		t.Error("Expected custom section name in output")
	}
}

// ================================================================
// Edge Cases and Error Handling
// ================================================================

func TestRegistry_GenerateTemplate_EmptyRegistry(t *testing.T) {
	registry := NewRegistry([]EnvVar{})
	result := registry.GenerateTemplate(TemplateOptions{
		Header: []string{"# Header"},
		Footer: []string{"# Footer"},
	})

	if !strings.Contains(result, "# Header") {
		t.Error("Expected header even with empty registry")
	}
	if !strings.Contains(result, "# Footer") {
		t.Error("Expected footer even with empty registry")
	}
}

func TestRegistry_GenerateTemplate_SpecialCharacters(t *testing.T) {
	vars := []EnvVar{
		{Name: "URL", Default: "https://example.com/path?query=value&other=123", Group: "Test"},
		{Name: "JSON", Default: `{"key":"value"}`, Group: "Test"},
		{Name: "MULTILINE", Default: "line1\\nline2", Group: "Test"},
	}

	registry := NewRegistry(vars)
	result := registry.GenerateTemplate(TemplateOptions{})

	expectedContains := []string{
		`URL=https://example.com/path?query=value&other=123`,
		`JSON={"key":"value"}`,
		`MULTILINE=line1\nline2`,
	}

	for _, needle := range expectedContains {
		if !strings.Contains(result, needle) {
			t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", needle, result)
		}
	}
}

func TestRegistry_GenerateTemplate_NonexistentGroupOrder(t *testing.T) {
	vars := []EnvVar{
		{Name: "VAR1", Group: "GroupA"},
		{Name: "VAR2", Group: "GroupB"},
	}

	registry := NewRegistry(vars)
	result := registry.GenerateTemplate(TemplateOptions{
		GroupOrder: []string{"NonExistent", "GroupB", "AnotherMissing"},
	})

	// Should only include GroupB (exists in GroupOrder and in vars)
	if !strings.Contains(result, "VAR2=") {
		t.Error("Expected VAR2 from GroupB")
	}

	// GroupA not in GroupOrder, should not appear
	if strings.Contains(result, "VAR1=") {
		t.Error("Did not expect VAR1 from GroupA (not in GroupOrder)")
	}
}

func TestRegistry_GenerateEnvList_RequiredAndSecretMarkers(t *testing.T) {
	vars := []EnvVar{
		{Name: "BOTH", Required: true, Secret: true, Group: "Test"},
		{Name: "JUST_REQUIRED", Required: true, Secret: false, Group: "Test"},
		{Name: "JUST_SECRET", Required: false, Secret: true, Group: "Test"},
		{Name: "NEITHER", Required: false, Secret: false, Group: "Test"},
	}

	registry := NewRegistry(vars)
	result := registry.GenerateEnvList("")

	expectedContains := []string{
		"BOTH [REQUIRED] [SECRET]",
		"JUST_REQUIRED [REQUIRED]",
		"JUST_SECRET [SECRET]",
		"NEITHER",
	}

	for _, needle := range expectedContains {
		if !strings.Contains(result, needle) {
			t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", needle, result)
		}
	}
}

func TestRegistry_GenerateTemplate_HeaderFooterNewlines(t *testing.T) {
	vars := []EnvVar{
		{Name: "VAR1", Group: "Test"},
	}

	registry := NewRegistry(vars)

	// Test with newlines already present
	result1 := registry.GenerateTemplate(TemplateOptions{
		Header: []string{"# Line 1\n", "# Line 2\n"},
		Footer: []string{"# Footer\n"},
	})

	// Test without newlines
	result2 := registry.GenerateTemplate(TemplateOptions{
		Header: []string{"# Line 1", "# Line 2"},
		Footer: []string{"# Footer"},
	})

	// Both should produce valid output with newlines
	if !strings.Contains(result1, "# Line 1\n") || !strings.Contains(result1, "# Footer\n") {
		t.Error("Expected newlines to be preserved when provided")
	}

	if !strings.Contains(result2, "# Line 1\n") || !strings.Contains(result2, "# Footer\n") {
		t.Error("Expected newlines to be added when not provided")
	}
}

// ================================================================
// GenerateTOMLSecretsList Tests
// ================================================================

func TestRegistry_GenerateTOMLSecretsList(t *testing.T) {
	tests := []struct {
		name          string
		vars          []EnvVar
		importCommand string
		want          []string
		notWant       []string
	}{
		{
			name: "basic secrets list",
			vars: []EnvVar{
				{Name: "DATABASE_URL", Secret: true, Required: true},
				{Name: "API_KEY", Secret: true},
				{Name: "SERVER_PORT", Default: "8080"}, // not secret
			},
			importCommand: "go run . fly-secrets-import",
			want: []string{
				"# === START AUTO-GENERATED SECRETS LIST ===",
				"# Secrets (set via: go run . fly-secrets-import)",
				"# - DATABASE_URL",
				"# - API_KEY",
				"# === END AUTO-GENERATED SECRETS LIST ===",
			},
			notWant: []string{
				"SERVER_PORT",
			},
		},
		{
			name:          "no secrets",
			vars:          []EnvVar{{Name: "VAR1", Default: "val"}},
			importCommand: "fly secrets import",
			want: []string{
				"# === START AUTO-GENERATED SECRETS LIST ===",
				"# Secrets (set via: fly secrets import)",
				"# === END AUTO-GENERATED SECRETS LIST ===",
			},
			notWant: []string{
				"VAR1",
			},
		},
		{
			name: "multiple secrets different groups",
			vars: []EnvVar{
				{Name: "DB_PASSWORD", Secret: true, Group: "Database"},
				{Name: "STRIPE_KEY", Secret: true, Group: "APIs"},
				{Name: "JWT_SECRET", Secret: true, Group: "Auth"},
			},
			importCommand: "make secrets-import",
			want: []string{
				"# - DB_PASSWORD",
				"# - STRIPE_KEY",
				"# - JWT_SECRET",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := NewRegistry(tt.vars)
			result := reg.GenerateTOMLSecretsList(tt.importCommand)

			for _, want := range tt.want {
				if !strings.Contains(result, want) {
					t.Errorf("GenerateTOMLSecretsList() missing expected content:\n  want: %q\n  got: %s", want, result)
				}
			}

			for _, notWant := range tt.notWant {
				if strings.Contains(result, notWant) {
					t.Errorf("GenerateTOMLSecretsList() contains unexpected content:\n  don't want: %q\n  got: %s", notWant, result)
				}
			}
		})
	}
}

// ================================================================
// GenerateDockerComposeEnv Tests
// ================================================================

func TestRegistry_GenerateDockerComposeEnv(t *testing.T) {
	tests := []struct {
		name     string
		vars     []EnvVar
		comments []string
		want     []string
		notWant  []string
	}{
		{
			name: "basic docker-compose env",
			vars: []EnvVar{
				{Name: "SERVER_PORT", Default: "8080"},
				{Name: "LOG_LEVEL", Default: "info"},
				{Name: "DATABASE_URL", Secret: true, Required: true}, // secret - excluded
			},
			comments: []string{
				"AUTO-GENERATED - DO NOT EDIT",
				"Update with: make sync",
			},
			want: []string{
				"    environment:",
				"      # AUTO-GENERATED - DO NOT EDIT",
				"      # Update with: make sync",
				`      SERVER_PORT: "8080"`,
				`      LOG_LEVEL: "info"`,
			},
			notWant: []string{
				"DATABASE_URL",
			},
		},
		{
			name: "no defaults",
			vars: []EnvVar{
				{Name: "VAR1"}, // no default
				{Name: "VAR2", Secret: true},
			},
			comments: []string{"Generated"},
			want: []string{
				"    environment:",
				"      # Generated",
			},
			notWant: []string{
				"VAR1",
				"VAR2",
			},
		},
		{
			name: "mixed vars with defaults",
			vars: []EnvVar{
				{Name: "PUBLIC_VAR", Default: "public"},
				{Name: "NO_DEFAULT"},
				{Name: "SECRET_VAR", Default: "secret", Secret: true},
				{Name: "ANOTHER_PUBLIC", Default: "value"},
			},
			comments: []string{},
			want: []string{
				"    environment:",
				`      PUBLIC_VAR: "public"`,
				`      ANOTHER_PUBLIC: "value"`,
			},
			notWant: []string{
				"NO_DEFAULT",
				"SECRET_VAR",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := NewRegistry(tt.vars)
			result := reg.GenerateDockerComposeEnv(tt.comments)

			for _, want := range tt.want {
				if !strings.Contains(result, want) {
					t.Errorf("GenerateDockerComposeEnv() missing expected content:\n  want: %q\n  got: %s", want, result)
				}
			}

			for _, notWant := range tt.notWant {
				if strings.Contains(result, notWant) {
					t.Errorf("GenerateDockerComposeEnv() contains unexpected content:\n  don't want: %q\n  got: %s", notWant, result)
				}
			}
		})
	}
}
