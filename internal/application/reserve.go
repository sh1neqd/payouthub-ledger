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

type reserveRequestV1 struct {
	Version   int       `json:"version"`
	Kind      string    `json:"kind"`
	PayoutID  uuid.UUID `json:"payout_id"`
	AccountID int64     `json:"account_id"`
	Amount    int64     `json:"amount"`
	Currency  string    `json:"currency"`
}

func (s *ReserveService) Reserve(ctx context.Context, input ReserveInput) (ReserveResult, error) {
	if input.OperationID == uuid.Nil {
		return ReserveResult{}, ledger.ErrInvalidOperationID
	}
	if input.PayoutID == uuid.Nil {
		return ReserveResult{}, ledger.ErrInvalidPayoutID
	}
	if input.AccountID <= 0 {
		return ReserveResult{}, ledger.ErrInvalidAccountID
	}
	if input.Amount <= 0 {
		return ReserveResult{}, ledger.ErrInvalidAmount
	}
	if input.Currency != "RUB" {
		return ReserveResult{}, ledger.ErrUnsupportedCurrency
	}

	request, err := json.Marshal(reserveRequestV1{
		Version:   1,
		Kind:      "reserve",
		PayoutID:  input.PayoutID,
		AccountID: input.AccountID,
		Amount:    input.Amount,
		Currency:  input.Currency,
	})
	if err != nil {
		return ReserveResult{}, fmt.Errorf("encode reserve request: %w", err)
	}

	hash := sha256.Sum256(request)

	var result ReserveResult
	var rejected *ledger.OperationRejectedError

	err = s.tx.WithinTransaction(ctx, func(repos TxRepositories) error {
		created, err := repos.Operations.TryCreate(ctx, ledger.Operation{
			ID:          input.OperationID,
			Kind:        "reserve",
			Request:     request,
			RequestHash: hash[:],
		})
		if err != nil {
			return err
		}

		if !created {
			operation, err := repos.Operations.GetByID(
				ctx,
				input.OperationID,
			)
			if err != nil {
				return err
			}

			if operation.Kind != "reserve" ||
				!bytes.Equal(operation.RequestHash, hash[:]) {
				return ledger.ErrIdempotencyConflict
			}

			switch operation.Status {
			case "succeeded":
				if err := json.Unmarshal(operation.Result, &result); err != nil {
					return fmt.Errorf("decode reserve result: %w", err)
				}

				if result.Version != 1 {
					return fmt.Errorf(
						"unsupported reserve result version: %d",
						result.Version,
					)
				}

				return nil

			case "rejected":
				if operation.ErrorCode == nil || *operation.ErrorCode == "" {
					return fmt.Errorf("rejected operation has no error code")
				}

				rejected = &ledger.OperationRejectedError{
					Code: *operation.ErrorCode,
				}

				return nil

			default:
				return fmt.Errorf(
					"unexpected stored operation status: %s",
					operation.Status,
				)
			}
		}

		result, err = s.applyReserve(ctx, repos, input)
		if err != nil {
			if errors.As(err, &rejected) {
				return repos.Operations.MarkRejected(
					ctx,
					input.OperationID,
					rejected.Code,
				)
			}

			return err
		}

		savedResult, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("encode reserve result: %w", err)
		}

		return repos.Operations.MarkSucceeded(
			ctx,
			input.OperationID,
			savedResult,
		)
	})
	if err != nil {
		return ReserveResult{}, fmt.Errorf("reserve: %w", err)
	}

	if rejected != nil {
		return ReserveResult{}, rejected
	}

	return result, nil
}
