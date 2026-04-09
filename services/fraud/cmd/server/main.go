package main

import (
	"context"
	"fmt"
	"ledgerflow/pkg/logger"
	"ledgerflow/services/fraud/internal/app"
	kafkaconsumer "ledgerflow/services/fraud/internal/infra/kafka"
	rds "ledgerflow/services/fraud/internal/infra/redis"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

func main() {
	redisAddr := os.Getenv("REDIS_URL")
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	logger, err := logger.New("info")
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		os.Exit(1)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	consumerClient, err := kgo.NewClient(
		kgo.SeedBrokers(kafkaBrokers),
		kgo.ConsumeTopics("transaction.created"),
	)
	if err != nil {
		logger.Fatal("init kafka consumer", zap.Error(err))
	}
	defer consumerClient.Close()
	
	producerClient, err := kgo.NewClient(
		kgo.SeedBrokers(kafkaBrokers),
	)
	if err != nil {
		logger.Fatal("init kafka producer", zap.Error(err))
	}
	defer producerClient.Close()

	velocityChecker := rds.NewVelocityChecker(redisClient, 60*time.Second, 5)
	fraudChecker := app.NewFraudChecker(velocityChecker, producerClient)
	consumer := kafkaconsumer.NewConsumer(consumerClient, fraudChecker, logger)

	wg.Add(1)
	go func() {
		defer wg.Done()
		consumer.Run(ctx)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	wg.Wait()
}