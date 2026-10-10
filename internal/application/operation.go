package application

import (
	"context"
	"encoding/json"
	ledger "ledger/internal/domain"

	"github.com/google/uuid"
)

type OperationRepository interface {
	TryCreate(ctx context.Context, operation ledger.Operation) (bool, error)
	GetByID(ctx context.Context, id uuid.UUID) (ledger.Operation, error)

	MarkSucceeded(ctx context.Context, id uuid.UUID, result json.RawMessage) error
	MarkRejected(ctx context.Context, id uuid.UUID, errorCode string) error
}

type OperationService struct {
	repo OperationRepository
}

func NewOperationService(repo OperationRepository) *OperationService {
	return &OperationService{repo: repo}
}

func (s *OperationService) GetOperation(
	ctx context.Context,
	id uuid.UUID,
) (ledger.Operation, error) {
	if id == uuid.Nil {
		return ledger.Operation{}, ledger.ErrInvalidOperationID
	}
	return s.repo.GetByID(ctx, id)
}
