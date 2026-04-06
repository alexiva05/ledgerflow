package main

import (
	"context"
	txpb "ledgerflow/proto/transaction"
	"ledgerflow/services/transaction/internal/app"
	"ledgerflow/services/transaction/internal/infra/postgres"
	grpcserver "ledgerflow/services/transaction/internal/transport/grpc"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		os.Exit(1)
	}
	repo := postgres.NewTransactionRepo(pool)
	service := app.NewTransactionService(repo)
	grpcHandler := grpcserver.NewGRPCServer(service)
	grpcSrv := grpc.NewServer()
	txpb.RegisterTransactionServiceServer(grpcSrv, grpcHandler)
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		os.Exit(1)
	}

	grpcSrv.Serve(lis)
}
