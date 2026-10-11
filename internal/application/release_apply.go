package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	ledger "ledger/internal/domain"
)

type ReleaseService struct {
	tx TransactionRunner
}

func NewReleaseService(tx TransactionRunner) *ReleaseService {
	return &ReleaseService{tx: tx}
}

func (s *ReleaseService) applyRelease(ctx context.Context, repos TxRepositories, input ReleaseInput) (ReleaseResult, error) {
	initialHold, err := repos.Holds.GetByID(ctx, input.HoldID)
	if errors.Is(err, ledger.ErrHoldNotFound) {
		return ReleaseResult{}, &ledger.OperationRejectedError{
			Code: "HOLD_NOT_FOUND",
		}
	}
	if err != nil {
		return ReleaseResult{}, err
	}

	wallet, err := repos.Accounts.GetByIDForUpdate(ctx, initialHold.AccountID)
	if err != nil {
		return ReleaseResult{}, err
	}

	hold, err := repos.Holds.GetByIDForUpdate(ctx, input.HoldID)
	if err != nil {
		return ReleaseResult{}, err
	}

	if hold.AccountID != wallet.Id ||
		wallet.Kind != "wallet" ||
		hold.Currency != wallet.Currency {
		return ReleaseResult{}, fmt.Errorf(
			"hold %s does not match its wallet",
			hold.ID,
		)
	}

	switch hold.Status {
	case "released":
		return s.resultFromReleasedHold(ctx, repos, hold)

	case "captured":
		return ReleaseResult{}, &ledger.OperationRejectedError{
			Code: "HOLD_STATE_CONFLICT",
		}

	case "active":

	default:
		return ReleaseResult{}, fmt.Errorf(
			"unexpected hold status: %s",
			hold.Status,
		)
	}

	if wallet.Reserved < hold.Amount {
		return ReleaseResult{}, fmt.Errorf(
			"wallet %d reserved amount is less than hold %s amount",
			wallet.Id,
			hold.ID,
		)
	}

	wallet.Reserved -= hold.Amount

	if err := repos.Accounts.SaveBalances(ctx, wallet); err != nil {
		return ReleaseResult{}, err
	}

	if err := repos.Holds.MarkReleased(
		ctx,
		hold.ID,
		input.OperationID,
	); err != nil {
		return ReleaseResult{}, err
	}

	return ReleaseResult{
		Version:              1,
		EffectiveOperationID: input.OperationID,
		HoldID:               hold.ID,
		HoldStatus:           "released",
		AccountID:            wallet.Id,
		Amount:               hold.Amount,
		Currency:             hold.Currency,
		Balance:              wallet.Balance,
		Reserved:             wallet.Reserved,
		Available:            wallet.Available(),
	}, nil
}

func (s *ReleaseService) resultFromReleasedHold(ctx context.Context, repos TxRepositories, hold ledger.Hold) (ReleaseResult, error) {
	if hold.ReleaseOperationID == nil {
		return ReleaseResult{}, fmt.Errorf(
			"released hold %s has no release operation",
			hold.ID,
		)
	}

	operation, err := repos.Operations.GetByID(ctx, *hold.ReleaseOperationID)
	if err != nil {
		return ReleaseResult{}, err
	}

	if operation.Kind != "release" || operation.Status != "succeeded" {
		return ReleaseResult{}, fmt.Errorf(
			"hold %s references an invalid release operation",
			hold.ID,
		)
	}

	var result ReleaseResult

	if err := json.Unmarshal(operation.Result, &result); err != nil {
		return ReleaseResult{}, fmt.Errorf(
			"decode original release result: %w",
			err,
		)
	}

	if result.Version != 1 ||
		result.HoldID != hold.ID ||
		result.HoldStatus != "released" ||
		result.EffectiveOperationID != *hold.ReleaseOperationID {
		return ReleaseResult{}, fmt.Errorf("invalid stored release result")
	}

	return result, nil
}
