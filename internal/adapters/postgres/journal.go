package postgres

import (
	"context"
	"fmt"
	ledger "ledger/internal/domain"
)

type JournalRepository struct {
	db queryer
}

func NewJournalRepository(db queryer) *JournalRepository {
	return &JournalRepository{db: db}
}

func (r *JournalRepository) Create(ctx context.Context, journal ledger.JournalTransaction) error {
	if err := journal.Validate(); err != nil {
		return fmt.Errorf("validate journal: %w", err)
	}

	const insertTransaction = `INSERT INTO journal_transactions 
    									(id, operation_id, kind, currency, hold_id)
    									VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(ctx, insertTransaction, journal.ID, journal.OperationID, journal.Kind, journal.Currency, journal.HoldID)

	if err != nil {
		return fmt.Errorf("insert journal transaction: %w", err)
	}

	const insertEntries = `
		INSERT INTO journal_entries (transaction_id, account_id, delta)
		VALUES ($1, $2, $3), ($1, $4, $5)
	`

	_, err = r.db.Exec(
		ctx,
		insertEntries,
		journal.ID,
		journal.Entries[0].AccountID,
		journal.Entries[0].Delta,
		journal.Entries[1].AccountID,
		journal.Entries[1].Delta,
	)
	if err != nil {
		return fmt.Errorf("insert journal entries: %w", err)
	}

	return nil
}
