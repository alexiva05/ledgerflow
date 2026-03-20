package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionStatus string
const (
	TransactionStatusPending TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed TransactionStatus = "failed"
	TransactionStatusCanclled TransactionStatus = "cancelled"
)

type Transaction struct {
	ID uuid.UUID
	FromAccountID uuid.UUID
	ToAccountID uuid.UUID
	Amount decimal.Decimal
	Currency string
	Status TransactionStatus
	CreatedAt time.Time
	CompletedAt *time.Time
}