package ledger

import "errors"

var (
	ErrInvalidAmount       = errors.New("amount must be positive")
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrAccountNotFound     = errors.New("account not found")
	ErrInvalidMerchantID   = errors.New("merchant_id must be a non-zero UUID")
	ErrUnsupportedCurrency = errors.New("currency must be RUB")
	ErrOperationNotFound   = errors.New("operation not found")
	ErrInvalidCursor       = errors.New("after_entry_id must be non-negative")
	ErrInvalidLimit        = errors.New("limit must be between 1 and 100")
	ErrHoldNotFound        = errors.New("hold not found")
)
