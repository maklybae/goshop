package grpc

import (
	"context"

	"github.com/google/uuid"
	paymentpb "github.com/maklybae/goshop/payment/genproto/payment"
	"github.com/maklybae/goshop/payment/internal/service"
)

type PaymentGRPCServer struct {
	paymentpb.UnimplementedPaymentServiceServer
	service *service.PaymentService
}

func NewPaymentGRPCServer(svc *service.PaymentService) *PaymentGRPCServer {
	return &PaymentGRPCServer{service: svc}
}

func (h *PaymentGRPCServer) CreateAccount(ctx context.Context, req *paymentpb.CreateAccountRequest) (*paymentpb.AccountResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, err
	}
	if err := h.service.CreateAccount(ctx, userID); err != nil {
		return nil, err
	}
	return &paymentpb.AccountResponse{UserId: req.GetUserId(), Amount: 0}, nil
}

func (h *PaymentGRPCServer) Deposit(ctx context.Context, req *paymentpb.DepositRequest) (*paymentpb.AccountResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, err
	}
	if err := h.service.Deposit(ctx, userID, req.GetAmount()); err != nil {
		return nil, err
	}
	amount, err := h.service.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &paymentpb.AccountResponse{UserId: req.GetUserId(), Amount: amount}, nil
}

func (h *PaymentGRPCServer) GetBalance(ctx context.Context, req *paymentpb.GetBalanceRequest) (*paymentpb.AccountResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, err
	}
	amount, err := h.service.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &paymentpb.AccountResponse{UserId: req.GetUserId(), Amount: amount}, nil
}
