package service

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/maklybae/goshop/payment/internal/domain"
)

type PaymentService struct {
	repo domain.Repository
}

func NewPaymentService(repo domain.Repository) *PaymentService {
	log.Printf("[PaymentService] Initializing PaymentService")
	return &PaymentService{repo: repo}
}

// CreateAccount создает счет для пользователя, если его еще нет
func (s *PaymentService) CreateAccount(ctx context.Context, userID uuid.UUID) error {
	log.Printf("[PaymentService] CreateAccount called: userID=%s", userID)
	_, err := s.repo.GetByID(ctx, userID)
	if err == nil {
		log.Printf("[PaymentService] Account already exists: userID=%s", userID)
		return errors.New("account already exists")
	}
	acc := &domain.Account{
		UserID: userID,
		Amount: 0,
	}
	if err := s.repo.Create(ctx, acc); err != nil {
		log.Printf("[PaymentService] Failed to create account: %v", err)
		return err
	}
	log.Printf("[PaymentService] Account created: userID=%s", userID)
	return nil
}

// Deposit пополняет счет пользователя
func (s *PaymentService) Deposit(ctx context.Context, userID uuid.UUID, amount int64) error {
	log.Printf("[PaymentService] Deposit called: userID=%s amount=%d", userID, amount)
	_, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		log.Printf("[PaymentService] Account not found: userID=%s", userID)
		return errors.New("account not found")
	}
	if err := s.repo.ChangeBalance(ctx, userID, amount); err != nil {
		log.Printf("[PaymentService] Failed to change balance: %v", err)
		return err
	}
	log.Printf("[PaymentService] Balance changed: userID=%s amount=%d", userID, amount)
	return nil
}

// GetBalance возвращает баланс пользователя
func (s *PaymentService) GetBalance(ctx context.Context, userID uuid.UUID) (int64, error) {
	log.Printf("[PaymentService] GetBalance called: userID=%s", userID)
	acc, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		log.Printf("[PaymentService] Account not found: userID=%s", userID)
		return 0, errors.New("account not found")
	}
	log.Printf("[PaymentService] Balance: userID=%s amount=%d", userID, acc.Amount)
	return acc.Amount, nil
}
