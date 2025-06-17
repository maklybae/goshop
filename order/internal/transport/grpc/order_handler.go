package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/maklybae/goshop/order/genproto/order"
	"github.com/maklybae/goshop/order/internal/service"
)

type OrderGRPCServer struct {
	order.UnimplementedOrderServiceServer
	service *service.OrderService
}

func NewOrderGRPCServer(svc *service.OrderService) *OrderGRPCServer {
	return &OrderGRPCServer{service: svc}
}

func (h *OrderGRPCServer) CreateOrder(ctx context.Context, req *order.CreateOrderRequest) (*order.OrderResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, err
	}
	ord, err := h.service.CreateOrder(ctx, userID, req.GetDescription(), req.GetAmount())
	if err != nil {
		return nil, err
	}
	return &order.OrderResponse{
		Id:          ord.ID.String(),
		UserId:      ord.UserID.String(),
		Description: ord.Description,
		Amount:      ord.Amount,
		Status:      string(ord.Status),
	}, nil
}

func (h *OrderGRPCServer) ListOrders(ctx context.Context, req *order.ListOrdersRequest) (*order.ListOrdersResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, err
	}
	orders, err := h.service.GetOrders(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := &order.ListOrdersResponse{}
	for _, ord := range orders {
		resp.Orders = append(resp.Orders, &order.OrderResponse{
			Id:          ord.ID.String(),
			UserId:      ord.UserID.String(),
			Description: ord.Description,
			Amount:      ord.Amount,
			Status:      string(ord.Status),
		})
	}
	return resp, nil
}

func (h *OrderGRPCServer) GetOrderStatus(ctx context.Context, req *order.GetOrderStatusRequest) (*order.GetOrderStatusResponse, error) {
	orderID, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, err
	}
	status, err := h.service.GetOrderStatus(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return &order.GetOrderStatusResponse{Status: string(status)}, nil
}
