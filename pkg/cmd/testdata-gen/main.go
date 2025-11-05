// Package testdatagen: CLI wrapper for pkg/testgen
//
// This is a thin CLI wrapper around the testgen package.
// All the actual logic lives in pkg/testgen for reusability.
package testdatagen

import (
	"flag"
	"log"

	"github.com/joeblew999/wellknown/pkg/testgen"
)

// Main is the entry point for the testdata-gen service
// args contains the command-line arguments after the service name
func Main(args []string) {
	fs := flag.NewFlagSet("gen-testdata", flag.ExitOnError)
	outputDir := fs.String("output", "tests/e2e/generated", "Output directory")
	verbose := fs.Bool("v", false, "Verbose logging")
	fs.Parse(args)

	log.Println("🔧 Generating schema-validated test data...")

	opts := testgen.GenerateOptions{
		OutputDir: *outputDir,
		Verbose:   *verbose,
	}

	if err := testgen.Generate(opts); err != nil {
		log.Fatalf("❌ Generation failed: %v", err)
	}

	log.Println("✨ Complete!")
}
