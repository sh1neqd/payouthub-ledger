package postgres

import (
	"context"
	"fmt"
	"ledger/internal/application"
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

func (r *JournalRepository) ListByAccount(ctx context.Context, accountID, afterEntryID int64, limit int) ([]application.JournalEntryView, error) {
	const query = `
		SELECT e.id, e.transaction_id, j.operation_id,
		       e.account_id, j.kind, j.currency, e.delta, j.created_at
		FROM journal_entries e
		JOIN journal_transactions j ON j.id = e.transaction_id
		WHERE e.account_id = $1 AND e.id > $2
		ORDER BY e.id ASC
		LIMIT $3
	`

	rows, err := r.db.Query(
		ctx,
		query,
		accountID,
		afterEntryID,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list journal entries: %w", err)
	}
	defer rows.Close()
	
	entries := make([]application.JournalEntryView, 0)

	for rows.Next() {
		var entry application.JournalEntryView

		if err := rows.Scan(
			&entry.ID,
			&entry.TransactionID,
			&entry.OperationID,
			&entry.AccountID,
			&entry.Kind,
			&entry.Currency,
			&entry.Delta,
			&entry.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan journal entry: %w", err)
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read journal entries: %w", err)
	}

	return entries, nil
}
