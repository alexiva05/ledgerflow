package domain

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type AuditEntryRepository interface {
	Save(ctx context.Context, tx pgx.Tx, entry *AuditEntry) error
	GetLast(ctx context.Context, tx pgx.Tx) (*AuditEntry, error)
}