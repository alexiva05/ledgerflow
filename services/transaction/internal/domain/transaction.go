package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCanclled  TransactionStatus = "cancelled"
)

type Transaction struct {
	ID            uuid.UUID         `json:"id"`
	FromAccountID uuid.UUID         `json:"from_account_id"`
	ToAccountID   uuid.UUID         `json:"to_account_id"`
	Amount        decimal.Decimal   `json:"amount"`
	Currency      string            `json:"currency"`
	Status        TransactionStatus `json:"status"`
	CreatedAt     time.Time         `json:"created_at"`
	CompletedAt   *time.Time        `json:"completed_at"`
}
