package domain

import (
	"context"

	"github.com/google/uuid"
)

type AccountRepository interface {
	Create(ctx context.Context, account Account) error
	GetByID(ctx context.Context, id uuid.UUID) (*Account, error)
	GetJournalEntries(ctx context.Context, accountID uuid.UUID) ([]*JournalEntry, error) 
}