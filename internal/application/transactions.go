package application

import (
	"context"
	ledger "ledger/internal/domain"
)

type AccountTransactionRepository interface {
	GetSystemAccount(ctx context.Context, kind, currency string) (ledger.Account, error)
	GetByIDForUpdate(ctx context.Context, id int64) (ledger.Account, error)
	SaveBalances(ctx context.Context, account ledger.Account) error
}

type TxRepositories struct {
	Accounts   AccountTransactionRepository
	Operations OperationRepository
	Journals   JournalRepository
	Holds      HoldRepository
}

type TransactionRunner interface {
	WithinTransaction(ctx context.Context, fn func(ctx TxRepositories) error) error
}
