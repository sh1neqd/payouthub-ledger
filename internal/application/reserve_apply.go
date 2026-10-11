package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	ledger "ledger/internal/domain"

	"github.com/google/uuid"
)

type ReserveService struct {
	tx TransactionRunner
}

func NewReserveService(tx TransactionRunner) *ReserveService {
	return &ReserveService{
		tx: tx,
	}
}

func (s *ReserveService) applyReserve(ctx context.Context, repos TxRepositories, input ReserveInput) (ReserveResult, error) {
	wallet, err := repos.Accounts.GetByIDForUpdate(ctx, input.AccountID)

	if errors.Is(err, ledger.ErrAccountNotFound) {
		return ReserveResult{}, &ledger.OperationRejectedError{
			Code: "ACCOUNT_NOT_FOUND",
		}
	}
	if err != nil {
		return ReserveResult{}, err
	}

	if wallet.Kind != "wallet" {
		return ReserveResult{}, &ledger.OperationRejectedError{
			Code: "INVALID_ACCOUNT_KIND",
		}
	}

	if wallet.Currency != input.Currency {
		return ReserveResult{}, &ledger.OperationRejectedError{
			Code: "CURRENCY_MISMATCH",
		}
	}

	existing, err := repos.Holds.GetByPayoutID(ctx, input.PayoutID)
	if err == nil {
		return s.resultFromExistingHold(ctx, repos, existing, input)
	}
	if !errors.Is(err, ledger.ErrHoldNotFound) {
		return ReserveResult{}, err
	}

	if err := wallet.Reserve(input.Amount); err != nil {
		if errors.Is(err, ledger.ErrInsufficientFunds) {
			return ReserveResult{}, &ledger.OperationRejectedError{
				Code: "INSUFFICIENT_FUNDS",
			}
		}
		return ReserveResult{}, err
	}

	hold := ledger.Hold{
		ID:                 uuid.New(),
		PayoutID:           input.PayoutID,
		AccountID:          wallet.Id,
		Amount:             input.Amount,
		Currency:           input.Currency,
		Status:             "active",
		ReserveOperationID: input.OperationID,
	}

	created, err := repos.Holds.TryCreate(ctx, hold)
	if err != nil {
		return ReserveResult{}, err
	}

	if !created {
		existing, err := repos.Holds.GetByPayoutID(ctx, input.PayoutID)
		if err != nil {
			return ReserveResult{}, err
		}

		return s.resultFromExistingHold(ctx, repos, existing, input)
	}

	if err := repos.Accounts.SaveBalances(ctx, wallet); err != nil {
		return ReserveResult{}, err
	}

	return ReserveResult{
		Version:              1,
		EffectiveOperationID: input.OperationID,
		PayoutID:             hold.PayoutID,
		HoldID:               hold.ID,
		HoldStatus:           hold.Status,
		AccountID:            wallet.Id,
		Amount:               hold.Amount,
		Currency:             hold.Currency,
		Balance:              wallet.Balance,
		Reserved:             wallet.Reserved,
		Available:            wallet.Available(),
	}, nil
}

func (s *ReserveService) resultFromExistingHold(ctx context.Context, repos TxRepositories, hold ledger.Hold, input ReserveInput) (ReserveResult, error) {
	if hold.AccountID != input.AccountID ||
		hold.Amount != input.Amount ||
		hold.Currency != input.Currency {
		return ReserveResult{}, &ledger.OperationRejectedError{
			Code: "PAYOUT_CONFLICT",
		}
	}

	operation, err := repos.Operations.GetByID(
		ctx,
		hold.ReserveOperationID,
	)
	if err != nil {
		return ReserveResult{}, err
	}

	if operation.Kind != "reserve" || operation.Status != "succeeded" {
		return ReserveResult{}, fmt.Errorf(
			"hold references an invalid reserve operation",
		)
	}

	var result ReserveResult

	if err := json.Unmarshal(operation.Result, &result); err != nil {
		return ReserveResult{}, fmt.Errorf(
			"decode original reserve result: %w",
			err,
		)
	}

	if result.Version != 1 ||
		result.HoldID != hold.ID ||
		result.EffectiveOperationID != hold.ReserveOperationID {
		return ReserveResult{}, fmt.Errorf("invalid stored reserve result")
	}

	return result, nil
}
