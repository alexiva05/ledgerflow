package domain

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionStatus string

const (
	TransactionCompleted TransactionStatus = "transaction.completed"
	TransactionFailed    TransactionStatus = "transaction.failed"
)

type TransactionEvent struct {
	TransactionID uuid.UUID         `json:"transaction_id"`
	AccountID     uuid.UUID         `json:"account_id"`
	Amount        decimal.Decimal   `json:"amount"`
	Status        TransactionStatus `json:"status"`
}
