package grpc

import (
	"context"
	txpb "ledgerflow/proto/transaction"
	"ledgerflow/services/transaction/internal/app"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	txpb.UnimplementedTransactionServiceServer
	service *app.TransactionService
}

func NewGRPCServer(service *app.TransactionService) *grpcServer {
	return &grpcServer{
		service: service,
	}
}

func (s *grpcServer) Transfer(ctx context.Context, req *txpb.TransferRequest) (*txpb.TransferResponse, error) {

	fromID, err := uuid.Parse(req.FromAccountId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid from_account_id: %v", err)
	}

	toID, err := uuid.Parse(req.ToAccountId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid to_account_id: %v", err)
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid amount %v", err)
	}

	transaction, err := s.service.Transfer(ctx, fromID, toID, amount, req.Currency)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "transfer failed: %v", err)
	}

	resp := &txpb.TransferResponse{
		TransactionId: transaction.ID.String(),
		Status:        string(transaction.Status),
	}

	return resp, nil
}
