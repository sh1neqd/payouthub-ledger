package postgres

import (
	"context"
	"errors"
	"fmt"
	ledger "ledger/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type queryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}
type AccountRepository struct {
	db queryer
}

func NewAccountRepository(db queryer) *AccountRepository {
	return &AccountRepository{
		db: db,
	}
}

func (r *AccountRepository) Create(ctx context.Context, merchantID uuid.UUID, currency string) (ledger.Account, error) {
	const query = `
		INSERT INTO accounts (kind, merchant_id, currency)
		VALUES ('wallet', $1, $2)
		ON CONFLICT (merchant_id, currency) DO NOTHING
		RETURNING id, kind, merchant_id, currency, balance, reserved
	`

	account, err := scanAccount(r.db.QueryRow(ctx, query, merchantID.String(), currency))

	if errors.Is(err, pgx.ErrNoRows) {
		const selectQuery = `
			SELECT id, kind, merchant_id, currency, balance, reserved
			FROM accounts
			WHERE merchant_id = $1 AND currency = $2
		`

		account, err = scanAccount(r.db.QueryRow(ctx, selectQuery, merchantID.String(), currency))
	}

	if err != nil {
		return ledger.Account{}, fmt.Errorf("create account: %w", err)
	}

	return account, nil

}

func (r *AccountRepository) GetByID(ctx context.Context, id int64) (ledger.Account, error) {
	const query = `
	SELECT id, kind, merchant_id, currency, balance, reserved
	FROM accounts
	WHERE id = $1`

	a, err := scanAccount(r.db.QueryRow(ctx, query, id))

	if errors.Is(err, pgx.ErrNoRows) {
		return ledger.Account{}, ledger.ErrAccountNotFound
	}

	if err != nil {
		return ledger.Account{}, fmt.Errorf("get account: %w", err)

	}
	return a, nil
}

func scanAccount(row pgx.Row) (ledger.Account, error) {
	var account ledger.Account

	err := row.Scan(
		&account.Id,
		&account.Kind,
		&account.MerchantID,
		&account.Currency,
		&account.Balance,
		&account.Reserved,
	)
	if err != nil {
		return ledger.Account{}, err
	}

	return account, nil
}
