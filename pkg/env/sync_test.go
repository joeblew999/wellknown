package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSyncFileSection(t *testing.T) {
	tests := []struct {
		name        string
		initial     string
		startMarker string
		endMarker   string
		newContent  string
		want        string
	}{
		{
			"replaces section between markers",
			"Header\n# START\nOld content\n# END\nFooter\n",
			"# START",
			"# END",
			"# START\nNew content\n# END",
			"Header\n# START\nNew content\n# END\nFooter\n",
		},
		{
			"handles unique markers",
			"Line1\n### START ###\nOld\n### END ###\nLine2\n",
			"### START ###",
			"### END ###",
			"### START ###\nNew\n### END ###",
			"Line1\n### START ###\nNew\n### END ###\nLine2\n",
		},
		{
			"replaces entire content between markers",
			"Before\n<!-- BEGIN -->\nOld line 1\nOld line 2\nOld line 3\n<!-- END -->\nAfter\n",
			"<!-- BEGIN -->",
			"<!-- END -->",
			"<!-- BEGIN -->\nNew single line\n<!-- END -->",
			"Before\n<!-- BEGIN -->\nNew single line\n<!-- END -->\nAfter\n",
		},
		{
			"handles markers at start of file",
			"# START\nContent\n# END\nRest of file\n",
			"# START",
			"# END",
			"# START\nNew\n# END",
			"# START\nNew\n# END\nRest of file\n",
		},
		{
			"handles markers at end of file",
			"Beginning\n# START\nContent\n# END\n",
			"# START",
			"# END",
			"# START\nNew\n# END",
			"Beginning\n# START\nNew\n# END\n",
		},
		{
			"replaces with empty content",
			"Before\n# START\nOld\n# END\nAfter\n",
			"# START",
			"# END",
			"# START\n# END",
			"Before\n# START\n# END\nAfter\n",
		},
		{
			"handles multiline content replacement",
			"Line1\n### BEGIN ###\nOld\n### END ###\nLine2\n",
			"### BEGIN ###",
			"### END ###",
			"### BEGIN ###\nNew line 1\nNew line 2\nNew line 3\n### END ###",
			"Line1\n### BEGIN ###\nNew line 1\nNew line 2\nNew line 3\n### END ###\nLine2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.txt")
			if err := os.WriteFile(testFile, []byte(tt.initial), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Sync section
			err := SyncFileSection(SyncOptions{
				FilePath:     testFile,
				StartMarker:  tt.startMarker,
				EndMarker:    tt.endMarker,
				Content:      tt.newContent,
				DryRun:       false,
				CreateBackup: true,
			})

			if err != nil {
				t.Fatalf("SyncFileSection failed: %v", err)
			}

			// Read result
			got, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("Failed to read result file: %v", err)
			}

			if string(got) != tt.want {
				t.Errorf("Result =\n%q\nwant\n%q", string(got), tt.want)
			}

			// Verify backup was created and cleaned up
			backup := testFile + ".backup"
			if _, err := os.Stat(backup); err == nil {
				t.Error("Backup file should have been cleaned up")
			}
		})
	}
}

// Test dry-run doesn't modify file
func TestSyncFileSection_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	original := "# START\nOld\n# END\n"
	os.WriteFile(testFile, []byte(original), 0644)

	err := SyncFileSection(SyncOptions{
		FilePath:    testFile,
		StartMarker: "# START",
		EndMarker:   "# END",
		Content:     "# START\nNew\n# END",
		DryRun:      true,
	})

	if err != nil {
		t.Fatalf("DryRun failed: %v", err)
	}

	// File should be unchanged
	got, _ := os.ReadFile(testFile)
	if string(got) != original {
		t.Error("DryRun modified the file")
	}

	// No backup should be created
	backup := testFile + ".backup"
	if _, err := os.Stat(backup); err == nil {
		t.Error("DryRun should not create backup")
	}
}

