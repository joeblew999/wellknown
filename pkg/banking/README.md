# Banking Domain

This package provides business logic for banking operations, including account management and transaction processing.

## Architecture

The banking domain follows a **domain-driven design** approach, separating business logic from the generated PocketBase models:

```
pkg/
  banking/              # Banking domain (business logic)
    service.go          # Main banking service
    accounts.go         # Account management
    transactions.go     # Transaction processing
  pb/codegen/models/    # Generated models (all collections)
```

## Why This Structure?

- **Separation of Concerns**: Business logic is separated from generated code
- **Domain Organization**: Related functionality is grouped by domain (banking, auth, etc.)
- **Maintainable**: Generated code stays in one place; business logic is organized by purpose
- **Testable**: Domain services can be tested independently of PocketBase internals

## Usage

### Initialize Banking Service

```go
import (
    "github.com/joeblew999/wellknown/pkg/banking"
    "github.com/pocketbase/pocketbase"
)

// In your PocketBase app initialization
app := pocketbase.New()
bankingService := banking.NewService(app)
```

### Account Operations

```go
// Create account
account, err := bankingService.Accounts.CreateAccount(
    "user123",           // User ID
    "ACC001",            // Account number
    "Main Checking",     // Account name
    "checking",          // Account type
    "USD",               // Currency
    1000.00,             // Initial balance
)

// Get user accounts
accounts, err := bankingService.Accounts.GetUserAccounts("user123")

// Get account balance
balance, err := bankingService.Accounts.GetAccountBalance(account.Id)

// Deactivate account
err := bankingService.Accounts.DeactivateAccount(account.Id)
```

### Transaction Operations

```go
import "time"

// Create a debit transaction (withdrawal)
tx, err := bankingService.Transactions.CreateTransaction(banking.CreateTransactionParams{
    AccountID:       account.Id,
    TransactionType: banking.TransactionTypeDebit,
    Amount:          50.00,
    Currency:        "USD",
    Description:     "Coffee at Starbucks",
    Category:        "food",
    Merchant:        "Starbucks",
    Reference:       "TXN001",
    TransactionDate: time.Now(),
    IsPending:       false,
})

// Create a credit transaction (deposit)
tx, err := bankingService.Transactions.CreateTransaction(banking.CreateTransactionParams{
    AccountID:       account.Id,
    TransactionType: banking.TransactionTypeCredit,
    Amount:          500.00,
    Currency:        "USD",
    Description:     "Salary deposit",
    Category:        "income",
    TransactionDate: time.Now(),
    IsPending:       false,
})

// Get account transactions
transactions, err := bankingService.Transactions.GetAccountTransactions(account.Id, 50, 0)

// Get transactions by date range
startDate := time.Now().AddDate(0, -1, 0) // 1 month ago
endDate := time.Now()
transactions, err := bankingService.Transactions.GetTransactionsByDateRange(
    account.Id, startDate, endDate,
)

// Settle pending transaction
err := bankingService.Transactions.SettlePendingTransaction(txID)
```

## Collections

The banking domain uses these PocketBase collections:

### `accounts`
- `user_id` (relation to users)
- `account_number` (text)
- `account_name` (text)
- `account_type` (text: checking, savings, etc.)
- `balance` (number)
- `currency` (text: USD, EUR, etc.)
- `is_active` (bool)

### `transactions`
- `account_id` (relation to accounts)
- `transaction_type` (text: debit, credit)
- `amount` (number)
- `currency` (text)
- `description` (text)
- `category` (text)
- `merchant` (text)
- `reference` (text)
- `transaction_date` (date)
- `is_pending` (bool)

## Business Rules

1. **Balance Updates**: Balances are updated automatically when non-pending transactions are created
2. **Overdraft Protection**: Transactions that would result in negative balance are rejected
3. **Inactive Accounts**: Transactions cannot be created for inactive accounts
4. **Pending Transactions**: Pending transactions don't affect balance until settled
5. **Transaction Types**:
   - `debit`: Decreases account balance (withdrawals, purchases)
   - `credit`: Increases account balance (deposits, refunds)

## Integrating with PocketBase Hooks

To expose banking operations via API endpoints, add hooks in your PocketBase app:

```go
// In pkg/cmd/pocketbase/pb_hooks/banking.go
package pb_hooks

import (
    "github.com/joeblew999/wellknown/pkg/banking"
    "github.com/labstack/echo/v5"
    "github.com/pocketbase/pocketbase/core"
)

func RegisterBankingHooks(app core.App) {
    bankingService := banking.NewService(app)

    // GET /api/banking/accounts/:userId
    app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
        e.Router.GET("/api/banking/accounts/:userId", func(c echo.Context) error {
            userID := c.PathParam("userId")
            accounts, err := bankingService.Accounts.GetUserAccounts(userID)
            if err != nil {
                return c.JSON(500, map[string]string{"error": err.Error()})
            }
            return c.JSON(200, accounts)
        })
        return nil
    })
}
```

## Testing

The banking services can be tested independently:

```go
func TestCreateAccount(t *testing.T) {
    app := setupTestApp(t)
    service := banking.NewAccountService(app)

    account, err := service.CreateAccount(
        "user123", "ACC001", "Test Account", "checking", "USD", 100.00,
    )

    assert.NoError(t, err)
    assert.Equal(t, 100.00, account.GetFloat("balance"))
}
```

## Future Enhancements

- Add database transaction support (atomic operations)
- Implement transfer operations between accounts
- Add transaction reversal/refund functionality
- Implement account statements generation
- Add multi-currency support with exchange rates
- Implement overdraft/credit limit functionality
- Add audit logging for all operations
