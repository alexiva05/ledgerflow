package postgres

import (
	"context"
	"errors"
	"fmt"
	"ledgerflow/services/audit/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type auditRepo struct {
	db *pgxpool.Pool
}

func NewAuditRepo(db *pgxpool.Pool) *auditRepo {
	return &auditRepo{
		db: db,
	}
}

func (a *auditRepo) Save(ctx context.Context, tx pgx.Tx, entry *domain.AuditEntry) error {

	q := `
		INSERT INTO audit_log (trace_id, topic, event_type, payload, hmac, prev_hmac)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := tx.Exec(ctx, q,
						entry.TraceID,
						entry.Topic,
						entry.EventType,
						entry.Payload,
						entry.Hmac,
						entry.PrevHmac)
	
	if err != nil {
		return fmt.Errorf("auditRepo - Save(): %w", err)
	}

	return nil
}

func (a *auditRepo) GetLast(ctx context.Context, tx pgx.Tx) (*domain.AuditEntry, error) {

	q := `
		SELECT id, trace_id, topic, event_type, payload, hmac, prev_hmac, created_at
		FROM audit_log
		ORDER BY created_at DESC
		LIMIT 1 FOR UPDATE
	`

	row := tx.QueryRow(ctx, q)
	entry := &domain.AuditEntry{}
	err := row.Scan(&entry.ID, &entry.TraceID, &entry.Topic, &entry.EventType, &entry.Payload, &entry.Hmac, &entry.PrevHmac, &entry.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("auditRepo - GetLast(): %w", err)
	}

	return entry, nil
}