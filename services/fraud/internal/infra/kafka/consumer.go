package kafka

import (
	"context"
	"encoding/json"
	"ledgerflow/services/fraud/internal/app"
	"ledgerflow/services/fraud/internal/domain"

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
							var event domain.TransactionEvent
							if err := json.Unmarshal(record.Value, &event); err != nil {
								c.logger.Error("unmarshal json failed", zap.Error(err))
								continue
							}
							if err := c.checker.Check(ctx, &event); err != nil {
								c.logger.Error("fraud check failed", zap.Error(err))
								continue
							}
						}
					}
				}
			}
		}
	}
}
