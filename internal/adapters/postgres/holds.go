package postgres

import (
	"context"
	"errors"
	"fmt"
	ledger "ledger/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type HoldRepository struct {
	db queryer
}

func NewHoldRepository(db queryer) *HoldRepository {
	return &HoldRepository{
		db: db,
	}
}

const holdColumns = `
	id, payout_id, account_id, amount, currency, status,
	reserve_operation_id, capture_operation_id, release_operation_id,
	created_at, completed_at
`

func (r HoldRepository) TryCreate(ctx context.Context, hold ledger.Hold) (bool, error) {
	const query = `
				INSERT INTO holds (id, payout_id, account_id, amount, currency, reserve_operation_id)
				VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT (payout_id) DO NOTHING
			`

	tag, err := r.db.Exec(ctx, query, hold.ID, hold.PayoutID, hold.AccountID, hold.Amount, hold.Currency, hold.ReserveOperationID)

	if err != nil {
		return false, fmt.Errorf("failed to insert hold: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}

func scanHold(row pgx.Row) (ledger.Hold, error) {
	var hold ledger.Hold

	err := row.Scan(
		&hold.ID,
		&hold.PayoutID,
		&hold.AccountID,
		&hold.Amount,
		&hold.Currency,
		&hold.Status,
		&hold.ReserveOperationID,
		&hold.CaptureOperationID,
		&hold.ReleaseOperationID,
		&hold.CreatedAt,
		&hold.CompletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ledger.Hold{}, ledger.ErrHoldNotFound
	}
	if err != nil {
		return ledger.Hold{}, fmt.Errorf("scan hold: %w", err)
	}

	return hold, nil
}

func (r HoldRepository) GetByPayoutID(ctx context.Context, payoutID uuid.UUID) (ledger.Hold, error) {
	const query = "SELECT " + holdColumns + " FROM holds WHERE payout_id = $1"

	return scanHold(r.db.QueryRow(ctx, query, payoutID))
}

func (r HoldRepository) GetByID(ctx context.Context, id uuid.UUID) (ledger.Hold, error) {
	const query = "SELECT " + holdColumns + " FROM holds WHERE id = $1"

	return scanHold(r.db.QueryRow(ctx, query, id))
}

func (r HoldRepository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (ledger.Hold, error) {
	const query = "SELECT " + holdColumns + " FROM holds WHERE id = $1 FOR UPDATE"

	return scanHold(r.db.QueryRow(ctx, query, id))
}

func (r HoldRepository) MarkCaptured(ctx context.Context, holdID, operationID uuid.UUID) error {
	const query = `
		UPDATE holds
		SET status = 'captured',
		    capture_operation_id = $2,
		    completed_at = clock_timestamp()
		WHERE id = $1 AND status = 'active'
	`

	tag, err := r.db.Exec(ctx, query, holdID, operationID)
	if err != nil {
		return fmt.Errorf("mark hold captured: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf(
			"mark hold captured: active hold %s not found",
			holdID,
		)
	}

	return nil
}

func (r HoldRepository) MarkReleased(ctx context.Context, holdID, operationID uuid.UUID) error {
	const query = `
		UPDATE holds
		SET status = 'released',
		    release_operation_id = $2,
		    completed_at = clock_timestamp()
		WHERE id = $1 AND status = 'active'
	`

	tag, err := r.db.Exec(ctx, query, holdID, operationID)
	if err != nil {
		return fmt.Errorf("mark hold released: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf(
			"mark hold released: active hold %s not found",
			holdID,
		)
	}

	return nil
}
