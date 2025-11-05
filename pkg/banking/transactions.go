package banking

import (
	"fmt"
	"time"

	"github.com/joeblew999/wellknown/pkg/pb/codegen/models"
	"github.com/pocketbase/pocketbase/core"
)

// TransactionType constants
const (
	TransactionTypeDebit  = "debit"
	TransactionTypeCredit = "credit"
)

// TransactionService provides business logic for transaction management
type TransactionService struct {
	app            core.App
	accountService *AccountService
}

// NewTransactionService creates a new transaction service
func NewTransactionService(app core.App, accountService *AccountService) *TransactionService {
	return &TransactionService{
		app:            app,
		accountService: accountService,
	}
}

// CreateTransactionParams holds parameters for creating a transaction
type CreateTransactionParams struct {
	AccountID       string
	TransactionType string
	Amount          float64
	Currency        string
	Description     string
	Category        string
	Merchant        string
	Reference       string
	TransactionDate time.Time
	IsPending       bool
}

// CreateTransaction creates a new transaction and updates account balance
func (s *TransactionService) CreateTransaction(params CreateTransactionParams) (*core.Record, error) {
	// Validate transaction type
	if params.TransactionType != TransactionTypeDebit && params.TransactionType != TransactionTypeCredit {
		return nil, fmt.Errorf("invalid transaction type: must be 'debit' or 'credit'")
	}

	// Validate amount
	if params.Amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	// Get account to validate it exists and get current balance
	account, err := s.accountService.GetAccount(params.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Check if account is active
	if !account.GetBool("is_active") {
		return nil, fmt.Errorf("account is inactive")
	}

	currentBalance := account.GetFloat("balance")

	// Calculate new balance
	var newBalance float64
	if params.TransactionType == TransactionTypeDebit {
		newBalance = currentBalance - params.Amount
	} else {
		newBalance = currentBalance + params.Amount
	}

	// Check for overdraft (optional business rule)
	if newBalance < 0 {
		return nil, fmt.Errorf("insufficient funds: would result in negative balance")
	}

	// Create transaction record
	collection, err := s.app.FindCollectionByNameOrId("transactions")
	if err != nil {
		return nil, fmt.Errorf("failed to find transactions collection: %w", err)
	}

	record := core.NewRecord(collection)
	record.Set("account_id", params.AccountID)
	record.Set("transaction_type", params.TransactionType)
	record.Set("amount", params.Amount)
	record.Set("currency", params.Currency)
	record.Set("description", params.Description)
	record.Set("category", params.Category)
	record.Set("merchant", params.Merchant)
	record.Set("reference", params.Reference)
	record.Set("transaction_date", params.TransactionDate)
	record.Set("is_pending", params.IsPending)

	// Save transaction and update balance in a "transaction" (note: PocketBase doesn't have true DB transactions)
	// In production, you'd want to use DB transactions or implement saga pattern
	if err := s.app.Save(record); err != nil {
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	// Update account balance only if not pending
	if !params.IsPending {
		if err := s.accountService.UpdateBalance(params.AccountID, newBalance); err != nil {
			// In production, you'd want to rollback the transaction here
			return nil, fmt.Errorf("failed to update balance: %w", err)
		}
	}

	return record, nil
}

// GetAccountTransactions retrieves all transactions for an account
func (s *TransactionService) GetAccountTransactions(accountID string, limit int, offset int) ([]*core.Record, error) {
	records, err := s.app.FindRecordsByFilter(
		"transactions",
		"account_id = {:accountId}",
		"-transaction_date",
		limit,
		offset,
		map[string]any{"accountId": accountID},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find transactions: %w", err)
	}
	return records, nil
}

// GetTransaction retrieves a specific transaction by ID
func (s *TransactionService) GetTransaction(transactionID string) (*core.Record, error) {
	record, err := s.app.FindRecordById("transactions", transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find transaction: %w", err)
	}
	return record, nil
}

// GetTransactionsByDateRange retrieves transactions within a date range
func (s *TransactionService) GetTransactionsByDateRange(accountID string, startDate, endDate time.Time) ([]*core.Record, error) {
	records, err := s.app.FindRecordsByFilter(
		"transactions",
		"account_id = {:accountId} && transaction_date >= {:startDate} && transaction_date <= {:endDate}",
		"-transaction_date",
		-1,
		0,
		map[string]any{
			"accountId": accountID,
			"startDate": startDate.Format(time.RFC3339),
			"endDate":   endDate.Format(time.RFC3339),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find transactions by date range: %w", err)
	}
	return records, nil
}

// GetTransactionsByCategory retrieves transactions by category
func (s *TransactionService) GetTransactionsByCategory(accountID, category string) ([]*core.Record, error) {
	records, err := s.app.FindRecordsByFilter(
		"transactions",
		"account_id = {:accountId} && category = {:category}",
		"-transaction_date",
		-1,
		0,
		map[string]any{
			"accountId": accountID,
			"category":  category,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find transactions by category: %w", err)
	}
	return records, nil
}

// SettlePendingTransaction settles a pending transaction and updates the balance
func (s *TransactionService) SettlePendingTransaction(transactionID string) error {
	transaction, err := s.GetTransaction(transactionID)
	if err != nil {
		return err
	}

	// Check if already settled
	if !transaction.GetBool("is_pending") {
		return fmt.Errorf("transaction is already settled")
	}

	accountID := transaction.GetString("account_id")
	amount := transaction.GetFloat("amount")
	transactionType := transaction.GetString("transaction_type")

	// Get current account balance
	account, err := s.accountService.GetAccount(accountID)
	if err != nil {
		return err
	}

	currentBalance := account.GetFloat("balance")
	var newBalance float64
	if transactionType == TransactionTypeDebit {
		newBalance = currentBalance - amount
	} else {
		newBalance = currentBalance + amount
	}

	// Check for overdraft
	if newBalance < 0 {
		return fmt.Errorf("insufficient funds to settle transaction")
	}

	// Mark as settled
	transaction.Set("is_pending", false)
	if err := s.app.Save(transaction); err != nil {
		return fmt.Errorf("failed to settle transaction: %w", err)
	}

	// Update balance
	if err := s.accountService.UpdateBalance(accountID, newBalance); err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	return nil
}

// TransactionSummary represents a summary of transaction information
type TransactionSummary struct {
	ID              string    `json:"id"`
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
	CreatedAt       time.Time `json:"created_at"`
}

// GetTransactionSummary returns a simplified transaction summary
func (s *TransactionService) GetTransactionSummary(transactionID string) (*TransactionSummary, error) {
	record, err := s.GetTransaction(transactionID)
	if err != nil {
		return nil, err
	}

	return &TransactionSummary{
		ID:              record.Id,
		AccountID:       record.GetString("account_id"),
		TransactionType: record.GetString("transaction_type"),
		Amount:          record.GetFloat("amount"),
		Currency:        record.GetString("currency"),
		Description:     record.GetString("description"),
		Category:        record.GetString("category"),
		Merchant:        record.GetString("merchant"),
		Reference:       record.GetString("reference"),
		TransactionDate: record.GetDateTime("transaction_date").Time(),
		IsPending:       record.GetBool("is_pending"),
		CreatedAt:       record.GetDateTime("created").Time(),
	}, nil
}

// CalculateAccountBalance recalculates account balance from all transactions
// Useful for reconciliation
func (s *TransactionService) CalculateAccountBalance(accountID string) (float64, error) {
	transactions, err := s.GetAccountTransactions(accountID, -1, 0)
	if err != nil {
		return 0, err
	}

	var balance float64
	for _, tx := range transactions {
		// Only count settled transactions
		if !tx.GetBool("is_pending") {
			amount := tx.GetFloat("amount")
			txType := tx.GetString("transaction_type")

			if txType == TransactionTypeCredit {
				balance += amount
			} else if txType == TransactionTypeDebit {
				balance -= amount
			}
		}
	}

	return balance, nil
}

// Ensure TransactionService works with generated models (for reference)
var _ = func() *models.Transactions {
	// Example: If you need to use generated proxy methods:
	// collection, _ := app.FindCollectionByNameOrId("transactions")
	// record := core.NewRecord(collection)
	// proxy := models.NewTransactionsRecord(record)
	// proxy.SetAmount(100.00)
	return nil
}()
