package wellknown

import (
	"fmt"
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// registerBankingRoutes sets up banking API endpoints
func registerBankingRoutes(wk *Wellknown, e *core.ServeEvent) error {
	wk.Logger().Info("🔍 DEBUG: Starting banking routes registration...")
	wk.Logger().Info(fmt.Sprintf("🔍 DEBUG: Router object: %p", e.Router))

	// List accounts for a user
	wk.Logger().Info("🔍 DEBUG: Registering GET /api/banking/accounts...")
	e.Router.GET("/api/banking/accounts", func(c *core.RequestEvent) error {
		userID := c.Request.URL.Query().Get("user_id")
		if userID == "" {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"error": "user_id query parameter required",
			})
		}

		records, err := wk.FindRecordsByFilter(
			"accounts",
			"user_id = {:userID}",
			"-created",
			100,
			0,
			map[string]any{"userID": userID},
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"accounts": records,
		})
	})

	// Get account details
	wk.Logger().Info("🔍 DEBUG: Registering GET /api/banking/accounts/:id...")
	e.Router.GET("/api/banking/accounts/:id", func(c *core.RequestEvent) error {
		id := c.Request.PathValue("id")

		record, err := wk.FindRecordById("accounts", id)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]any{
				"error": "Account not found",
			})
		}

		return c.JSON(http.StatusOK, record)
	})

	// List transactions for an account
	e.Router.GET("/api/banking/accounts/:id/transactions", func(c *core.RequestEvent) error {
		accountID := c.Request.PathValue("id")

		records, err := wk.FindRecordsByFilter(
			"transactions",
			"account_id = {:accountID}",
			"-transaction_date",
			100,
			0,
			map[string]any{"accountID": accountID},
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"transactions": records,
		})
	})

	// Create account
	wk.Logger().Info("🔍 DEBUG: Registering POST /api/banking/accounts...")
	e.Router.POST("/api/banking/accounts", func(c *core.RequestEvent) error {
		data := &struct {
			UserID        string  `json:"user_id"`
			AccountNumber string  `json:"account_number"`
			AccountName   string  `json:"account_name"`
			AccountType   string  `json:"account_type"`
			Balance       float64 `json:"balance"`
			Currency      string  `json:"currency"`
			IsActive      bool    `json:"is_active"`
		}{}

		if err := c.BindBody(data); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"error": "Invalid request body",
			})
		}

		collection, err := wk.FindCollectionByNameOrId("accounts")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{
				"error": "Collection not found",
			})
		}

		record := core.NewRecord(collection)
		record.Set("user_id", data.UserID)
		record.Set("account_number", data.AccountNumber)
		record.Set("account_name", data.AccountName)
		record.Set("account_type", data.AccountType)
		record.Set("balance", data.Balance)
		record.Set("currency", data.Currency)
		record.Set("is_active", data.IsActive)

		if err := wk.Save(record); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusCreated, record)
	})

	// Create transaction
	e.Router.POST("/api/banking/transactions", func(c *core.RequestEvent) error {
		data := &struct {
			AccountID       string  `json:"account_id"`
			TransactionType string  `json:"transaction_type"`
			Amount          float64 `json:"amount"`
			Currency        string  `json:"currency"`
			Description     string  `json:"description"`
			Category        string  `json:"category"`
			Merchant        string  `json:"merchant"`
			Reference       string  `json:"reference"`
			TransactionDate string  `json:"transaction_date"`
			IsPending       bool    `json:"is_pending"`
		}{}

		if err := c.BindBody(data); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"error": "Invalid request body",
			})
		}

		collection, err := wk.FindCollectionByNameOrId("transactions")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{
				"error": "Collection not found",
			})
		}

		record := core.NewRecord(collection)
		record.Set("account_id", data.AccountID)
		record.Set("transaction_type", data.TransactionType)
		record.Set("amount", data.Amount)
		record.Set("currency", data.Currency)
		record.Set("description", data.Description)
		record.Set("category", data.Category)
		record.Set("merchant", data.Merchant)
		record.Set("reference", data.Reference)
		record.Set("transaction_date", data.TransactionDate)
		record.Set("is_pending", data.IsPending)

		if err := wk.Save(record); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{
				"error": err.Error(),
			})
		}

		// Update account balance
		account, err := wk.FindRecordById("accounts", data.AccountID)
		if err == nil {
			currentBalance := account.GetFloat("balance")
			newBalance := currentBalance

			switch data.TransactionType {
			case "credit":
				newBalance += data.Amount
			case "debit":
				newBalance -= data.Amount
			}

			account.Set("balance", newBalance)
			wk.Save(account)
		}

		return c.JSON(http.StatusCreated, record)
	})

	wk.Logger().Info("✅ Banking API routes registered")
	return e.Next()
}
