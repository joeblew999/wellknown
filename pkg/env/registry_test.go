package env

import (
	"os"
	"reflect"
	"testing"
)

// Test NewRegistry creates registry with index
func TestNewRegistry(t *testing.T) {
	vars := []EnvVar{
		{Name: "TEST_VAR", Description: "Test", Required: true},
		{Name: "OPTIONAL", Description: "Optional", Required: false},
	}

	registry := NewRegistry(vars)

	// Test index was built
	if registry.ByName("TEST_VAR") == nil {
		t.Error("Expected TEST_VAR in index")
	}
	if registry.ByName("OPTIONAL") == nil {
		t.Error("Expected OPTIONAL in index")
	}
	if registry.ByName("NONEXISTENT") != nil {
		t.Error("Expected nil for nonexistent variable")
	}

	// Test that we have the right number of variables
	if len(registry.All()) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(registry.All()))
	}
}

// Test ByName returns correct variable
func TestRegistry_ByName(t *testing.T) {
	vars := []EnvVar{
		{Name: "VAR1", Description: "First"},
		{Name: "VAR2", Description: "Second"},
	}
	registry := NewRegistry(vars)

	tests := []struct {
		name     string
		varName  string
		wantNil  bool
		wantDesc string
	}{
		{"found VAR1", "VAR1", false, "First"},
		{"found VAR2", "VAR2", false, "Second"},
		{"not found", "VAR3", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.ByName(tt.varName)
			if tt.wantNil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Fatal("Expected variable, got nil")
				}
				if result.Description != tt.wantDesc {
					t.Errorf("Description = %q, want %q", result.Description, tt.wantDesc)
				}
			}
		})
	}
}

