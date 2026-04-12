package domain

import (
	"context"

	"github.com/google/uuid"
)

type Notifier interface {
	Send(ctx context.Context, userID uuid.UUID, message string) error
}