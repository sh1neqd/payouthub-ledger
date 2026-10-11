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

type releaseRequestV1 struct {
	Version int       `json:"version"`
	Kind    string    `json:"kind"`
	HoldID  uuid.UUID `json:"hold_id"`
}

func (s *ReleaseService) Release(ctx context.Context, input ReleaseInput) (ReleaseResult, error) {
	if input.OperationID == uuid.Nil {
		return ReleaseResult{}, ledger.ErrInvalidOperationID
	}
	if input.HoldID == uuid.Nil {
		return ReleaseResult{}, ledger.ErrInvalidHoldID
	}

	request, err := json.Marshal(releaseRequestV1{
		Version: 1,
		Kind:    "release",
		HoldID:  input.HoldID,
	})
	if err != nil {
		return ReleaseResult{}, fmt.Errorf("encode release request: %w", err)
	}

	hash := sha256.Sum256(request)

	var result ReleaseResult
	var rejected *ledger.OperationRejectedError

	err = s.tx.WithinTransaction(ctx, func(repos TxRepositories) error {
		created, err := repos.Operations.TryCreate(ctx, ledger.Operation{
			ID:          input.OperationID,
			Kind:        "release",
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

			if operation.Kind != "release" ||
				!bytes.Equal(operation.RequestHash, hash[:]) {
				return ledger.ErrIdempotencyConflict
			}

			switch operation.Status {
			case "succeeded":
				if err := json.Unmarshal(operation.Result, &result); err != nil {
					return fmt.Errorf("decode release result: %w", err)
				}

				if result.Version != 1 {
					return fmt.Errorf(
						"unsupported release result version: %d",
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

		result, err = s.applyRelease(ctx, repos, input)
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
			return fmt.Errorf("encode release result: %w", err)
		}

		return repos.Operations.MarkSucceeded(
			ctx,
			input.OperationID,
			savedResult,
		)
	})
	if err != nil {
		return ReleaseResult{}, fmt.Errorf("release: %w", err)
	}
	
	if rejected != nil {
		return ReleaseResult{}, rejected
	}

	return result, nil
}
