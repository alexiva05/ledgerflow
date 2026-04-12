package main

import (
	"context"
	"fmt"
	"ledgerflow/pkg/logger"
	"ledgerflow/services/notification/internal/app"
	"ledgerflow/services/notification/internal/infra/kafka"
	"ledgerflow/services/notification/internal/infra/notify"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

func main() {
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := logger.New("info")
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		os.Exit(1)
	}

	consumerClient, err := kgo.NewClient(
		kgo.SeedBrokers(kafkaBrokers),
		kgo.ConsumeTopics("transaction.completed", "transaction.failed"),
	)
	if err != nil {
		logger.Fatal("init kafka consumer", zap.Error(err))
	}
	defer consumerClient.Close()

	logNotifier := notify.NewLogNotifier(logger)
	notificationService := app.NewNotificationService(logNotifier)
	consumer := kafka.NewConsumer(consumerClient, notificationService, logger)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumer.Run(ctx)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	wg.Wait()
}