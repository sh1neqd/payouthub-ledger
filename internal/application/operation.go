package application

import (
	"context"
	"encoding/json"
	ledger "ledger/internal/domain"

	"github.com/google/uuid"
)

type OperationRepository interface {
	TryCreate(ctx context.Context, account ledger.Operation) (bool, error)
	GetByID(ctx context.Context, id uuid.UUID) (ledger.Operation, error)

	MarkSucceeded(ctx context.Context, id uuid.UUID, result json.RawMessage) error
	MarkRejected(ctx context.Context, id uuid.UUID, errorCode string) error
}
