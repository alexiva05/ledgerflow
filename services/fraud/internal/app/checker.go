package app

import (
	"context"
	"encoding/json"
	"fmt"
	"ledgerflow/services/fraud/internal/domain"

	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaProducer interface {
	ProduceSync(ctx context.Context, rs ...*kgo.Record) kgo.ProduceResults
}

type FraudChecker struct {
	velocity domain.VelocityChecker
	producer KafkaProducer
}

func NewFraudChecker(velocity domain.VelocityChecker, producer KafkaProducer) *FraudChecker {
	return &FraudChecker{
		velocity: velocity,
		producer: producer,
	}
}

func (f *FraudChecker) Check(ctx context.Context, event *domain.TransactionEvent) error {

	flag, err := f.velocity.Check(ctx, event.AccountID, event.TransactionID, event.Timestamp)
	if err != nil {
		return fmt.Errorf("FraudChecker - Check() - velocity.Check(): %w", err)
	}
	if flag {
		payload, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("FraudChecker - Check() - json.Marshal(): %w", err)
		}
		results := f.producer.ProduceSync(ctx, &kgo.Record{
			Key: []byte(event.AccountID.String()),
			Value: payload,
			Topic: "fraud.alert",
		})
		if results.FirstErr() != nil {
			return fmt.Errorf("FraudChecker - Check() - results.FirstError(): %w", results.FirstErr())
		}
	}

	return nil
}