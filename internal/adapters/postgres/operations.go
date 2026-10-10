package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	ledger "ledger/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OperationRepository struct {
	db queryer
}

func NewOperationRepository(db queryer) *OperationRepository {
	return &OperationRepository{
		db: db,
	}
}

func (r *OperationRepository) TryCreate(ctx context.Context, operation ledger.Operation) (bool, error) {
	const query = `
		INSERT INTO operations (id, kind, request, request_hash)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING 
	`

	result, err := r.db.Exec(ctx, query, operation.ID, operation.Kind, operation.Request, operation.RequestHash)

	if err != nil {
		return false, fmt.Errorf("failed to insert operation: %w", err)
	}

	return result.RowsAffected() == 1, nil
}

func (r *OperationRepository) GetByID(ctx context.Context, id uuid.UUID) (ledger.Operation, error) {
	const query = `
	SELECT id, kind, request, request_hash, status, result, error_code, created_at, completed_at
	FROM operations
	WHERE id = $1`

	var operation ledger.Operation
	err := r.db.QueryRow(ctx, query, id).Scan(
		&operation.ID,
		&operation.Kind,
		&operation.Request,
		&operation.RequestHash,
		&operation.Status,
		&operation.Result,
		&operation.ErrorCode,
		&operation.CreatedAt,
		&operation.CompletedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return ledger.Operation{}, ledger.ErrOperationNotFound
	}

	if err != nil {
		return ledger.Operation{}, fmt.Errorf("get account: %w", err)

	}
	return operation, nil
}

func (r *OperationRepository) MarkSucceeded(ctx context.Context, id uuid.UUID, result json.RawMessage) error {
	const query = `
		UPDATE operations
		SET status = 'succeeded',
		    result = $2,
		    error_code = NULL,
		    completed_at = clock_timestamp()
		WHERE id = $1 AND status = 'processing'
	`

	tag, err := r.db.Exec(ctx, query, id, result)

	if err != nil {
		return fmt.Errorf("failed to mark operation as succeeded: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf(
			"operation %s is missing or already completed",
			id,
		)
	}

	return nil
}

func (r *OperationRepository) MarkRejected(ctx context.Context, id uuid.UUID, errorCode string) error {
	const query = `
		UPDATE operations
		SET status = 'rejected',
			result = NULL,
			error_code = $2,
			completed_at = clock_timestamp()
			WHERE id = $1 AND status = 'processing'
	`
	tag, err := r.db.Exec(ctx, query, id, errorCode)
	if err != nil {
		return fmt.Errorf("failed to mark operation as rejected: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf(
			"operation %s is missing or already completed",
			id,
		)
	}
	return nil
}
