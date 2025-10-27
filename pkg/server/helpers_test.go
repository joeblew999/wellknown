package server

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadSchemaFromFile tests schema loading from external JSON files
func TestLoadSchemaFromFile(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Find project root (where go.mod is)
	projectRoot := findProjectRoot(t)

	tests := []struct {
		name      string
		platform  string
		appType   string
		schemaType string
		changeDir string // Directory to change to before loading
		wantError bool
		checkContent bool // Whether to verify content is valid JSON
	}{
		{
			name:       "Load Google Calendar schema from project root",
			platform:   "google",
			appType:    "calendar",
			schemaType: "schema",
			changeDir:  projectRoot,
			wantError:  false,
			checkContent: true,
		},
		{
			name:       "Load Google Calendar UI schema from project root",
			platform:   "google",
			appType:    "calendar",
			schemaType: "uischema",
			changeDir:  projectRoot,
			wantError:  false,
			checkContent: true,
		},
		{
			name:       "Load Apple Calendar schema from project root",
			platform:   "apple",
			appType:    "calendar",
			schemaType: "schema",
			changeDir:  projectRoot,
			wantError:  false,
			checkContent: true,
		},
		{
			name:       "Load Apple Calendar UI schema from project root",
			platform:   "apple",
			appType:    "calendar",
			schemaType: "uischema",
			changeDir:  projectRoot,
			wantError:  false,
			checkContent: true,
		},
		{
			name:       "Load from cmd/server directory (Air scenario)",
			platform:   "google",
			appType:    "calendar",
			schemaType: "schema",
			changeDir:  filepath.Join(projectRoot, "cmd", "server"),
			wantError:  false,
			checkContent: true,
		},
		{
			name:       "Non-existent schema",
			platform:   "nonexistent",
			appType:    "calendar",
			schemaType: "schema",
			changeDir:  projectRoot,
			wantError:  true,
			checkContent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Change to test directory
			if tt.changeDir != "" {
				if err := os.Chdir(tt.changeDir); err != nil {
					t.Fatalf("Failed to change directory to %s: %v", tt.changeDir, err)
				}
			}

			// Load schema
			content, err := loadSchemaFromFile(tt.platform, tt.appType, tt.schemaType)

			// Check error expectation
			if (err != nil) != tt.wantError {
				t.Errorf("loadSchemaFromFile() error = %v, wantError %v", err, tt.wantError)
				return
			}

			// Check content if expected
			if tt.checkContent && !tt.wantError {
				if len(content) == 0 {
					t.Error("loadSchemaFromFile() returned empty content")
				}
				// Verify it's valid JSON-like content (starts with { or [)
				trimmed := content[0:1]
				if trimmed != "{" && trimmed != "[" {
					t.Errorf("loadSchemaFromFile() content doesn't look like JSON: starts with %q", trimmed)
				}
			}
		})
	}
}

// findProjectRoot finds the project root directory (where go.mod is)
func findProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Walk up until we find go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root (go.mod)")
		}
		dir = parent
	}
}

// TestLoadSchemaFromFile_Caching tests that schemas can be loaded multiple times
// This is important for performance - we want to ensure loading is fast enough
// even without caching (for now)
func TestLoadSchemaFromFile_MultipleCalls(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	projectRoot := findProjectRoot(t)
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	// Load the same schema multiple times
	const iterations = 10
	for i := 0; i < iterations; i++ {
		content, err := loadSchemaFromFile("google", "calendar", "schema")
		if err != nil {
			t.Fatalf("Iteration %d: loadSchemaFromFile() error = %v", i, err)
		}
		if len(content) == 0 {
			t.Fatalf("Iteration %d: empty content", i)
		}
	}
}