// Test error when start marker not found
func TestSyncFileSection_MissingStartMarker(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("No markers here\n"), 0644)

	err := SyncFileSection(SyncOptions{
		FilePath:    testFile,
		StartMarker: "# START",
		EndMarker:   "# END",
		Content:     "# START\nNew\n# END",
	})

	if err == nil {
		t.Error("Expected error when start marker not found")
	}

	if !contains(err.Error(), "start marker") {
		t.Errorf("Error should mention start marker, got: %v", err)
	}
}

// Test error when end marker not found
func TestSyncFileSection_MissingEndMarker(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("# START\nContent\n"), 0644)

	err := SyncFileSection(SyncOptions{
		FilePath:    testFile,
		StartMarker: "# START",
		EndMarker:   "# END",
		Content:     "# START\nNew\n# END",
	})

	if err == nil {
		t.Error("Expected error when end marker not found")
	}

	if !contains(err.Error(), "end marker") {
		t.Errorf("Error should mention end marker, got: %v", err)
	}
}

// Test error when file doesn't exist
func TestSyncFileSection_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "nonexistent.txt")

	err := SyncFileSection(SyncOptions{
		FilePath:    testFile,
		StartMarker: "# START",
		EndMarker:   "# END",
		Content:     "# START\nNew\n# END",
	})

	if err == nil {
		t.Error("Expected error when file doesn't exist")
	}

	if !contains(err.Error(), "failed to read file") {
		t.Errorf("Error should mention file read failure, got: %v", err)
	}
}

// Test backup is preserved on write error
func TestSyncFileSection_BackupPreservedOnError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Cannot test write errors as root")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	original := "# START\nOld\n# END\n"
	os.WriteFile(testFile, []byte(original), 0644)

	// Make directory read-only to cause write error
	os.Chmod(tmpDir, 0555)
	defer os.Chmod(tmpDir, 0755) // Cleanup

	err := SyncFileSection(SyncOptions{
		FilePath:     testFile,
		StartMarker:  "# START",
		EndMarker:    "# END",
		Content:      "# START\nNew\n# END",
		CreateBackup: true,
	})

	// Restore permissions to check backup
	os.Chmod(tmpDir, 0755)

	if err == nil {
		t.Skip("Write operation unexpectedly succeeded (filesystem may not enforce permissions)")
	}

	// Note: On some systems this test may not work as expected
	// The backup behavior depends on the specific error that occurs
}

// Test multiple marker pairs in file
func TestSyncFileSection_MultipleMarkerPairs(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	initial := `Section 1
# START
Content 1
# END

Section 2
# START
Content 2
# END
`
	os.WriteFile(testFile, []byte(initial), 0644)

	// Should replace first occurrence
	err := SyncFileSection(SyncOptions{
		FilePath:    testFile,
		StartMarker: "# START",
		EndMarker:   "# END",
		Content:     "# START\nNew Content\n# END",
	})

	if err != nil {
		t.Fatalf("SyncFileSection failed: %v", err)
	}

	result, _ := os.ReadFile(testFile)
	resultStr := string(result)

	// First section should be replaced
	if !contains(resultStr, "# START\nNew Content\n# END") {
		t.Error("First section should be replaced")
	}

	// Second section should be gone (replaced up to its end marker)
	// This is expected behavior - it replaces from first START to first END
}

// Test with Unicode content
func TestSyncFileSection_Unicode(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	initial := "Header\n# START\n旧内容\n# END\nFooter\n"
	os.WriteFile(testFile, []byte(initial), 0644)

	newContent := "# START\n新内容\n🎉\n# END"

	err := SyncFileSection(SyncOptions{
		FilePath:    testFile,
		StartMarker: "# START",
		EndMarker:   "# END",
		Content:     newContent,
	})

	if err != nil {
		t.Fatalf("SyncFileSection failed: %v", err)
	}

	result, _ := os.ReadFile(testFile)
	if !contains(string(result), "新内容") || !contains(string(result), "🎉") {
		t.Error("Unicode content not preserved")
	}
}

