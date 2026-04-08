package app

import (
	"context"
	"ledgerflow/services/transaction/internal/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	CreateFn func(ctx context.Context, t *domain.Transaction) error
}

func (m *mockRepo) Create(ctx context.Context, t *domain.Transaction) error {
	return m.CreateFn(ctx, t)
}

func (m *mockRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	panic("not implemented")
}

func (m *mockRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.TransactionStatus) error {
	panic("not implemented")
}

func TestTransfer_Success(t *testing.T) {
	repo := & mockRepo{
		CreateFn: func(ctx context.Context, t *domain.Transaction) error{
			return nil
		},
	}

	svc := NewTransactionService(repo)

	result, err := svc.Transfer(
		context.Background(),
		uuid.New(),
		uuid.New(),
		decimal.NewFromInt(100),
		"USD",
	)

	require.NoError(t, err)
	assert.Equal(t, decimal.NewFromInt(100), result.Amount)
	assert.Equal(t, domain.TransactionStatusPending, result.Status)
}

func TestTransfer_InsufficientFunds(t *testing.T) {
	repo := &mockRepo{
		CreateFn: func(ctx context.Context, t *domain.Transaction) error {
			return domain.ErrInsufficientFunds
		},
	}

	svc := NewTransactionService(repo)
	_, err := svc.Transfer(context.Background(), uuid.New(), uuid.New(), decimal.NewFromInt(100), "USD")

	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrInsufficientFunds)
}

func TestTransfer_DuplicateIdempotencyKey(t *testing.T) {
	repo := &mockRepo{
		CreateFn: func(ctx context.Context, t *domain.Transaction) error {
			return domain.ErrDuplicateIdempotencyKey	
		},
	}

	svc := NewTransactionService(repo)

	_, err := svc.Transfer(context.Background(), uuid.New(), uuid.New(), decimal.NewFromInt(100), "USD")
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrDuplicateIdempotencyKey)
}