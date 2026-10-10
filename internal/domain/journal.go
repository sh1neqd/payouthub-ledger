package ledger

import (
	"errors"

	"github.com/google/uuid"
)

type JournalEntry struct {
	AccountID int64
	Delta     int64
}

type JournalTransaction struct {
	ID          uuid.UUID
	OperationID uuid.UUID
	Kind        string
	Currency    string
	HoldID      *uuid.UUID
	Entries     [2]JournalEntry
}

func (j JournalTransaction) Validate() error {
	if j.ID == uuid.Nil || j.OperationID == uuid.Nil {
		return errors.New("journal transaction must have an ID")
	}

	if j.Currency != "RUB" {
		return errors.New("journal currency must be RUB")
	}

	switch j.Kind {
	case "fund":
		if j.HoldID != nil {
			return errors.New("fund journal must not reference a hold")
		}
	case "capture":
		if j.HoldID == nil || *j.HoldID == uuid.Nil {
			return errors.New("capture journal requires a hold")
		}
	default:
		return errors.New("unsupported journal kind")
	}

	from, to := j.Entries[0], j.Entries[1]

	if from.AccountID <= 0 || to.AccountID <= 0 ||
		from.AccountID == to.AccountID {
		return errors.New("journal requires two distinct valid accounts")
	}

	if from.Delta >= 0 || to.Delta <= 0 {
		return errors.New(
			"journal requires a negative entry followed by a positive entry",
		)
	}

	if from.Delta != -to.Delta {
		return errors.New("journal entries must balance")
	}

	return nil

}
