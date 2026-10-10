package application

import "github.com/google/uuid"

type FundAccountInput struct {
	OperationID uuid.UUID
	AccountID   int64
	Amount      int64
	Currency    string
}

type FundAccountResult struct {
	Version              int       `json:"version"`
	EffectiveOperationID uuid.UUID `json:"effective_operation_id"`
	AccountID            int64     `json:"account_id"`
	Amount               int64     `json:"amount"`
	Currency             string    `json:"currency"`
	Balance              int64     `json:"balance"`
	Reserved             int64     `json:"reserved"`
	Available            int64     `json:"available"`
	JournalID            uuid.UUID `json:"journal_id"`
}

type fundRequestV1 struct {
	Version   int    `json:"version"`
	Kind      string `json:"kind"`
	AccountID int64  `json:"account_id"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
}
