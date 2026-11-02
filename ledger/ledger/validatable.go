package ledger

import "errors"

// Validatable describes types that can validate their own fields.
type Validatable interface {
	Validate() error
}

// CheckValid runs validation for any Validatable value (polymorphic helper).
func CheckValid(v Validatable) error {
	return v.Validate()
}

// Validate checks transaction fields before persistence.
func (t Transaction) Validate() error {
	if t.Amount <= 0 {
		return errors.New("transaction amount must be positive")
	}
	if t.Category == "" {
		return errors.New("transaction category must not be empty")
	}
	return nil
}

// Validate checks budget fields before persistence.
func (b Budget) Validate() error {
	if b.Category == "" {
		return errors.New("budget category must not be empty")
	}
	if b.Limit <= 0 {
		return errors.New("budget limit must be positive")
	}
	return nil
}
