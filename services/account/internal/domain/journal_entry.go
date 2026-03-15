package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type EntryDirection string 
const (
	DirectionDebit EntryDirection = "debit"
	DirectionCredit EntryDirection = "credit"
)

type JournalEntry struct {
	ID uuid.UUID
	AccountID uuid.UUID
	TransactionID uuid.UUID
	Direction EntryDirection
	Amount decimal.Decimal
	CreatedAt time.Time
}