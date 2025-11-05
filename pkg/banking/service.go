package banking

import (
	"github.com/pocketbase/pocketbase/core"
)

// Service provides a unified interface for all banking operations
type Service struct {
	Accounts     *AccountService
	Transactions *TransactionService
}

// NewService creates a new banking service with all sub-services
func NewService(app core.App) *Service {
	accountService := NewAccountService(app)
	transactionService := NewTransactionService(app, accountService)

	return &Service{
		Accounts:     accountService,
		Transactions: transactionService,
	}
}

// Example usage:
//
//   banking := banking.NewService(app)
//
//   // Create account
//   account, err := banking.Accounts.CreateAccount(
//       userID, "ACC001", "Savings", "savings", "USD", 1000.00,
//   )
//
//   // Create transaction
//   tx, err := banking.Transactions.CreateTransaction(banking.CreateTransactionParams{
//       AccountID:       account.Id,
//       TransactionType: banking.TransactionTypeDebit,
//       Amount:          50.00,
//       Currency:        "USD",
//       Description:     "Coffee",
//       Category:        "food",
//       TransactionDate: time.Now(),
//       IsPending:       false,
//   })
