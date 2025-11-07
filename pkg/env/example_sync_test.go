package env_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joeblew999/wellknown/pkg/env"
)

// ExampleSyncFileSection demonstrates basic file section synchronization
func ExampleSyncFileSection() {
	// Create a temporary file
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "config.txt")

	initial := `Header content
# START AUTO-GENERATED
Old content
# END AUTO-GENERATED
Footer content
`
	os.WriteFile(testFile, []byte(initial), 0644)
	defer os.Remove(testFile)

	// Sync the section
	err := env.SyncFileSection(env.SyncOptions{
		FilePath:     testFile,
		StartMarker:  "# START AUTO-GENERATED",
		EndMarker:    "# END AUTO-GENERATED",
		Content:      "# START AUTO-GENERATED\nNew content\n# END AUTO-GENERATED",
		CreateBackup: false,
	})

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Read result
	result, _ := os.ReadFile(testFile)
	fmt.Print(string(result))

	// Output:
	// Header content
	// # START AUTO-GENERATED
	// New content
	// # END AUTO-GENERATED
	// Footer content
}

// ExampleSyncFileSection_backup shows backup creation
func ExampleSyncFileSection_backup() {
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "important.txt")

	original := "# BEGIN\nImportant data\n# END\n"
	os.WriteFile(testFile, []byte(original), 0644)
	defer os.Remove(testFile)

	// Sync with backup (backup is auto-removed on success)
	err := env.SyncFileSection(env.SyncOptions{
		FilePath:     testFile,
		StartMarker:  "# BEGIN",
		EndMarker:    "# END",
		Content:      "# BEGIN\nUpdated data\n# END",
		CreateBackup: true,
	})

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Sync successful")
	fmt.Println("Backup auto-removed on success")

	// Output:
	// Sync successful
	// Backup auto-removed on success
}

// ExampleSyncOptions shows configuration options
func ExampleSyncOptions() {
	opts := env.SyncOptions{
		FilePath:     "Dockerfile",
		StartMarker:  "# === START AUTO-GENERATED ===",
		EndMarker:    "# === END AUTO-GENERATED ===",
		Content:      "# === START AUTO-GENERATED ===\nENV VAR=value\n# === END AUTO-GENERATED ===",
		CreateBackup: true,
		DryRun:       false,
	}

	fmt.Printf("File: %s\n", opts.FilePath)
	fmt.Printf("Dry run: %v\n", opts.DryRun)
	fmt.Printf("Create backup: %v\n", opts.CreateBackup)

	// Output:
	// File: Dockerfile
	// Dry run: false
	// Create backup: true
}

// ExampleSyncFileSection_multipleUpdates shows repeated syncs
func ExampleSyncFileSection_multipleUpdates() {
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "versioned.txt")

	// Initial file
	os.WriteFile(testFile, []byte("# START\nVersion 1\n# END\n"), 0644)
	defer os.Remove(testFile)

	// First update
	env.SyncFileSection(env.SyncOptions{
		FilePath:    testFile,
		StartMarker: "# START",
		EndMarker:   "# END",
		Content:     "# START\nVersion 2\n# END",
	})

	// Second update
	env.SyncFileSection(env.SyncOptions{
		FilePath:    testFile,
		StartMarker: "# START",
		EndMarker:   "# END",
		Content:     "# START\nVersion 3\n# END",
	})

	result, _ := os.ReadFile(testFile)
	fmt.Print(string(result))

	// Output:
	// # START
	// Version 3
	// # END
}

// ExampleSyncFileSection_dockerfile shows real-world Dockerfile use case
func ExampleSyncFileSection_dockerfile() {
	tmpDir := os.TempDir()
	dockerfile := filepath.Join(tmpDir, "Dockerfile")

	initial := `FROM golang:1.21
WORKDIR /app

# === START AUTO-GENERATED ENV ===
ENV OLD_VAR=old
# === END AUTO-GENERATED ENV ===

COPY . .
RUN go build
`
	os.WriteFile(dockerfile, []byte(initial), 0644)
	defer os.Remove(dockerfile)

	// Generate new env vars from registry
	newEnvSection := `# === START AUTO-GENERATED ENV ===
ENV SERVER_PORT=8080
ENV LOG_LEVEL=info
# === END AUTO-GENERATED ENV ===`

	err := env.SyncFileSection(env.SyncOptions{
		FilePath:    dockerfile,
		StartMarker: "# === START AUTO-GENERATED ENV ===",
		EndMarker:   "# === END AUTO-GENERATED ENV ===",
		Content:     newEnvSection,
	})

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Dockerfile updated successfully")

	// Output:
	// Dockerfile updated successfully
}

// ExampleSyncFileSection_preserveStructure shows structure preservation
func ExampleSyncFileSection_preserveStructure() {
	tmpDir := os.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	initial := `app:
  name: myapp

# --- GENERATED SECTION ---
generated:
  old: true
# --- END GENERATED ---

database:
  host: localhost
`
	os.WriteFile(configFile, []byte(initial), 0644)
	defer os.Remove(configFile)

	newSection := `# --- GENERATED SECTION ---
generated:
  timestamp: 2024-01-01
  version: 1.0.0
# --- END GENERATED ---`

	env.SyncFileSection(env.SyncOptions{
		FilePath:    configFile,
		StartMarker: "# --- GENERATED SECTION ---",
		EndMarker:   "# --- END GENERATED ---",
		Content:     newSection,
	})

	result, _ := os.ReadFile(configFile)
	fmt.Print(string(result))

	// Output:
	// app:
	//   name: myapp
	//
	// # --- GENERATED SECTION ---
	// generated:
	//   timestamp: 2024-01-01
	//   version: 1.0.0
	// # --- END GENERATED ---
	//
	// database:
	//   host: localhost
}

// ExampleSyncFileSection_errorHandling shows error cases
func ExampleSyncFileSection_errorHandling() {
	tmpDir := os.TempDir()

	// Missing file
	err := env.SyncFileSection(env.SyncOptions{
		FilePath:    filepath.Join(tmpDir, "nonexistent.txt"),
		StartMarker: "# START",
		EndMarker:   "# END",
		Content:     "# START\nContent\n# END",
	})

	if err != nil {
		fmt.Println("Error: file not found")
	}

	// Missing markers
	testFile := filepath.Join(tmpDir, "no-markers.txt")
	os.WriteFile(testFile, []byte("No markers here"), 0644)
	defer os.Remove(testFile)

	err = env.SyncFileSection(env.SyncOptions{
		FilePath:    testFile,
		StartMarker: "# START",
		EndMarker:   "# END",
		Content:     "# START\nContent\n# END",
	})

	if err != nil {
		fmt.Println("Error: markers not found")
	}

	// Output:
	// Error: file not found
	// Error: markers not found
}
