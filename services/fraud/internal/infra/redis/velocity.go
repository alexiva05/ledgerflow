package redis

import (
	"context"
	"fmt"
	"ledgerflow/services/fraud/internal/domain"
	"time"

	"github.com/google/uuid"
	redis "github.com/redis/go-redis/v9"
)

type velocityChecker struct {
	client    *redis.Client
	window    time.Duration
	threshold int64
}

func NewVelocityChecker(client *redis.Client, window time.Duration, threshold int64) domain.VelocityCheckerRepository {
	return &velocityChecker{
		client: client,
		window: window,
		threshold: threshold,
	}
}

func (v *velocityChecker) Check(ctx context.Context, accountID uuid.UUID, txID uuid.UUID, at time.Time) (bool, error) {
	key := fmt.Sprintf("velocity:%s", accountID)
	score := float64(at.Unix())
	minScore := at.Add(-v.window).Unix()

	err := v.client.ZAdd(ctx, key, redis.Z{
		Score: score,
		Member: txID.String(),
	}).Err()
	if err != nil {
		return false, fmt.Errorf("velocityChecker - Check() - zadd: %w", err)
	}

	err = v.client.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", minScore)).Err()
	if err != nil {
		return false, fmt.Errorf("velocityChecker - Check() - zremrangebyscore: %w", err)
	}

	count, err := v.client.ZCard(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("velocityChecker - Check() - zcard: %w", err)
	}

	return count > v.threshold, nil
}
