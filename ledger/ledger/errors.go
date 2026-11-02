package ledger

import "errors"

// ErrBudgetExceeded is returned when a transaction would exceed the category budget.
var ErrBudgetExceeded = errors.New("budget exceeded")
