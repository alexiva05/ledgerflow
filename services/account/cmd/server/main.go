package main

import (
	"context"
	"fmt"
	"ledgerflow/pkg/metrics"
	"ledgerflow/services/account/internal/app"
	"ledgerflow/services/account/internal/infra/postgres"
	"ledgerflow/services/account/internal/transport/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8080"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger init: %v\n", err)
		os.Exit(1)
	}

	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Fatal("db init", zap.Error(err))
	}
	defer db.Close()

	repo := postgres.NewAccountRepo(db)
	accountService := app.NewAccountService(repo)
	handler := http.NewHandler(accountService)

	server := gin.Default()
	metrics.Register()
	server.Use(metrics.PrometheusMiddleware())
	server.GET("/metrics", gin.WrapH(promhttp.Handler()))
	handler.RegisterRoutes(server)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := server.Run(httpAddr)
		logger.Error("server error", zap.Error(err))
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	cancel()
	wg.Wait()
}