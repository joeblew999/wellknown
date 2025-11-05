package banking

import (
	"fmt"
	"time"

	"github.com/joeblew999/wellknown/pkg/pb/codegen/models"
	"github.com/pocketbase/pocketbase/core"
)

// AccountService provides business logic for account management
type AccountService struct {
	app core.App
}

// NewAccountService creates a new account service
func NewAccountService(app core.App) *AccountService {
	return &AccountService{app: app}
}

// CreateAccount creates a new account for a user
func (s *AccountService) CreateAccount(userID, accountNumber, accountName, accountType, currency string, initialBalance float64) (*core.Record, error) {
	collection, err := s.app.FindCollectionByNameOrId("accounts")
	if err != nil {
		return nil, fmt.Errorf("failed to find accounts collection: %w", err)
	}

	record := core.NewRecord(collection)
	record.Set("user_id", userID)
	record.Set("account_number", accountNumber)
	record.Set("account_name", accountName)
	record.Set("account_type", accountType)
	record.Set("balance", initialBalance)
	record.Set("currency", currency)
	record.Set("is_active", true)

	if err := s.app.Save(record); err != nil {
		return nil, fmt.Errorf("failed to save account: %w", err)
	}

	return record, nil
}

// GetUserAccounts retrieves all accounts for a user
func (s *AccountService) GetUserAccounts(userID string) ([]*core.Record, error) {
	records, err := s.app.FindRecordsByFilter(
		"accounts",
		"user_id = {:userId}",
		"-created",
		-1,
		0,
		map[string]any{"userId": userID},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find user accounts: %w", err)
	}
	return records, nil
}

// GetAccount retrieves a specific account by ID
func (s *AccountService) GetAccount(accountID string) (*core.Record, error) {
	record, err := s.app.FindRecordById("accounts", accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to find account: %w", err)
	}
	return record, nil
}

// UpdateBalance updates the account balance (use with caution - should be done via transactions)
func (s *AccountService) UpdateBalance(accountID string, newBalance float64) error {
	record, err := s.GetAccount(accountID)
	if err != nil {
		return err
	}

	record.Set("balance", newBalance)
	if err := s.app.Save(record); err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	return nil
}

// DeactivateAccount marks an account as inactive
func (s *AccountService) DeactivateAccount(accountID string) error {
	record, err := s.GetAccount(accountID)
	if err != nil {
		return err
	}

	record.Set("is_active", false)
	if err := s.app.Save(record); err != nil {
		return fmt.Errorf("failed to deactivate account: %w", err)
	}

	return nil
}

// GetAccountBalance retrieves the current balance of an account
func (s *AccountService) GetAccountBalance(accountID string) (float64, error) {
	record, err := s.GetAccount(accountID)
	if err != nil {
		return 0, err
	}

	balance := record.GetFloat("balance")
	return balance, nil
}

// ValidateAccountAccess checks if a user has access to an account
func (s *AccountService) ValidateAccountAccess(accountID, userID string) error {
	record, err := s.GetAccount(accountID)
	if err != nil {
		return err
	}

	if record.GetString("user_id") != userID {
		return fmt.Errorf("unauthorized: user does not own this account")
	}

	return nil
}

// AccountSummary represents a summary of account information
type AccountSummary struct {
	ID            string    `json:"id"`
	AccountNumber string    `json:"account_number"`
	AccountName   string    `json:"account_name"`
	AccountType   string    `json:"account_type"`
	Balance       float64   `json:"balance"`
	Currency      string    `json:"currency"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

// GetAccountSummary returns a simplified account summary
func (s *AccountService) GetAccountSummary(accountID string) (*AccountSummary, error) {
	record, err := s.GetAccount(accountID)
	if err != nil {
		return nil, err
	}

	return &AccountSummary{
		ID:            record.Id,
		AccountNumber: record.GetString("account_number"),
		AccountName:   record.GetString("account_name"),
		AccountType:   record.GetString("account_type"),
		Balance:       record.GetFloat("balance"),
		Currency:      record.GetString("currency"),
		IsActive:      record.GetBool("is_active"),
		CreatedAt:     record.GetDateTime("created").Time(),
	}, nil
}

// Ensure AccountService works with generated models (for reference)
// This shows how to use the generated proxy types if needed
var _ = func() *models.Accounts {
	// Example: If you need to use generated proxy methods:
	// collection, _ := app.FindCollectionByNameOrId("accounts")
	// record := core.NewRecord(collection)
	// proxy := models.NewAccountsRecord(record)
	// proxy.SetAccountName("My Account")
	return nil
}()
