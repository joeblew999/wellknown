package pb_migrations

import (
	"github.com/pocketbase/pocketbase/core"
)

func init() {
	core.AppMigrations.Register(
		// Up: Create google_tokens collection
		func(txApp core.App) error {
			collection := core.NewBaseCollection("google_tokens")

			// Add only custom fields (NOT system fields like email!)
			collection.Fields.Add(
				&core.TextField{
					Name:     "user_id",
					Required: true,
				},
				&core.TextField{
					Name:     "access_token",
					Required: true,
				},
				&core.TextField{
					Name:     "refresh_token",
					Required: false,
				},
				&core.TextField{
					Name:     "token_type",
					Required: false,
				},
				&core.DateField{
					Name:     "expiry",
					Required: false,
				},
			)

			// Note: NO email field - that's for auth collections only!
			// If you need user association, use user_id (relation field)

			return txApp.Save(collection)
		},

		// Down: Remove google_tokens collection
		func(txApp core.App) error {
			collection, err := txApp.FindCollectionByNameOrId("google_tokens")
			if err != nil {
				return err
			}
			return txApp.Delete(collection)
		},
	)
}
