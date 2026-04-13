package kafka

import (
	"context"
	"encoding/json"
	"ledgerflow/pkg/metrics"
	"ledgerflow/services/notification/internal/app"
	"ledgerflow/services/notification/internal/domain"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type Consumer struct {
	client *kgo.Client
	notification *app.NotificationService
	logger *zap.Logger
}

func NewConsumer(client *kgo.Client, notification *app.NotificationService, logger *zap.Logger) *Consumer {
	return &Consumer{
		client: client,
		notification: notification,
		logger: logger,
	}
}

func (c *Consumer) Run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		fetches := c.client.PollFetches(ctx)
		if err := fetches.Err(); err != nil {
			c.logger.Error("poll fetches failed", zap.Error(err))
		}
		for _, fetch := range fetches {
			for _, fetchTopic := range fetch.Topics {
				if fetchTopic.Topic == string(domain.TransactionCompleted) || fetchTopic.Topic == string(domain.TransactionFailed) {
					for _, partition := range fetchTopic.Partitions {
						for _, record := range partition.Records {
							c.processRecord(ctx, fetchTopic.Topic, record)
						}
					}
				}
			}
		}
	}
}

func (c *Consumer) processRecord(ctx context.Context, topic string, record *kgo.Record) {
	start := time.Now()
	status := "success"
	defer func() {
		metrics.KafkaMessagesTotal.WithLabelValues(topic, status).Inc()
		metrics.KafkaMessageDuration.WithLabelValues(topic).Observe(time.Since(start).Seconds())
	}()

	var event domain.TransactionEvent
	if err := json.Unmarshal(record.Value, &event); err != nil {
		c.logger.Error("unmarshal json failed", zap.Error(err))
		status = "error"
		return
	}
	if err := c.notification.Handle(ctx, event, topic); err != nil {
		c.logger.Error("notification send failed", zap.Error(err), zap.String("topic", topic), zap.String("transaction_id", event.TransactionID.String()))
		status = "error"
		return
	}
}