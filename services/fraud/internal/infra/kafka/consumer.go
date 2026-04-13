package kafka

import (
	"context"
	"encoding/json"
	"ledgerflow/pkg/metrics"
	"ledgerflow/services/fraud/internal/app"
	"ledgerflow/services/fraud/internal/domain"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type Consumer struct {
	client  *kgo.Client
	checker *app.FraudChecker
	logger  *zap.Logger
}

func NewConsumer(client *kgo.Client, checker *app.FraudChecker, logger *zap.Logger) *Consumer {
	return &Consumer{
		client:  client,
		checker: checker,
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
			c.logger.Error("poll fetches failed", zap.Error(err))
		}
		for _, fetch := range fetches {
			for _, fetchTopic := range fetch.Topics {
				if fetchTopic.Topic == "transaction.created" {
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
	if err := c.checker.Check(ctx, &event); err != nil {
		c.logger.Error("fraud check failed", zap.Error(err))
		status = "error"
		return
	}
}
