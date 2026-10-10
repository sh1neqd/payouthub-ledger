package application

import (
	"context"
	"errors"
	"fmt"
	ledger "ledger/internal/domain"
	"math"

	"github.com/google/uuid"
)

func (s *FundingService) applyFund(ctx context.Context, repos TxRepositories, input FundAccountInput) (FundAccountResult, error) {
	funding, err := repos.Accounts.GetSystemAccount(ctx, "funding", input.Currency)

	if err != nil {
		return FundAccountResult{}, err
	}

	if funding.Id == input.AccountID {
		return FundAccountResult{}, &ledger.OperationRejectedError{
			Code: "INVALID_ACCOUNT_KIND",
		}
	}

	ids := [2]int64{funding.Id, input.AccountID}

	if ids[0] > ids[1] {
		ids[0], ids[1] = ids[1], ids[0]
	}

	var wallet ledger.Account

	for _, id := range ids {
		account, err := repos.Accounts.GetByIDForUpdate(ctx, id)
		if err != nil {
			if id == input.AccountID && errors.Is(err, ledger.ErrAccountNotFound) {
				return FundAccountResult{}, &ledger.OperationRejectedError{
					Code: "ACCOUNT_NOT_FOUND",
				}
			}
			return FundAccountResult{}, err
		}

		if id == funding.Id {
			funding = account
		} else {
			wallet = account
		}
	}

	if funding.Kind != "funding" || funding.Currency != input.Currency {
		return FundAccountResult{}, fmt.Errorf("invalid funding account configuration")
	}

	if wallet.Kind != "wallet" {
		return FundAccountResult{}, &ledger.OperationRejectedError{
			Code: "INVALID_ACCOUNT_KIND",
		}
	}

	if wallet.Currency != input.Currency {
		return FundAccountResult{}, &ledger.OperationRejectedError{
			Code: "CURRENCY_MISMATCH",
		}
	}

	if wallet.Balance > math.MaxInt64-input.Amount ||
		funding.Balance < math.MinInt64+input.Amount {
		return FundAccountResult{}, &ledger.OperationRejectedError{
			Code: "AMOUNT_OVERFLOW",
		}
	}

	funding.Balance -= input.Amount
	wallet.Balance += input.Amount

	if err := repos.Accounts.SaveBalances(ctx, funding); err != nil {
		return FundAccountResult{}, err
	}

	if err := repos.Accounts.SaveBalances(ctx, wallet); err != nil {
		return FundAccountResult{}, err
	}

	journal := ledger.JournalTransaction{
		ID:          uuid.New(),
		OperationID: input.OperationID,
		Kind:        "fund",
		Currency:    input.Currency,
		Entries: [2]ledger.JournalEntry{
			{AccountID: funding.Id, Delta: -input.Amount},
			{AccountID: wallet.Id, Delta: input.Amount},
		},
	}

	if err := repos.Journals.Create(ctx, journal); err != nil {
		return FundAccountResult{}, err
	}

	return FundAccountResult{
		Version:              1,
		EffectiveOperationID: input.OperationID,
		AccountID:            wallet.Id,
		Amount:               input.Amount,
		Currency:             wallet.Currency,
		Balance:              wallet.Balance,
		Reserved:             wallet.Reserved,
		Available:            wallet.Available(),
		JournalID:            journal.ID,
	}, nil
}
