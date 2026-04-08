package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type VelocityCheckerRepository interface {
	Check(ctx context.Context, accountID uuid.UUID, txID uuid.UUID, at time.Time) (bool, error)
}