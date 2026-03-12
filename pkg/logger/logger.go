package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type contextKey struct{}

func New(level string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()

	err := cfg.Level.UnmarshalText([]byte(level))
	if err != nil {
		return nil, fmt.Errorf("couldn't get logger level: %w", err)
	}

	return cfg.Build()
}

func WithContext(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

func FromContext(ctx context.Context) *zap.Logger {
	val := ctx.Value(contextKey{})
	l, ok := val.(*zap.Logger)

	if !ok {
		return zap.NewNop()
	}

	return l
}