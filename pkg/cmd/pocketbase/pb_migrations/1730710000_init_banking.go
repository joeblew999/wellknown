package pb_migrations

import (
	"github.com/pocketbase/pocketbase/core"
)

func init() {
	core.AppMigrations.Register(
		// Up: Create banking collections
		func(txApp core.App) error {
			// Create accounts collection
			accounts := core.NewBaseCollection("accounts")
			accounts.Fields.Add(
				&core.TextField{
					Name:     "user_id",
					Required: true,
				},
				&core.TextField{
					Name:     "account_number",
					Required: true,
				},
				&core.TextField{
					Name:     "account_name",
					Required: true,
				},
				&core.TextField{
					Name:     "account_type",
					Required: true, // checking, savings, credit
				},
				&core.NumberField{
					Name:     "balance",
					Required: true,
				},
				&core.TextField{
					Name:     "currency",
					Required: true,
				},
				&core.BoolField{
					Name: "is_active",
				},
			)
			if err := txApp.Save(accounts); err != nil {
				return err
			}

			// Create transactions collection
			transactions := core.NewBaseCollection("transactions")
			transactions.Fields.Add(
				&core.RelationField{
					Name:         "account_id",
					Required:     true,
					CollectionId: accounts.Id,
				},
				&core.TextField{
					Name:     "transaction_type",
					Required: true, // debit, credit, transfer
				},
				&core.NumberField{
					Name:     "amount",
					Required: true,
				},
				&core.TextField{
					Name:     "currency",
					Required: true,
				},
				&core.TextField{
					Name:     "description",
					Required: false,
				},
				&core.TextField{
					Name:     "category",
					Required: false, // groceries, utilities, salary, etc.
				},
				&core.TextField{
					Name:     "merchant",
					Required: false,
				},
				&core.TextField{
					Name:     "reference",
					Required: false,
				},
				&core.DateField{
					Name:     "transaction_date",
					Required: true,
				},
				&core.BoolField{
					Name: "is_pending",
				},
			)
			return txApp.Save(transactions)
		},

		// Down: Remove banking collections
		func(txApp core.App) error {
			// Delete in reverse order due to relations
			transactions, err := txApp.FindCollectionByNameOrId("transactions")
			if err == nil {
				if err := txApp.Delete(transactions); err != nil {
					return err
				}
			}

			accounts, err := txApp.FindCollectionByNameOrId("accounts")
			if err == nil {
				return txApp.Delete(accounts)
			}
			return err
		},
	)
}
