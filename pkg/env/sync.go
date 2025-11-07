// Package env provides utilities for synchronizing content within files.
package env

import (
	"fmt"
	"os"
	"strings"
)

// SyncOptions configures file section synchronization.
// This is a generic pattern for replacing content between markers in a file,
// useful for keeping auto-generated sections in config files up to date.
type SyncOptions struct {
	FilePath       string // Path to file to modify
	StartMarker    string // Start marker string (exact match)
	EndMarker      string // End marker string (exact match)
	Content        string // New content to insert between markers
	IncludeMarkers bool   // Whether markers are part of replaced content
	DryRun         bool   // Preview changes without writing
	CreateBackup   bool   // Create .backup file before changes
}

// SyncFileSection replaces content between markers in a file.
// This is a generic utility for keeping auto-generated sections synchronized.
//
// The function:
//  1. Reads the file
//  2. Finds start and end markers
//  3. Replaces content between markers
//  4. Optionally creates a backup
//  5. Writes the updated file
//  6. Removes backup on success
//
// Example:
//
//	opts := SyncOptions{
//	  FilePath: "Dockerfile",
//	  StartMarker: "# === START AUTO-GENERATED ===",
//	  EndMarker: "# === END AUTO-GENERATED ===",
//	  Content: newContent,
//	  CreateBackup: true,
//	}
//	err := SyncFileSection(opts)
func SyncFileSection(opts SyncOptions) error {
	// Read file
	data, err := os.ReadFile(opts.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", opts.FilePath, err)
	}

	content := string(data)

	// Find start marker
	startIdx := strings.Index(content, opts.StartMarker)
	if startIdx == -1 {
		return fmt.Errorf("could not find start marker in %s: %q", opts.FilePath, opts.StartMarker)
	}

	// Find end marker (search from after start marker)
	endIdx := strings.Index(content[startIdx:], opts.EndMarker)
	if endIdx == -1 {
		return fmt.Errorf("could not find end marker in %s: %q", opts.FilePath, opts.EndMarker)
	}
	// Convert relative index to absolute
	endIdx = startIdx + endIdx

	// Calculate replacement end position
	replaceEnd := endIdx + len(opts.EndMarker)

	// Build new content
	// Replace from start marker through end marker
	newContent := content[:startIdx] + opts.Content + content[replaceEnd:]

	// Dry run - just print what would change
	if opts.DryRun {
		fmt.Printf("=== Dry Run: %s ===\n", opts.FilePath)
		fmt.Println(opts.Content)
		fmt.Println("=== End Dry Run ===")
		return nil
	}

	// Create backup if requested
	if opts.CreateBackup {
		backupPath := opts.FilePath + ".backup"
		if err := os.WriteFile(backupPath, data, 0644); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		// Remove backup on success (deferred)
		defer func() {
			if err == nil {
				os.Remove(backupPath)
			}
		}()
	}

	// Write updated file
	if err := os.WriteFile(opts.FilePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", opts.FilePath, err)
	}

	return nil
}
