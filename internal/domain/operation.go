package ledger

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Operation struct {
	ID          uuid.UUID
	Kind        string
	Request     json.RawMessage
	RequestHash []byte
	Status      string
	Result      json.RawMessage
	ErrorCode   *string
	CreatedAt   time.Time
	CompletedAt *time.Time
}
