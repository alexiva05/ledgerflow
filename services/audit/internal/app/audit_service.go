package app

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"ledgerflow/services/audit/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditService struct {
	db *pgxpool.Pool
	repo    domain.AuditEntryRepository
	hmacKey []byte
}

func NewAuditService(db *pgxpool.Pool, repo domain.AuditEntryRepository, hmacKey []byte) *AuditService {
	return &AuditService{
		db: db,
		repo:    repo,
		hmacKey: hmacKey,
	}
}

func (a AuditService) Record(ctx context.Context, entry *domain.AuditEntry) error {

	tx, err := a.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("AuditService - Record(): %w", err)
	}
	defer tx.Rollback(ctx)

	auditEntry, err := a.repo.GetLast(ctx, tx)
	if err != nil {
		return fmt.Errorf("AuditService - Record() - repo.GetLast(): %w", err)
	}

	var prevHmac *string
	if auditEntry != nil {
		prevHmac = &auditEntry.Hmac
	}

	mac := hmac.New(sha256.New, a.hmacKey)
	mac.Write(entry.Payload)
	if prevHmac != nil {
		mac.Write([]byte(*prevHmac))
	}
	
	computedHmac := hex.EncodeToString(mac.Sum(nil))
	entry.Hmac = computedHmac
	entry.PrevHmac = prevHmac

	err = a.repo.Save(ctx, tx, entry)
	if err != nil {
		return fmt.Errorf("AuditService - Record() - repo.Save(): %w", err)
	}

	return tx.Commit(ctx)
}
