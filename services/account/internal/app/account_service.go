package app

import (
	"context"
	"fmt"
	"ledgerflow/services/account/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	pkgerrors "ledgerflow/pkg/errors"
)

type AccountService struct {
	repo domain.AccountRepository
}

func NewAccountService(repo domain.AccountRepository) *AccountService {
	return &AccountService{
		repo: repo,
	}
}

func (a *AccountService) CreateAccount(ctx context.Context, owner uuid.UUID, currency string) (*domain.Account, error) {
	
	if currency == "" {
		return nil, fmt.Errorf("accountService.CreateAccount: currency is required: %w", pkgerrors.ErrInvalidInput)
	}
	
	account := &domain.Account{
		ID: uuid.New(),
		Owner: owner,
		Currency: currency,
		Status: domain.StatusActive,
	}

	err := a.repo.Create(ctx, *account)
	if err != nil {
		return nil, fmt.Errorf("accountService.CreateAccount: %w", err)
	}

	return account, nil
}

func (a *AccountService) GetBalance(ctx context.Context, accountID uuid.UUID) (decimal.Decimal, error) {

	_, err := a.repo.GetByID(ctx, accountID)
	if err != nil {
		return decimal.Zero, fmt.Errorf("accountService.GetBalance: %w", err)
	}

	entries, err := a.repo.GetJournalEntries(ctx, accountID)
	if err != nil {
		return decimal.Zero, fmt.Errorf("accountService.GetBalance: %w", err)
	}

	balance := decimal.Zero

	for _, e := range entries {
		switch e.Direction {
		case domain.DirectionCredit:
			balance = balance.Add(e.Amount)
		case domain.DirectionDebit:
			balance = balance.Sub(e.Amount)
		default:
			return decimal.Zero, fmt.Errorf("unknown direction: %s", e.Direction)
		}
	}

	return balance, nil
}