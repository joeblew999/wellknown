package wellknown

import (
	"log"

	"github.com/pocketbase/pocketbase/core"
)

// setupCollections creates required collections if they don't exist
func setupCollections(wk *Wellknown) error {
	// Check if google_tokens collection exists
	_, err := wk.FindCollectionByNameOrId("google_tokens")
	if err != nil {
		// Collection doesn't exist, create it
		log.Println("Creating google_tokens collection...")
		if err := createGoogleTokensCollection(wk); err != nil {
			return err
		}
	}

	return nil
}

// createGoogleTokensCollection creates the collection for storing Google OAuth tokens
func createGoogleTokensCollection(wk *Wellknown) error {
	collection := core.NewBaseCollection("google_tokens")

	// Add fields
	collection.Schema.AddField(&core.TextField{
		Name:     "user_id",
		Required: true,
	})

	collection.Schema.AddField(&core.TextField{
		Name:     "access_token",
		Required: true,
	})

	collection.Schema.AddField(&core.TextField{
		Name:     "refresh_token",
		Required: false,
	})

	collection.Schema.AddField(&core.TextField{
		Name:     "token_type",
		Required: false,
	})

	collection.Schema.AddField(&core.DateField{
		Name:     "expiry",
		Required: false,
	})

	collection.Schema.AddField(&core.EmailField{
		Name:     "email",
		Required: false,
	})

	// Save collection
	if err := wk.Save(collection); err != nil {
		return err
	}

	log.Println("✅ google_tokens collection created")
	return nil
}
