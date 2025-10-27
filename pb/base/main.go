package main

import (
	"log"
	"os"

	wellknown "github.com/joeblew999/wellknown/pb"
)

func main() {
	// Check required environment variables
	requiredEnvVars := []string{
		"GOOGLE_CLIENT_ID",
		"GOOGLE_CLIENT_SECRET",
		"GOOGLE_REDIRECT_URL",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("Missing required environment variable: %s", envVar)
		}
	}

	// Create wellknown Pocketbase app
	app := wellknown.New()

	log.Println("🚀 Starting Wellknown Pocketbase server...")
	log.Println("📧 Access admin UI: http://localhost:8090/_/")
	log.Println("🔐 Google OAuth: http://localhost:8090/auth/google")
	log.Println("")
	log.Println("Required environment variables:")
	log.Println("  GOOGLE_CLIENT_ID")
	log.Println("  GOOGLE_CLIENT_SECRET")
	log.Println("  GOOGLE_REDIRECT_URL")

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
