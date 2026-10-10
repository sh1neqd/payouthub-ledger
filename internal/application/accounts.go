package application

import (
	"context"
	ledger "ledger/internal/domain"

	"github.com/google/uuid"
)

type AccountRepository interface {
	Create(
		ctx context.Context,
		merchantID uuid.UUID,
		currency string,
	) (ledger.Account, error)

	GetByID(ctx context.Context, id int64) (ledger.Account, error)
}

type AccountService struct {
	repo AccountRepository
}

type CreateAccountInput struct {
	MerchantID uuid.UUID
	Currency   string
}

func NewAccountService(repo AccountRepository) *AccountService {
	return &AccountService{
		repo: repo,
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, input CreateAccountInput) (ledger.Account, error) {
	if input.MerchantID == uuid.Nil {
		return ledger.Account{}, ledger.ErrInvalidMerchantID
	}

	if input.Currency != "RUB" {
		return ledger.Account{}, ledger.ErrUnsupportedCurrency
	}

	return s.repo.Create(ctx, input.MerchantID, input.Currency)
}

func (s *AccountService) GetAccount(ctx context.Context, id int64) (ledger.Account, error) {
	return s.repo.GetByID(ctx, id)
}
