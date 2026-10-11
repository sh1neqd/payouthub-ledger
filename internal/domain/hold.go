package ledger

import (
	"time"

	"github.com/google/uuid"
)

type Hold struct {
	ID                 uuid.UUID
	PayoutID           uuid.UUID
	AccountID          int64
	Amount             int64
	Currency           string
	Status             string
	ReserveOperationID uuid.UUID
	CaptureOperationID *uuid.UUID
	ReleaseOperationID *uuid.UUID
	CreatedAt          time.Time
	CompletedAt        *time.Time
}
