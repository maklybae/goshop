package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/maklybae/goshop/order/internal/domain"
	"github.com/maklybae/goshop/shared/outbox"
	"github.com/maklybae/goshop/shared/txs"
)

type OrderService struct {
	outbox     outbox.Repository
	transactor txs.Transactor
	repository domain.OrderRepository
}

func NewOrderService(outboxRepo outbox.Repository, transactor txs.Transactor, orderRepo domain.OrderRepository) *OrderService {
	log.Printf("[OrderService] Initializing OrderService")
	return &OrderService{
		outbox:     outboxRepo,
		transactor: transactor,
		repository: orderRepo,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID uuid.UUID, description string, amount int64) (*domain.Order, error) {
	log.Printf("[OrderService] CreateOrder called: userID=%s, description=%s, amount=%d", userID, description, amount)
	order := &domain.Order{
		ID:          uuid.New(),
		UserID:      userID,
		Description: description,
		Amount:      amount,
		Status:      domain.StatusNew,
	}

	if err := s.repository.Create(ctx, order); err != nil {
		log.Printf("[OrderService] Failed to create order: %v", err)
		return nil, err
	}

	event := struct {
		OrderID uuid.UUID `json:"order_id"`
		UserID  uuid.UUID `json:"user_id"`
		Amount  int64     `json:"amount"`
	}{
		OrderID: order.ID,
		UserID:  order.UserID,
		Amount:  order.Amount,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("[OrderService] Failed to marshal event: %v", err)
		return nil, err
	}

	msg := &outbox.Message{
		ID:        uuid.New(),
		EventType: "order.payments",
		Payload:   payload,
	}
	if err := s.outbox.Add(ctx, msg); err != nil {
		log.Printf("[OrderService] Failed to add message to outbox: %v", err)
		return nil, err
	}

	log.Printf("[OrderService] Order created successfully: %+v", order)
	return order, nil
}

func (s *OrderService) GetOrders(ctx context.Context, userID uuid.UUID) ([]domain.Order, error) {
	log.Printf("[OrderService] GetOrders called: userID=%s", userID)
	orders, err := s.repository.List(ctx, userID)
	if err != nil {
		log.Printf("[OrderService] Failed to get orders: %v", err)
	}
	return orders, err
}

func (s *OrderService) GetOrderStatus(ctx context.Context, orderID uuid.UUID) (domain.OrderStatus, error) {
	log.Printf("[OrderService] GetOrderStatus called: orderID=%s", orderID)
	order, err := s.repository.Get(ctx, orderID)
	if err != nil {
		log.Printf("[OrderService] Failed to get order: %v", err)
		return "", err
	}
	log.Printf("[OrderService] Order status: %s", order.Status)
	return order.Status, nil
}

func (s *OrderService) HandlePaymentEvent(ctx context.Context, msg *sarama.ConsumerMessage) error {
	var event struct {
		OrderID uuid.UUID `json:"order_id"`
		UserID  uuid.UUID `json:"user_id"`
	}

	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal payment event: %w", err)
	}

	return s.transactor.WithTransaction(ctx, func(ctx context.Context) error {
		order, err := s.repository.Get(ctx, event.OrderID)
		if err != nil {
			return err
		}

		order.Status = domain.StatusFinished
		if err := s.repository.Update(ctx, order); err != nil {
			return err
		}

		return nil
	})
}
