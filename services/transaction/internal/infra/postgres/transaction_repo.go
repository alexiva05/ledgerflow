package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"ledgerflow/services/transaction/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type transactionRepo struct {
	pool *pgxpool.Pool
}

func NewTransactionRepo(pool *pgxpool.Pool) domain.TransactionRepository {
	return &transactionRepo{
		pool: pool,
	}
}

func (trx *transactionRepo) Create(ctx context.Context, transaction *domain.Transaction) error {

	tx, err := trx.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("transactionRepo - Create: %w", err)
	}
	defer tx.Rollback(ctx)

	q1 := `
		INSERT INTO transactions (id, from_account_id, to_account_id, amount, currency, status)
		VALUES
		($1, $2, $3, $4, $5, $6)
	`

	_, err = tx.Exec(ctx, q1, transaction.ID, transaction.FromAccountID,
		transaction.ToAccountID, transaction.Amount,
		transaction.Currency, transaction.Status)
	if err != nil {
		return fmt.Errorf("transactionRepo - Create: %w", err)
	}

	payload, err := json.Marshal(transaction)
	if err != nil {
		return fmt.Errorf("transactionRepo - Create - Kafka payload: %w", err)
	}

	q2 := `
		INSERT INTO outbox (id, topic, key, payload)
		VALUES
		($1, $2, $3, $4)
	`

	_, err = tx.Exec(ctx, q2, uuid.New(), "transaction.created",
		transaction.FromAccountID.String(), payload)
	if err != nil {
		return fmt.Errorf("transactionRepo - Create: %w", err)
	}

	return tx.Commit(ctx)
}

func (trx *transactionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {

	var transaction domain.Transaction

	q := `
		SELECT id, from_account_id, to_account_id, amount, currency, status, created_at, completed_at
		FROM transactions
		WHERE transactions.id = $1
	`

	err := trx.pool.QueryRow(ctx, q, id).Scan(
		&transaction.ID,
		&transaction.FromAccountID,
		&transaction.ToAccountID,
		&transaction.Amount,
		&transaction.Currency,
		&transaction.Status,
		&transaction.CreatedAt,
		&transaction.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("transactionRepo - GetByID: %w", err)
	}

	return &transaction, nil
}

func (trx *transactionRepo) UpdateStatus(ctx context.Context, transactionID uuid.UUID, status domain.TransactionStatus) error {

	q := `
		UPDATE transactions
		SET
			status = $1,
			completed_at = CASE WHEN $1::transaction_status = 'completed' THEN NOW() ELSE completed_at END
		WHERE id = $2
	`

	res, err := trx.pool.Exec(ctx, q, status, transactionID)
	if err != nil {
		return fmt.Errorf("transactionRepo - UpdateStatus: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
