package app

import (
	"context"
	"fmt"
	"ledgerflow/services/transaction/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionService struct {
	repo domain.TransactionRepository
}

func NewTransactionService(repo domain.TransactionRepository) *TransactionService {
	return &TransactionService{
		repo: repo,
	}
}

func (s *TransactionService) Transfer(ctx context.Context, fromAccountID uuid.UUID, toAccountID uuid.UUID, amount decimal.Decimal, currency string) (*domain.Transaction, error) {

	transaction := domain.Transaction{
		ID: uuid.New(),
		FromAccountID: fromAccountID,
		ToAccountID: toAccountID,
		Amount: amount,
		Currency: currency,
		Status: domain.TransactionStatusPending,
	}

	if err := s.repo.Create(ctx, &transaction); err != nil {
		return nil, fmt.Errorf("TransactionService - Transfer - repo.Create: %w", err)
	}

	return &transaction, nil
}