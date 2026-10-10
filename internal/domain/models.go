package ledger

import "github.com/google/uuid"

type Account struct {
	Id         int64
	Kind       string
	MerchantID *uuid.UUID
	Currency   string
	Balance    int64
	Reserved   int64
}
