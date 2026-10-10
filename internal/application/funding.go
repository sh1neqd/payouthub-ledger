package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	ledger "ledger/internal/domain"

	"github.com/google/uuid"
)

type FundingService struct {
	tx TransactionRunner
}

func NewFundingService(tx TransactionRunner) *FundingService {
	return &FundingService{
		tx: tx,
	}
}

func (s *FundingService) FundAccount(ctx context.Context, input FundAccountInput) (FundAccountResult, error) {
	if input.OperationID == uuid.Nil {
		return FundAccountResult{}, ledger.ErrInvalidOperationID
	}

	if input.AccountID <= 0 {
		return FundAccountResult{}, ledger.ErrInvalidAccountID
	}

	if input.Amount <= 0 {
		return FundAccountResult{}, ledger.ErrInvalidAmount
	}

	if input.Currency != "RUB" {
		return FundAccountResult{}, ledger.ErrUnsupportedCurrency
	}

	request, err := json.Marshal(fundRequestV1{
		Version:   1,
		Kind:      "fund",
		AccountID: input.AccountID,
		Amount:    input.Amount,
		Currency:  input.Currency,
	})

	if err != nil {
		return FundAccountResult{}, fmt.Errorf("failed to marshal fund request: %w", err)
	}

	hash := sha256.Sum256(request)

	var result FundAccountResult
	var rejected *ledger.OperationRejectedError

	err = s.tx.WithinTransaction(ctx, func(repos TxRepositories) error {
		created, err := repos.Operations.TryCreate(ctx, ledger.Operation{
			ID:          input.OperationID,
			Kind:        "fund",
			Request:     request,
			RequestHash: hash[:],
		})

		if err != nil {
			return err
		}

		if !created {
			operation, err := repos.Operations.GetByID(ctx, input.OperationID)
			if err != nil {
				return err
			}

			if operation.Kind != "fund" || !bytes.Equal(operation.RequestHash, hash[:]) {
				return ledger.ErrIdempotencyConflict
			}

			switch operation.Status {
			case "succeeded":
				if err := json.Unmarshal(operation.Result, &result); err != nil {
					return fmt.Errorf("decode fund result: %w", err)
				}

				if result.Version != 1 {
					return fmt.Errorf("unsupported operation version %d", result.Version)
				}
				return nil

			case "rejected":
				if operation.ErrorCode == nil || *operation.ErrorCode == "" {
					return fmt.Errorf("rejected operation failed with no code")
				}

				rejected = &ledger.OperationRejectedError{
					Code: *operation.ErrorCode,
				}
				return nil

			default:
				return fmt.Errorf("unknown operation status %q", operation.Status)
			}
		}

		result, err = s.applyFund(ctx, repos, input)
		if err != nil {
			if errors.As(err, &rejected) {
				return repos.Operations.MarkRejected(ctx, input.OperationID, rejected.Code)
			}
			return err
		}
		savedResult, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result to JSON: %w", err)
		}

		return repos.Operations.MarkSucceeded(ctx, input.OperationID, savedResult)
	})

	if err != nil {
		return FundAccountResult{}, fmt.Errorf("fund account: %w", err)
	}

	if rejected != nil {
		return FundAccountResult{}, rejected
	}

	return result, nil
}
