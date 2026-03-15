package domain

import (
	"time"

	"github.com/google/uuid"
)

type AccountStatus string
const (
	StatusActive AccountStatus = "active"
	StatusFrozen AccountStatus = "frozen"
	StatusClosed AccountStatus = "closed"
)

type Account struct {
	ID uuid.UUID
	Owner uuid.UUID
	Currency string
	Status AccountStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}