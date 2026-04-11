package domain

import (
	"time"

	"github.com/google/uuid"
)

type AuditEventType string

const (
	TransactionCreated   AuditEventType = "transaction.created"
	TransactionCompleted AuditEventType = "transaction.completed"
	TransactionFailed    AuditEventType = "transaction.failed"
	BalanceUpdated       AuditEventType = "balance.updated"
	FraudAlert           AuditEventType = "fraud.alert"
)

type AuditEntry struct {
	ID        uuid.UUID
	TraceID   string
	Topic     string
	EventType AuditEventType
	Payload   []byte
	Hmac      string
	PrevHmac  *string
	CreatedAt time.Time
}
