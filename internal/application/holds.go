package application

import (
	"context"
	ledger "ledger/internal/domain"

	"github.com/google/uuid"
)

type HoldRepository interface {
	TryCreate(ctx context.Context, hold ledger.Hold) (bool, error)

	GetByPayoutID(ctx context.Context, payoutID uuid.UUID) (ledger.Hold, error)
}
