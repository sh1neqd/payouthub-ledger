package ledger

import "errors"

var (
	ErrInvalidOperationID = errors.New(
		"operation_id must be a non-zero UUID",
	)
	ErrInvalidAccountID = errors.New(
		"account_id must be positive",
	)
	ErrIdempotencyConflict = errors.New(
		"operation_id was used with different parameters",
	)
	ErrInvalidPayoutID = errors.New(
		"payout_id must be a non-zero UUID",
	)
	ErrInvalidHoldID = errors.New(
		"hold_id must be a non-zero UUID",
	)
)

type OperationRejectedError struct {
	Code string
}

func (e *OperationRejectedError) Error() string {
	return e.Code
}
