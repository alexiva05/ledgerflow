package kafka

import (
	"context"
	"ledgerflow/pkg/metrics"
	"ledgerflow/services/audit/internal/app"
	"ledgerflow/services/audit/internal/domain"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type Consumer struct {
	client  *kgo.Client
	service *app.AuditService
	logger  *zap.Logger
}

func NewConsumer(client *kgo.Client, service *app.AuditService, logger *zap.Logger) *Consumer {
	return &Consumer{
		client:  client,
		service: service,
		logger:  logger,
	}
}

func (c *Consumer) Run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		fetches := c.client.PollFetches(ctx)
		if err := fetches.Err(); err != nil {
			c.logger.Error("poll fetches failed: %w", zap.Error(err))
		}
		for _, fetch := range fetches {
			for _, fetchTopic := range fetch.Topics {
				for _, partition := range fetchTopic.Partitions {
					for _, record := range partition.Records {
						c.processRecord(ctx, fetchTopic.Topic, record)
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

	var traceID string
	for _, h := range record.Headers {
		if h.Key == "trace_id" {
			traceID = string(h.Value)
			break
		}
	}

	entry := &domain.AuditEntry{
		Payload: record.Value,
		Topic: topic,
		EventType: domain.AuditEventType(topic),
		TraceID: traceID,
	}

	if err := c.service.Record(ctx, entry); err != nil {
		c.logger.Error("record failed", zap.Error(err))
		status = "error"
		return
	}
}