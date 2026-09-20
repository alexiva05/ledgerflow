package main

import (
	"context"
	"errors"
	"fmt"
	"ledgerflow/pkg/logger"
	"ledgerflow/pkg/metrics"
	txpb "ledgerflow/proto/transaction"
	"ledgerflow/services/transaction/internal/app"
	"ledgerflow/services/transaction/internal/infra/postgres"
	grpcserver "ledgerflow/services/transaction/internal/transport/grpc"
	"net"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	log, err := logger.New("info")
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	metrics.Register()
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsAddr := os.Getenv("METRICS_ADDR")
	if metricsAddr == "" {
		metricsAddr = ":9091"
	}

	metricsServer := &http.Server{
		Addr:    metricsAddr,
		Handler: metricsMux,
	}

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		os.Exit(1)
	}
	repo := postgres.NewTransactionRepo(pool)
	service := app.NewTransactionService(repo)
	grpcHandler := grpcserver.NewGRPCServer(service)
	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(metrics.GRPCMetricsInterceptor()),
	)
	txpb.RegisterTransactionServiceServer(grpcSrv, grpcHandler)
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		os.Exit(1)
	}

	go func() {
		err := metricsServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("metrics server failed", zap.Error(err))
		}
	}()
	if err := grpcSrv.Serve(lis); err != nil {
		log.Error("grpc server failed", zap.Error(err))
	}
}
