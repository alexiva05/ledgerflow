package domain

import (
	"context"

	"github.com/google/uuid"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction *Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	UpdateStatus(ctx context.Context, transactionID uuid.UUID, status TransactionStatus) error
}