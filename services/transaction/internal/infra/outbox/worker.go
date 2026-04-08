package outbox

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type KafkaProducer interface {
	ProduceSync(ctx context.Context, rs ...*kgo.Record) kgo.ProduceResults
}

type Worker struct {
	db       *pgxpool.Pool
	kafka    KafkaProducer
	interval time.Duration
	logger   *zap.Logger
}

func NewWorker(db *pgxpool.Pool, kafka KafkaProducer, interval time.Duration, logger *zap.Logger) *Worker {
	return &Worker{
		db:       db,
		kafka:    kafka,
		interval: interval,
		logger:   logger,
	}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				w.logger.Error("outbox batch failed", zap.Error(err))
			}
		}
	}
}

func (w *Worker) processBatch(ctx context.Context) error {

	q := `
		SELECT id, topic, key, payload
		FROM outbox
		WHERE sent_at IS NULL
		ORDER BY created_at
		LIMIT 100
	`

	rows, err := w.db.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("Worker - processBatch: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		var topic, key string
		var payload []byte
		err = rows.Scan(&id, &topic, &key, &payload)
		if err != nil {
			return fmt.Errorf("Worker - processBatch: %w", err)
		}
		record := &kgo.Record{
			Topic: topic,
			Key: []byte(key),
			Value: payload,
		}
		results := w.kafka.ProduceSync(ctx, record)
		if err := results.FirstErr(); err != nil {
			w.logger.Error("failed to publish outbox event",
				zap.Error(err),
				zap.String("event_id", id.String()),
				zap.String("topic", topic),
			)	
			continue
		}
		if _, err := w.db.Exec(ctx, `UPDATE outbox SET sent_at = NOW() WHERE id = $1`, id); err != nil {
			w.logger.Warn("failed to mark outbox event as sent", zap.Error(err), zap.String("event_id", id.String()))
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("Worker - processBatch - rows: %w", err)
	}

	return nil
}
