package postgres

import (
	"context"
	"errors"
	"fmt"
	ledger "ledger/internal/domain"

	"github.com/jackc/pgx/v5"
)

func (r *AccountRepository) GetSystemAccount(ctx context.Context, kind, currency string) (ledger.Account, error) {
	const query = `
		SELECT id, kind, merchant_id, currency, balance, reserved
		FROM accounts
		WHERE kind = $1 AND currency = $2
			AND kind IN ('funding', 'settlement')
	`

	account, err := scanAccount(r.db.QueryRow(ctx, query, kind, currency))
	if err != nil {
		return ledger.Account{}, fmt.Errorf(
			"get system account %s/%s: %w",
			kind, currency, err,
		)
	}

	return account, nil
}

func (r *AccountRepository) GetByIDForUpdate(ctx context.Context, id int64) (ledger.Account, error) {
	const query = `
		SELECT id, kind, merchant_id, currency, balance, reserved
		FROM accounts
		WHERE id = $1
		FOR UPDATE
	`

	account, err := scanAccount(r.db.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return ledger.Account{}, ledger.ErrAccountNotFound
	}

	if err != nil {
		return ledger.Account{}, fmt.Errorf("lock account: %w", err)
	}

	return account, nil
}

func (r *AccountRepository) SaveBalances(ctx context.Context, account ledger.Account) error {
	const query = `
		UPDATE accounts
		SET balance = $2, reserved = $3
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, account.Id, account.Balance, account.Reserved)

	if err != nil {
		return fmt.Errorf("save balance: %w", err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("save balance: %w", ledger.ErrAccountNotFound)
	}

	return nil
}
