package postgres

import (
	"context"
	"fmt"

	"ledger/internal/application"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRunner struct {
	pool *pgxpool.Pool
}

func NewTransactionRunner(pool *pgxpool.Pool) *TransactionRunner {
	return &TransactionRunner{pool: pool}
}

func (r *TransactionRunner) WithinTransaction(
	ctx context.Context,
	fn func(application.TxRepositories) error,
) error {
	err := pgx.BeginTxFunc(
		ctx,
		r.pool,
		pgx.TxOptions{
			IsoLevel: pgx.ReadCommitted,
		},
		func(tx pgx.Tx) error {
			repos := application.TxRepositories{
				Accounts:   NewAccountRepository(tx),
				Operations: NewOperationRepository(tx),
				Journals:   NewJournalRepository(tx),
			}

			return fn(repos)
		},
	)
	if err != nil {
		return fmt.Errorf("postgres transaction: %w", err)
	}

	return nil
}

var _ application.TransactionRunner = (*TransactionRunner)(nil)
