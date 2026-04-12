package app

import (
	"context"
	"fmt"
	"ledgerflow/services/notification/internal/domain"
)

type NotificationService struct {
	notifier domain.Notifier
}

func NewNotificationService(notifier domain.Notifier) *NotificationService {
	return &NotificationService{
		notifier: notifier,
	}
}

func (n *NotificationService) Handle(ctx context.Context, event domain.TransactionEvent, topic string) error {

	switch topic {
	case string(domain.TransactionCompleted):
		err := n.notifier.Send(ctx, event.AccountID, fmt.Sprintf("Ваш перевод выполнен: %s", event.Amount))
		if err != nil {
			return fmt.Errorf("NotificationService - Handle(): %w", err)
		}
		return nil
	case string(domain.TransactionFailed):
		err := n.notifier.Send(ctx, event.AccountID, "Ваш перевод не удался")
		if err != nil {
			return fmt.Errorf("NotificationService - Handle(): %w", err)
		}
		return nil
	default:
		return fmt.Errorf("undefined topic: %s", topic)
	}
}