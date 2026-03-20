package domain

import "errors"

var ErrNotFound = errors.New("not found")
var ErrInsufficientFunds = errors.New("insufficient funds")
var ErrDuplicateIdempotencyKey = errors.New("idempotency key already exists")