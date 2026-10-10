package application

import (
	"context"

	ledger "ledger/internal/domain"
)

type JournalRepository interface {
	Create(ctx context.Context, journal ledger.JournalTransaction) error
}
