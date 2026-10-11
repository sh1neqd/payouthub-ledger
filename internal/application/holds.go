package application

import (
	"context"
	ledger "ledger/internal/domain"

	"github.com/google/uuid"
)

type HoldRepository interface {
	TryCreate(ctx context.Context, hold ledger.Hold) (bool, error)
	GetByPayoutID(ctx context.Context, payoutID uuid.UUID) (ledger.Hold, error)

	GetByID(ctx context.Context, id uuid.UUID) (ledger.Hold, error)
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (ledger.Hold, error)

	MarkCaptured(ctx context.Context, holdID, operationID uuid.UUID) error
	MarkReleased(ctx context.Context, holdID, operationID uuid.UUID) error
}
