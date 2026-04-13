package main

import (
	"context"
	"fmt"
	"ledgerflow/pkg/logger"
	"ledgerflow/pkg/metrics"
	"ledgerflow/services/audit/internal/app"
	"ledgerflow/services/audit/internal/infra/kafka"
	"ledgerflow/services/audit/internal/infra/postgres"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

func main() {
	auditHmacKey := os.Getenv("AUDIT_HMAC_KEY")
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := logger.New("info")
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		os.Exit(1)
	}

	metrics.Register()

	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Fatal("init postgres", zap.Error(err))
	}
	defer db.Close()

	consumerClient, err := kgo.NewClient(
		kgo.SeedBrokers(kafkaBrokers),
		kgo.ConsumeTopics(
			"transaction.created",
			"transaction.completed",
			"transaction.failed",
			"balance.updated",
			"fraud.alert",
		),
	)
	if err != nil {
		logger.Fatal("init kafka consumer", zap.Error(err))
	}
	defer consumerClient.Close()

	auditRepo := postgres.NewAuditRepo(db)
	auditService := app.NewAuditService(db, auditRepo, []byte(auditHmacKey))
	consumer := kafka.NewConsumer(consumerClient, auditService, logger)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":9090", mux)
	}()

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
