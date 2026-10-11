package application

import "github.com/google/uuid"

type CaptureInput struct {
	OperationID uuid.UUID
	HoldID      uuid.UUID
}

type CaptureResult struct {
	Version              int       `json:"version"`
	EffectiveOperationID uuid.UUID `json:"effective_operation_id"`
	TransactionID        uuid.UUID `json:"transaction_id"`
	HoldID               uuid.UUID `json:"hold_id"`
	HoldStatus           string    `json:"hold_status"`
	AccountID            int64     `json:"account_id"`
	Amount               int64     `json:"amount"`
	Currency             string    `json:"currency"`
	Balance              int64     `json:"balance"`
	Reserved             int64     `json:"reserved"`
	Available            int64     `json:"available"`
}