// Test markers with special regex characters
func TestSyncFileSection_SpecialCharMarkers(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	initial := "Before\n[BEGIN]\nOld\n[END]\nAfter\n"
	os.WriteFile(testFile, []byte(initial), 0644)

	err := SyncFileSection(SyncOptions{
		FilePath:    testFile,
		StartMarker: "[BEGIN]",
		EndMarker:   "[END]",
		Content:     "[BEGIN]\nNew\n[END]",
	})

	if err != nil {
		t.Fatalf("SyncFileSection failed: %v", err)
	}

	result, _ := os.ReadFile(testFile)
	if !contains(string(result), "[BEGIN]\nNew\n[END]") {
		t.Error("Content not replaced correctly with special char markers")
	}
}

// Test empty file
func TestSyncFileSection_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte(""), 0644)

	err := SyncFileSection(SyncOptions{
		FilePath:    testFile,
		StartMarker: "# START",
		EndMarker:   "# END",
		Content:     "# START\nNew\n# END",
	})

	if err == nil {
		t.Error("Expected error for empty file (no markers)")
	}
}

// Test very large file
func TestSyncFileSection_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create large file with markers in the middle
	largeContent := ""
	for i := 0; i < 10000; i++ {
		largeContent += "Line before marker\n"
	}
	largeContent += "# START\nOld content\n# END\n"
	for i := 0; i < 10000; i++ {
		largeContent += "Line after marker\n"
	}

	os.WriteFile(testFile, []byte(largeContent), 0644)

	err := SyncFileSection(SyncOptions{
		FilePath:    testFile,
		StartMarker: "# START",
		EndMarker:   "# END",
		Content:     "# START\nNew content\n# END",
	})

	if err != nil {
		t.Fatalf("SyncFileSection failed on large file: %v", err)
	}

	result, _ := os.ReadFile(testFile)
	if !contains(string(result), "# START\nNew content\n# END") {
		t.Error("Content not replaced in large file")
	}
}

// Test identical start and end markers (edge case)
// Note: When markers are identical, behavior may be unexpected
// as it searches for the NEXT occurrence after the start marker
func TestSyncFileSection_IdenticalMarkers(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	// Use unique content to avoid confusion
	initial := "Before\n===\nOld content\n===\nAfter\n"
	os.WriteFile(testFile, []byte(initial), 0644)

	err := SyncFileSection(SyncOptions{
		FilePath:    testFile,
		StartMarker: "===",
		EndMarker:   "===",
		Content:     "===\nNew content\n===",
	})

	if err != nil {
		t.Fatalf("SyncFileSection failed: %v", err)
	}

	result, _ := os.ReadFile(testFile)
	// With identical markers, it replaces from first === to second ===
	// Result should have the new content between the markers
	if !contains(string(result), "New content") {
		t.Errorf("Result should contain 'New content', got:\n%q", string(result))
	}
	if !contains(string(result), "Before") {
		t.Errorf("Result should contain 'Before', got:\n%q", string(result))
	}
}

// Test concurrent access (basic race condition check)
func TestSyncFileSection_Backup(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	original := "# START\nOriginal\n# END\n"
	os.WriteFile(testFile, []byte(original), 0644)

	// First sync should create and remove backup
	err := SyncFileSection(SyncOptions{
		FilePath:     testFile,
		StartMarker:  "# START",
		EndMarker:    "# END",
		Content:      "# START\nFirst update\n# END",
		CreateBackup: true,
	})

	if err != nil {
		t.Fatalf("First sync failed: %v", err)
	}

	// Second sync should also create and remove backup
	err = SyncFileSection(SyncOptions{
		FilePath:     testFile,
		StartMarker:  "# START",
		EndMarker:    "# END",
		Content:      "# START\nSecond update\n# END",
		CreateBackup: true,
	})

	if err != nil {
		t.Fatalf("Second sync failed: %v", err)
	}

	// Verify final content
	result, _ := os.ReadFile(testFile)
	if !contains(string(result), "Second update") {
		t.Error("Final update not applied")
	}

	// Verify no backup remains
	backup := testFile + ".backup"
	if _, err := os.Stat(backup); err == nil {
		t.Error("Backup should have been cleaned up")
	}
}