// Table-driven test for GetString
func TestEnvVar_GetString(t *testing.T) {
	tests := []struct {
		name     string
		envName  string
		envValue string
		def      string
		want     string
	}{
		{"returns env value", "TEST_VAR", "hello", "default", "hello"},
		{"returns default when not set", "NOT_SET", "", "default", "default"},
		{"returns empty when no default", "NOT_SET", "", "", ""},
		{"env overrides default", "TEST_VAR2", "env_value", "default_value", "env_value"},
		{"empty env uses default", "EMPTY_VAR", "", "fallback", "fallback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment
			os.Unsetenv(tt.envName)

			if tt.envValue != "" {
				os.Setenv(tt.envName, tt.envValue)
				defer os.Unsetenv(tt.envName)
			}

			v := EnvVar{Name: tt.envName, Default: tt.def}
			if got := v.GetString(); got != tt.want {
				t.Errorf("GetString() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Table-driven test for GetInt
func TestEnvVar_GetInt(t *testing.T) {
	tests := []struct {
		name     string
		envName  string
		envValue string
		def      string
		want     int
	}{
		{"parses env int", "TEST_INT", "42", "0", 42},
		{"returns default when not set", "NOT_SET", "", "10", 10},
		{"returns 0 when no default", "NOT_SET", "", "", 0},
		{"returns default on parse error", "BAD_INT", "not_a_number", "5", 5},
		{"handles negative numbers", "NEG_INT", "-100", "0", -100},
		{"handles zero", "ZERO_INT", "0", "999", 0},
		{"returns 0 on bad default", "NOT_SET", "", "bad_default", 0},
		{"env overrides default", "TEST_INT2", "123", "456", 123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment
			os.Unsetenv(tt.envName)

			if tt.envValue != "" {
				os.Setenv(tt.envName, tt.envValue)
				defer os.Unsetenv(tt.envName)
			}

			v := EnvVar{Name: tt.envName, Default: tt.def}
			if got := v.GetInt(); got != tt.want {
				t.Errorf("GetInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Table-driven test for GetBool
func TestEnvVar_GetBool(t *testing.T) {
	tests := []struct {
		name     string
		envName  string
		envValue string
		def      string
		want     bool
	}{
		// True values
		{"true literal", "TEST_BOOL", "true", "false", true},
		{"1 as true", "TEST_BOOL", "1", "false", true},
		{"yes as true", "TEST_BOOL", "yes", "false", true},
		{"TRUE uppercase", "TEST_BOOL", "TRUE", "false", true},
		{"Yes mixed case", "TEST_BOOL", "Yes", "false", true},

		// False values
		{"false literal", "TEST_BOOL", "false", "true", false},
		{"0 as false", "TEST_BOOL", "0", "true", false},
		{"no as false", "TEST_BOOL", "no", "true", false},
		{"FALSE uppercase", "TEST_BOOL", "FALSE", "true", false},

		// Invalid values fallback to default
		{"invalid falls to default true", "TEST_BOOL", "invalid", "true", true},
		{"invalid falls to default false", "TEST_BOOL", "invalid", "false", false},
		{"empty falls to default", "TEST_BOOL", "invalid", "1", true},

		// Not set uses default
		{"not set uses default true", "NOT_SET", "", "true", true},
		{"not set uses default false", "NOT_SET", "", "false", false},
		{"not set uses default yes", "NOT_SET", "", "yes", true},
		{"not set no default is false", "NOT_SET", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment
			os.Unsetenv(tt.envName)

			if tt.envValue != "" {
				os.Setenv(tt.envName, tt.envValue)
				defer os.Unsetenv(tt.envName)
			}

			v := EnvVar{Name: tt.envName, Default: tt.def}
			if got := v.GetBool(); got != tt.want {
				t.Errorf("GetBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test GetRequired returns only required variables
func TestRegistry_GetRequired(t *testing.T) {
	vars := []EnvVar{
		{Name: "REQUIRED1", Required: true},
		{Name: "OPTIONAL1", Required: false},
		{Name: "REQUIRED2", Required: true},
		{Name: "OPTIONAL2", Required: false},
	}
	registry := NewRegistry(vars)

	required := registry.GetRequired()

	if len(required) != 2 {
		t.Errorf("Expected 2 required vars, got %d", len(required))
	}

	// Check that we got the right ones
	names := make(map[string]bool)
	for _, v := range required {
		names[v.Name] = true
	}

	if !names["REQUIRED1"] || !names["REQUIRED2"] {
		t.Error("Expected REQUIRED1 and REQUIRED2")
	}
	if names["OPTIONAL1"] || names["OPTIONAL2"] {
		t.Error("Did not expect optional variables")
	}
}

// Test GetSecrets returns only secret variables
func TestRegistry_GetSecrets(t *testing.T) {
	vars := []EnvVar{
		{Name: "PUBLIC1", Secret: false},
		{Name: "SECRET1", Secret: true},
		{Name: "PUBLIC2", Secret: false},
		{Name: "SECRET2", Secret: true},
	}
	registry := NewRegistry(vars)

	secrets := registry.GetSecrets()

	if len(secrets) != 2 {
		t.Errorf("Expected 2 secret vars, got %d", len(secrets))
	}

	// Check that we got the right ones
	names := make(map[string]bool)
	for _, v := range secrets {
		names[v.Name] = true
	}

	if !names["SECRET1"] || !names["SECRET2"] {
		t.Error("Expected SECRET1 and SECRET2")
	}
	if names["PUBLIC1"] || names["PUBLIC2"] {
		t.Error("Did not expect public variables")
	}
}

// Test GetByGroup groups variables correctly
func TestRegistry_GetByGroup(t *testing.T) {
	vars := []EnvVar{
		{Name: "SERVER_PORT", Group: "Server"},
		{Name: "SERVER_HOST", Group: "Server"},
		{Name: "DB_URL", Group: "Database"},
		{Name: "DB_PASSWORD", Group: "Database"},
		{Name: "UNGROUPED", Group: ""},
	}
	registry := NewRegistry(vars)

	groups := registry.GetByGroup()

	if len(groups["Server"]) != 2 {
		t.Errorf("Expected 2 Server vars, got %d", len(groups["Server"]))
	}
	if len(groups["Database"]) != 2 {
		t.Errorf("Expected 2 Database vars, got %d", len(groups["Database"]))
	}
	if len(groups[""]) != 1 {
		t.Errorf("Expected 1 ungrouped var, got %d", len(groups[""]))
	}

	// Check specific variables
	serverNames := make(map[string]bool)
	for _, v := range groups["Server"] {
		serverNames[v.Name] = true
	}
	if !serverNames["SERVER_PORT"] || !serverNames["SERVER_HOST"] {
		t.Error("Server group missing expected variables")
	}
}

// Test AllSorted returns variables sorted by group then name
func TestRegistry_AllSorted(t *testing.T) {
	vars := []EnvVar{
		{Name: "Z_VAR", Group: "Z"},
		{Name: "A_VAR", Group: "A"},
		{Name: "B_VAR", Group: "A"},
		{Name: "A_VAR2", Group: "B"},
	}
	registry := NewRegistry(vars)

	sorted := registry.AllSorted()

	// Should be: A/A_VAR, A/B_VAR, B/A_VAR2, Z/Z_VAR
	expected := []string{"A_VAR", "B_VAR", "A_VAR2", "Z_VAR"}
	for i, v := range sorted {
		if v.Name != expected[i] {
			t.Errorf("Position %d: got %s, want %s", i, v.Name, expected[i])
		}
	}
}

// Test ValidateRequired with all required vars set
func TestRegistry_ValidateRequired_AllSet(t *testing.T) {
	vars := []EnvVar{
		{Name: "REQUIRED1", Required: true},
		{Name: "REQUIRED2", Required: true},
		{Name: "OPTIONAL", Required: false},
	}
	registry := NewRegistry(vars)

	// Set required vars
	os.Setenv("REQUIRED1", "value1")
	os.Setenv("REQUIRED2", "value2")
	defer func() {
		os.Unsetenv("REQUIRED1")
		os.Unsetenv("REQUIRED2")
	}()

	err := registry.ValidateRequired()
	if err != nil {
		t.Errorf("ValidateRequired() failed when all required vars set: %v", err)
	}
}

// Test ValidateRequired with missing required vars
func TestRegistry_ValidateRequired_Missing(t *testing.T) {
	vars := []EnvVar{
		{Name: "REQUIRED1", Required: true},
		{Name: "REQUIRED2", Required: true},
		{Name: "OPTIONAL", Required: false},
	}
	registry := NewRegistry(vars)

	// Clean environment
	os.Unsetenv("REQUIRED1")
	os.Unsetenv("REQUIRED2")
	os.Unsetenv("OPTIONAL")

	err := registry.ValidateRequired()
	if err == nil {
		t.Error("ValidateRequired() should fail when required vars missing")
	}

	// Error should mention both missing vars
	errMsg := err.Error()
	if !contains(errMsg, "REQUIRED1") || !contains(errMsg, "REQUIRED2") {
		t.Errorf("Error should mention missing vars, got: %s", errMsg)
	}
}

// Test ValidateRequired with partially missing vars
func TestRegistry_ValidateRequired_PartiallyMissing(t *testing.T) {
	vars := []EnvVar{
		{Name: "SET_VAR", Required: true},
		{Name: "MISSING_VAR", Required: true},
	}
	registry := NewRegistry(vars)

	// Set only one
	os.Setenv("SET_VAR", "value")
	os.Unsetenv("MISSING_VAR")
	defer os.Unsetenv("SET_VAR")

	err := registry.ValidateRequired()
	if err == nil {
		t.Error("ValidateRequired() should fail when some required vars missing")
	}

	errMsg := err.Error()
	if !contains(errMsg, "MISSING_VAR") {
		t.Errorf("Error should mention MISSING_VAR, got: %s", errMsg)
	}
	if contains(errMsg, "SET_VAR") {
		t.Errorf("Error should not mention SET_VAR, got: %s", errMsg)
	}
}

// Test All returns all variables
func TestRegistry_All(t *testing.T) {
	vars := []EnvVar{
		{Name: "VAR1"},
		{Name: "VAR2"},
		{Name: "VAR3"},
	}
	registry := NewRegistry(vars)

	all := registry.All()

	if len(all) != 3 {
		t.Errorf("Expected 3 variables, got %d", len(all))
	}

	// Should return same slice
	if !reflect.DeepEqual(all, vars) {
		t.Error("All() did not return expected variables")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}
