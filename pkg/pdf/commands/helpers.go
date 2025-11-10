package commands

import (
	"fmt"
	"os"
	"path/filepath"
)

// DetermineOutputPath determines the final output path for a file
// If outputDir is empty, uses baseName from dataPath with suffix
// If outputDir is a directory, creates file in that directory
// If outputDir is a file path, uses it directly
func DetermineOutputPath(dataPath, outputDir, suffix string) string {
	if outputDir == "" {
		// Use current directory with suffix
		base := filepath.Base(dataPath)
		name := base[:len(base)-len(filepath.Ext(base))]
		return name + suffix
	}

	// Check if outputDir is a directory
	if info, err := os.Stat(outputDir); err == nil && info.IsDir() {
		// It's a directory, create filename in it
		base := filepath.Base(dataPath)
		name := base[:len(base)-len(filepath.Ext(base))]
		return filepath.Join(outputDir, name+suffix)
	}

	// Treat as file path
	return outputDir
}

// EnsureOutputDir ensures the directory for the given file path exists
func EnsureOutputDir(filePath string) error {
	outputDir := filepath.Dir(filePath)
	if err := os.MkdirAll(outputDir, DefaultDirPerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	return nil
}

// EmitStageError emits a standardized error event with stage information
func EmitStageError(eventType EventType, stage string, err error, context map[string]interface{}) {
	if context == nil {
		context = make(map[string]interface{})
	}
	context["stage"] = stage
	EmitError(eventType, err, context)
}

// BaseNameWithoutExt returns the filename without path and extension
func BaseNameWithoutExt(path string) string {
	base := filepath.Base(path)
	return base[:len(base)-len(filepath.Ext(base))]
}
