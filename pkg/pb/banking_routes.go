package wellknown

import (
	"log"
	"net/http"
	"time"

	"github.com/joeblew999/wellknown/pkg/banking"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterBankingRoutes registers all banking-related routes
func RegisterBankingRoutes(wk *Wellknown, e *core.ServeEvent, registry *RouteRegistry) {
	// Pre-flight check: Validate required collections exist
	requiredCollections := []string{"accounts", "transactions"}
	for _, collectionName := range requiredCollections {
		if _, err := wk.FindCollectionByNameOrId(collectionName); err != nil {
			log.Printf("⚠️  Banking routes NOT registered: collection '%s' not found (migrations may not have run)", collectionName)
			log.Printf("   Run 'go run . migrate up' to create required collections")
			return // Skip banking routes registration
		}
	}

	log.Println("✅ Banking routes: Pre-flight checks passed")

	// Create banking service
	service := banking.NewService(wk.PocketBase)

	// Create route handler for banking domain
	handler := NewRouteHandler(registry, "Banking", e)

	// Register routes
	handler.GET("/api/banking/accounts", func(c *core.RequestEvent) error {
		return handleListAccounts(c, service)
	}, WithAuth(), WithDescription("List all accounts for authenticated user"))

	handler.POST("/api/banking/accounts", func(c *core.RequestEvent) error {
		return handleCreateAccount(c, service)
	}, WithAuth(), WithDescription("Create a new bank account"))

	handler.GET("/api/banking/accounts/:id", func(c *core.RequestEvent) error {
		return handleGetAccount(c, service)
	}, WithAuth(), WithDescription("Get account details by ID"))

	handler.GET("/api/banking/accounts/:id/transactions", func(c *core.RequestEvent) error {
		return handleListTransactions(c, service)
	}, WithAuth(), WithDescription("List transactions for an account"))

	handler.POST("/api/banking/transactions", func(c *core.RequestEvent) error {
		return handleCreateTransaction(c, service)
	}, WithAuth(), WithDescription("Create a new transaction"))
}

// handleListAccounts lists all accounts for the authenticated user
func handleListAccounts(c *core.RequestEvent, service *banking.Service) error {
	// Auth middleware ensures c.Auth is populated
	userRecord := c.Auth

	// Get user's accounts
	accounts, err := service.Accounts.GetUserAccounts(userRecord.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Convert to summaries
	summaries := make([]banking.AccountSummary, 0, len(accounts))
	for _, acc := range accounts {
		summary, err := service.Accounts.GetAccountSummary(acc.Id)
		if err != nil {
			continue
		}
		summaries = append(summaries, *summary)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"accounts": summaries,
		"count":    len(summaries),
	})
}

// CreateAccountRequest represents the request body for creating an account
type CreateAccountRequest struct {
	AccountNumber string  `json:"account_number"`
	AccountName   string  `json:"account_name"`
	AccountType   string  `json:"account_type"`
	Currency      string  `json:"currency"`
	InitialBalance float64 `json:"initial_balance"`
}

// handleCreateAccount creates a new bank account
func handleCreateAccount(c *core.RequestEvent, service *banking.Service) error {
	// Auth middleware ensures c.Auth is populated
	userRecord := c.Auth

	// Parse request
	var req CreateAccountRequest
	if err := c.BindBody(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.AccountNumber == "" || req.AccountName == "" || req.AccountType == "" || req.Currency == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing required fields",
		})
	}

	// Create account
	account, err := service.Accounts.CreateAccount(
		userRecord.Id,
		req.AccountNumber,
		req.AccountName,
		req.AccountType,
		req.Currency,
		req.InitialBalance,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Get summary
	summary, err := service.Accounts.GetAccountSummary(account.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, summary)
}

// handleGetAccount gets account details by ID
func handleGetAccount(c *core.RequestEvent, service *banking.Service) error {
	// Auth middleware ensures c.Auth is populated
	userRecord := c.Auth

	accountID := c.Request.PathValue("id")
	if accountID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Account ID required",
		})
	}

	// Validate user owns account
	if err := service.Accounts.ValidateAccountAccess(accountID, userRecord.Id); err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "Access denied",
		})
	}

	// Get account summary
	summary, err := service.Accounts.GetAccountSummary(accountID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Account not found",
		})
	}

	return c.JSON(http.StatusOK, summary)
}

// handleListTransactions lists transactions for an account
func handleListTransactions(c *core.RequestEvent, service *banking.Service) error {
	// Auth middleware ensures c.Auth is populated
	userRecord := c.Auth

	accountID := c.Request.PathValue("id")
	if accountID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Account ID required",
		})
	}

	// Validate user owns account
	if err := service.Accounts.ValidateAccountAccess(accountID, userRecord.Id); err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "Access denied",
		})
	}

	// Get transactions
	transactions, err := service.Transactions.GetAccountTransactions(accountID, 50, 0)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Convert to summaries
	summaries := make([]banking.TransactionSummary, 0, len(transactions))
	for _, tx := range transactions {
		summary, err := service.Transactions.GetTransactionSummary(tx.Id)
		if err != nil {
			continue
		}
		summaries = append(summaries, *summary)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"transactions": summaries,
		"count":        len(summaries),
	})
}

// CreateTransactionRequest represents the request body for creating a transaction
type CreateTransactionRequest struct {
	AccountID       string    `json:"account_id"`
	TransactionType string    `json:"transaction_type"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	Description     string    `json:"description"`
	Category        string    `json:"category"`
	Merchant        string    `json:"merchant"`
	Reference       string    `json:"reference"`
	TransactionDate time.Time `json:"transaction_date"`
	IsPending       bool      `json:"is_pending"`
}

// handleCreateTransaction creates a new transaction
func handleCreateTransaction(c *core.RequestEvent, service *banking.Service) error {
	// Auth middleware ensures c.Auth is populated
	userRecord := c.Auth

	// Parse request
	var req CreateTransactionRequest
	if err := c.BindBody(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.AccountID == "" || req.TransactionType == "" || req.Amount <= 0 || req.Currency == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing or invalid required fields",
		})
	}

	// Validate user owns account
	if err := service.Accounts.ValidateAccountAccess(req.AccountID, userRecord.Id); err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "Access denied",
		})
	}

	// Set default transaction date if not provided
	if req.TransactionDate.IsZero() {
		req.TransactionDate = time.Now()
	}

	// Create transaction
	transaction, err := service.Transactions.CreateTransaction(banking.CreateTransactionParams{
		AccountID:       req.AccountID,
		TransactionType: req.TransactionType,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Description:     req.Description,
		Category:        req.Category,
		Merchant:        req.Merchant,
		Reference:       req.Reference,
		TransactionDate: req.TransactionDate,
		IsPending:       req.IsPending,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Get summary
	summary, err := service.Transactions.GetTransactionSummary(transaction.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, summary)
}
