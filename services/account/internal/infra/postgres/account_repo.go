package postgres

import (
	"context"
	"errors"
	"fmt"
	"ledgerflow/services/account/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type accountRepo struct {
	pool *pgxpool.Pool
}

func NewAccountRepo(pool *pgxpool.Pool) domain.AccountRepository {
	return &accountRepo{
		pool: pool,
	}
}

func (a *accountRepo) Create(ctx context.Context, account domain.Account) error {

	q := `
		INSERT INTO accounts (id, owner, currency, status)
		VALUES 
		($1, $2, $3, $4)
	`

	_, err := a.pool.Exec(ctx, q, account.ID, account.Owner, account.Currency, account.Status)
	if err != nil {
		return fmt.Errorf("accountRepo.Create: %w", err)
	}

	return nil
}

func (a *accountRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {

	var account domain.Account

	q := `
		SELECT id, owner, currency, status, created_at, updated_at FROM accounts
		WHERE id = $1
	`

	err := a.pool.QueryRow(ctx, q, id).Scan(
		&account.ID,
		&account.Owner,
		&account.Currency,
		&account.Status,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("accountRepo.GetByID: %w", err)
	}

	return &account, nil
}

func (a *accountRepo) GetJournalEntries(ctx context.Context, accountID uuid.UUID) ([]*domain.JournalEntry, error) {

	q := `
		SELECT id, account_id, transaction_id, direction, amount, created_at
		FROM journal_entries
		WHERE account_id = $1
	`
	rows, err := a.pool.Query(ctx, q, accountID)
	if err != nil {
		return nil, fmt.Errorf("accountRepo.GetJournalEntries: %w", err)
	}
	defer rows.Close()

	var journalEntries []*domain.JournalEntry
	for rows.Next() {

		var j domain.JournalEntry

		err := rows.Scan(
			&j.ID,
			&j.AccountID,
			&j.TransactionID,
			&j.Direction,
			&j.Amount,
			&j.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("accountRepo.GetJournalEntries: %w", err)
		}

		journalEntries = append(journalEntries, &j)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("accountRepo.GetJournalEntries scan: %w", err)
	}

	return journalEntries, nil
}
