package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/maklybae/goshop/payment/internal/domain"
	"github.com/maklybae/goshop/shared/inbox"
	"github.com/maklybae/goshop/shared/outbox"
)

type PaymentHandler struct {
	repo   domain.Repository
	outbox outbox.Repository
}

func NewPaymentHandler(repo domain.Repository, outboxRepo outbox.Repository) *PaymentHandler {
	log.Printf("[PaymentHandler] Initializing PaymentHandler")
	return &PaymentHandler{repo: repo, outbox: outboxRepo}
}

// paymentEvent соответствует формату события из order
type paymentEvent struct {
	OrderID uuid.UUID `json:"order_id"`
	UserID  uuid.UUID `json:"user_id"`
	Amount  int64     `json:"amount"`
}

type paymentCompletedEvent struct {
	OrderID uuid.UUID `json:"order_id"`
	UserID  uuid.UUID `json:"user_id"`
}

func (h *PaymentHandler) Handle(ctx context.Context, msg *inbox.Message) error {
	log.Printf("[PaymentHandler] Handle called: messageID=%s eventType=%s", msg.ID, msg.EventType)
	var event paymentEvent
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		log.Printf("[PaymentHandler] Failed to unmarshal payment event: %v", err)
		return fmt.Errorf("failed to unmarshal payment event: %w", err)
	}

	acc, err := h.repo.GetByID(ctx, event.UserID)
	if err != nil {
		log.Printf("[PaymentHandler] Failed to get account: %v", err)
		return fmt.Errorf("failed to get account: %w", err)
	}

	if acc.Amount < event.Amount {
		log.Printf("[PaymentHandler] Insufficient funds: userID=%s, have=%d, need=%d", event.UserID, acc.Amount, event.Amount)
		return errors.New("insufficient funds")
	}

	if err := h.repo.ChangeBalance(ctx, event.UserID, -event.Amount); err != nil {
		log.Printf("[PaymentHandler] Failed to withdraw funds: %v", err)
		return fmt.Errorf("failed to withdraw funds: %w", err)
	}

	completed := paymentCompletedEvent{
		OrderID: event.OrderID,
		UserID:  event.UserID,
	}
	payload, err := json.Marshal(completed)
	if err != nil {
		log.Printf("[PaymentHandler] Failed to marshal paymentCompletedEvent: %v", err)
		return fmt.Errorf("failed to marshal paymentCompletedEvent: %w", err)
	}

	outboxMsg := &outbox.Message{
		ID:        uuid.New(),
		EventType: "payment.completed",
		Payload:   payload,
	}
	if err := h.outbox.Add(ctx, outboxMsg); err != nil {
		log.Printf("[PaymentHandler] Failed to add message to outbox: %v", err)
		return fmt.Errorf("failed to add message to outbox: %w", err)
	}
	log.Printf("[PaymentHandler] Payment processed successfully: orderID=%s userID=%s", event.OrderID, event.UserID)
	return nil
}
