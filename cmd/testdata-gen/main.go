// cmd/testdata-gen: CLI wrapper for pkg/testgen
//
// This is a thin CLI wrapper around the testgen package.
// All the actual logic lives in pkg/testgen for reusability.
package main

import (
	"flag"
	"log"

	"github.com/joeblew999/wellknown/pkg/testgen"
)

func main() {
	outputDir := flag.String("output", "tests/e2e/generated", "Output directory")
	verbose := flag.Bool("v", false, "Verbose logging")
	flag.Parse()

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
