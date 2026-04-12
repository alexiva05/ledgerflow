package notify

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type LogNotifier struct {
	logger *zap.Logger
}

func NewLogNotifier(logger *zap.Logger) *LogNotifier {
	return &LogNotifier{
		logger: logger,
	}
}

func (l *LogNotifier) Send(ctx context.Context, userID uuid.UUID, message string) error {
	l.logger.Info(
		"transaction status",
		zap.String("user_id", userID.String()),
		zap.String("message", message),
	)

	return nil
}