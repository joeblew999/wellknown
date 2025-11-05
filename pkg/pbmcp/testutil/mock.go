package testutil

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

// NewTestApp creates a new PocketBase app for testing
// Uses in-memory SQLite database
func NewTestApp() (core.App, error) {
	app, err := tests.NewTestApp()
	if err != nil {
		return nil, err
	}

	// Create test collections
	if err := CreateTestCollections(app); err != nil {
		return nil, err
	}

	return app, nil
}

// CreateTestCollections creates standard test collections
func CreateTestCollections(app core.App) error {
	// Create a users collection
	collection := core.NewCollection(core.CollectionTypeBase, "test_users")

	collection.Fields.Add(&core.TextField{
		Name:     "name",
		Required: true,
	})

	collection.Fields.Add(&core.TextField{
		Name:     "email",
		Required: true,
	})

	collection.Fields.Add(&core.NumberField{
		Name: "age",
	})

	// Set public rules for testing
	empty := ""
	collection.ListRule = &empty
	collection.ViewRule = &empty
	collection.CreateRule = &empty
	collection.UpdateRule = &empty
	collection.DeleteRule = &empty

	if err := app.Save(collection); err != nil {
		return err
	}

	// Create a posts collection
	collection2 := core.NewCollection(core.CollectionTypeBase, "test_posts")

	collection2.Fields.Add(&core.TextField{
		Name:     "title",
		Required: true,
	})

	collection2.Fields.Add(&core.TextField{
		Name: "content",
	})

	collection2.Fields.Add(&core.BoolField{
		Name: "published",
	})

	// Set public rules for testing
	collection2.ListRule = &empty
	collection2.ViewRule = &empty
	collection2.CreateRule = &empty
	collection2.UpdateRule = &empty
	collection2.DeleteRule = &empty

	if err := app.Save(collection2); err != nil {
		return err
	}

	return nil
}

// CreateTestRecord creates a test record in a collection
func CreateTestRecord(app core.App, collection string, data map[string]any) (*core.Record, error) {
	c, err := app.FindCollectionByNameOrId(collection)
	if err != nil {
		return nil, err
	}

	record := core.NewRecord(c)
	for key, value := range data {
		record.Set(key, value)
	}

	if err := app.Save(record); err != nil {
		return nil, err
	}

	return record, nil
}

// CleanupTestApp closes and cleans up the test app
func CleanupTestApp(app core.App) {
	if testApp, ok := app.(*tests.TestApp); ok {
		testApp.Cleanup()
	}
}
