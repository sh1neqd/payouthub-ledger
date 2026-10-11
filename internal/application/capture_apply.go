package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	ledger "ledger/internal/domain"
	"math"

	"github.com/google/uuid"
)

type CaptureService struct {
	tx TransactionRunner
}

func NewCaptureService(tx TransactionRunner) *CaptureService {
	return &CaptureService{tx: tx}
}

func (s *CaptureService) applyCapture(ctx context.Context, repos TxRepositories, input CaptureInput) (CaptureResult, error) {
	initialHold, err := repos.Holds.GetByID(ctx, input.HoldID)
	if errors.Is(err, ledger.ErrHoldNotFound) {
		return CaptureResult{}, &ledger.OperationRejectedError{
			Code: "HOLD_NOT_FOUND",
		}
	}
	if err != nil {
		return CaptureResult{}, err
	}

	settlementAccount, err := repos.Accounts.GetSystemAccount(ctx, "settlement", initialHold.Currency)
	if err != nil {
		return CaptureResult{}, err
	}

	if initialHold.AccountID == settlementAccount.Id {
		return CaptureResult{}, fmt.Errorf("hold account must differ from settlement account")
	}

	firstID, secondID := initialHold.AccountID, settlementAccount.Id
	if firstID > secondID {
		firstID, secondID = secondID, firstID
	}

	first, err := repos.Accounts.GetByIDForUpdate(ctx, firstID)
	if err != nil {
		return CaptureResult{}, err
	}

	second, err := repos.Accounts.GetByIDForUpdate(ctx, secondID)
	if err != nil {
		return CaptureResult{}, err
	}

	wallet, settlement := first, second
	if wallet.Id != initialHold.AccountID {
		wallet, settlement = second, first
	}

	hold, err := repos.Holds.GetByIDForUpdate(ctx, input.HoldID)
	if err != nil {
		return CaptureResult{}, err
	}

	if hold.AccountID != wallet.Id ||
		wallet.Kind != "wallet" ||
		settlement.Kind != "settlement" ||
		wallet.Currency != hold.Currency ||
		settlement.Currency != hold.Currency {
		return CaptureResult{}, fmt.Errorf("hold %s does not match capture accounts", hold.ID)
	}

	switch hold.Status {
	case "captured":
		return s.resultFromCapturedHold(ctx, repos, hold)

	case "released":
		return CaptureResult{}, &ledger.OperationRejectedError{
			Code: "HOLD_STATE_CONFLICT",
		}

	case "active":

	default:
		return CaptureResult{}, fmt.Errorf("unexpected hold status: %s", hold.Status)
	}

	if wallet.Reserved < hold.Amount || wallet.Balance < hold.Amount {
		return CaptureResult{}, fmt.Errorf("wallet %d does not cover active hold %s", wallet.Id, hold.ID)
	}

	if settlement.Balance > math.MaxInt64-hold.Amount {
		return CaptureResult{}, &ledger.OperationRejectedError{
			Code: "AMOUNT_OVERFLOW",
		}
	}

	wallet.Balance -= hold.Amount
	wallet.Reserved -= hold.Amount
	settlement.Balance += hold.Amount

	if err := repos.Accounts.SaveBalances(ctx, wallet); err != nil {
		return CaptureResult{}, err
	}

	if err := repos.Accounts.SaveBalances(ctx, settlement); err != nil {
		return CaptureResult{}, err
	}

	journal := ledger.JournalTransaction{
		ID:          uuid.New(),
		OperationID: input.OperationID,
		Kind:        "capture",
		Currency:    hold.Currency,
		HoldID:      &hold.ID,
		Entries: [2]ledger.JournalEntry{
			{
				AccountID: wallet.Id,
				Delta:     -hold.Amount,
			},
			{
				AccountID: settlement.Id,
				Delta:     hold.Amount,
			},
		},
	}

	if err := repos.Journals.Create(ctx, journal); err != nil {
		return CaptureResult{}, err
	}

	if err := repos.Holds.MarkCaptured(ctx, hold.ID, input.OperationID); err != nil {
		return CaptureResult{}, err
	}

	return CaptureResult{
		Version:              1,
		EffectiveOperationID: input.OperationID,
		TransactionID:        journal.ID,
		HoldID:               hold.ID,
		HoldStatus:           "captured",
		AccountID:            wallet.Id,
		Amount:               hold.Amount,
		Currency:             hold.Currency,
		Balance:              wallet.Balance,
		Reserved:             wallet.Reserved,
		Available:            wallet.Available(),
	}, nil
}

func (s *CaptureService) resultFromCapturedHold(ctx context.Context, repos TxRepositories, hold ledger.Hold) (CaptureResult, error) {
	if hold.CaptureOperationID == nil {
		return CaptureResult{}, fmt.Errorf("captured hold %s has no capture operation", hold.ID)
	}

	operation, err := repos.Operations.GetByID(ctx, *hold.CaptureOperationID)
	if err != nil {
		return CaptureResult{}, err
	}

	if operation.Kind != "capture" || operation.Status != "succeeded" {
		return CaptureResult{}, fmt.Errorf("hold %s references an invalid capture operation", hold.ID)
	}

	var result CaptureResult

	if err := json.Unmarshal(operation.Result, &result); err != nil {
		return CaptureResult{}, fmt.Errorf("decode original capture result: %w", err)
	}

	if result.Version != 1 ||
		result.HoldID != hold.ID ||
		result.HoldStatus != "captured" ||
		result.TransactionID == uuid.Nil ||
		result.EffectiveOperationID != *hold.CaptureOperationID {
		return CaptureResult{}, fmt.Errorf("invalid stored capture result")
	}

	return result, nil
}
