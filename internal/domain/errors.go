package ledger

import "errors"

var (
	ErrInvalidAmount       = errors.New("amount must be positive")
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrAccountNotFound     = errors.New("account not found")
	ErrInvalidMerchantID   = errors.New("merchant_id must be a non-zero UUID")
	ErrUnsupportedCurrency = errors.New("currency must be RUB")
	ErrOperationNotFound   = errors.New("operation not found")
)
