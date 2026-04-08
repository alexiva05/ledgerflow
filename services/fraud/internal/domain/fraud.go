package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionEvent struct {
	TransactionID uuid.UUID       `json:"transaction_id"`
	AccountID     uuid.UUID       `json:"account_id"`
	Amount        decimal.Decimal `json:"amount"`
	Timestamp     time.Time       `json:"timestamp"`
}
