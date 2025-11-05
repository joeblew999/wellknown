package pb_migrations

import (
	"fmt"
	"log"
	"os"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Read admin credentials from environment
		email := os.Getenv("PB_ADMIN_EMAIL")
		password := os.Getenv("PB_ADMIN_PASSWORD")

		// Skip if credentials not provided (allows manual admin creation)
		if email == "" || password == "" {
			log.Println("⚠️  PB_ADMIN_EMAIL or PB_ADMIN_PASSWORD not set - skipping admin upsert")
			log.Println("   Admin must be created manually via UI: /_/")
			return nil
		}

		log.Printf("🔐 Ensuring admin user exists: %s", email)

		// Get superusers collection
		superusersCol, err := app.FindCachedCollectionByNameOrId(core.CollectionNameSuperusers)
		if err != nil {
			return fmt.Errorf("failed to fetch superusers collection: %w", err)
		}

		// Try to find existing superuser by email
		superuser, err := app.FindAuthRecordByEmail(superusersCol, email)
		if err != nil {
			// Superuser doesn't exist - create new one
			log.Printf("   Creating new admin user: %s", email)
			superuser = core.NewRecord(superusersCol)
		} else {
			// Superuser exists - update password to match env var
			log.Printf("   Updating existing admin user: %s", email)
		}

		// Set/update email and password
		superuser.SetEmail(email)
		superuser.SetPassword(password)

		// Save superuser record
		if err := app.Save(superuser); err != nil {
			return fmt.Errorf("failed to save admin user: %w", err)
		}

		log.Printf("✅ Admin user ready: %s", email)
		return nil
	}, nil)
}
