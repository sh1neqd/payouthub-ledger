package application

import (
	"context"
	"time"

	ledger "ledger/internal/domain"

	"github.com/google/uuid"
)

type JournalEntryView struct {
	ID            int64
	TransactionID uuid.UUID
	OperationID   uuid.UUID
	AccountID     int64
	Kind          string
	Currency      string
	Delta         int64
	CreatedAt     time.Time
}

type ListEntriesInput struct {
	AccountID    int64
	AfterEntryID int64
	Limit        int
}

type JournalRepository interface {
	Create(ctx context.Context, journal ledger.JournalTransaction) error

	ListByAccount(
		ctx context.Context,
		accountID, afterEntryID int64,
		limit int,
	) ([]JournalEntryView, error)
}

type JournalService struct {
	journals JournalRepository
	accounts AccountRepository
}

func NewJournalService(
	journals JournalRepository,
	accounts AccountRepository,
) *JournalService {
	return &JournalService{
		journals: journals,
		accounts: accounts,
	}
}

func (s *JournalService) ListEntries(
	ctx context.Context,
	input ListEntriesInput,
) ([]JournalEntryView, error) {
	if input.AccountID <= 0 {
		return nil, ledger.ErrInvalidAccountID
	}
	if input.AfterEntryID < 0 {
		return nil, ledger.ErrInvalidCursor
	}
	if input.Limit < 1 || input.Limit > 100 {
		return nil, ledger.ErrInvalidLimit
	}

	if _, err := s.accounts.GetByID(ctx, input.AccountID); err != nil {
		return nil, err
	}

	return s.journals.ListByAccount(
		ctx,
		input.AccountID,
		input.AfterEntryID,
		input.Limit,
	)
}
