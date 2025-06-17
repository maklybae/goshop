package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maklybae/goshop/payment/internal/domain"
	"github.com/maklybae/goshop/shared/txs"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	log.Printf("[PaymentRepository] Initializing repository")
	return &Repository{db: db}
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	log.Printf("[PaymentRepository] GetByID called: userID=%s", id)
	querier := txs.GetQuerier(ctx, r.db)
	acc := &domain.Account{}
	err := querier.QueryRow(ctx, "SELECT user_id, amount FROM accounts WHERE user_id = $1", id).
		Scan(&acc.UserID, &acc.Amount)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[PaymentRepository] Account not found: userID=%s", id)
			return nil, fmt.Errorf("account not found: %w", err)
		}
		log.Printf("[PaymentRepository] Failed to get account: %v", err)
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	log.Printf("[PaymentRepository] Account found: userID=%s amount=%d", acc.UserID, acc.Amount)
	return acc, nil
}

func (r *Repository) ChangeBalance(ctx context.Context, userID uuid.UUID, amount int64) error {
	log.Printf("[PaymentRepository] ChangeBalance called: userID=%s amount=%d", userID, amount)
	querier := txs.GetQuerier(ctx, r.db)
	_, err := querier.Exec(ctx, "UPDATE accounts SET amount = amount + $1 WHERE user_id = $2", amount, userID)
	if err != nil {
		log.Printf("[PaymentRepository] Failed to change balance: %v", err)
		return fmt.Errorf("failed to change balance: %w", err)
	}
	log.Printf("[PaymentRepository] Balance changed: userID=%s amount=%d", userID, amount)
	return nil
}

func (r *Repository) List(ctx context.Context) ([]*domain.Account, error) {
	log.Printf("[PaymentRepository] List called")
	querier := txs.GetQuerier(ctx, r.db)
	rows, err := querier.Query(ctx, "SELECT user_id, amount FROM accounts")
	if err != nil {
		log.Printf("[PaymentRepository] Failed to list accounts: %v", err)
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*domain.Account
	for rows.Next() {
		acc := &domain.Account{}
		if err := rows.Scan(&acc.UserID, &acc.Amount); err != nil {
			log.Printf("[PaymentRepository] Failed to scan account: %v", err)
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, acc)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[PaymentRepository] Rows error: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}
	log.Printf("[PaymentRepository] Accounts found: %d", len(accounts))
	return accounts, nil
}

func (r *Repository) Create(ctx context.Context, account *domain.Account) error {
	log.Printf("[PaymentRepository] Create called: userID=%s amount=%d", account.UserID, account.Amount)
	querier := txs.GetQuerier(ctx, r.db)
	_, err := querier.Exec(ctx, "INSERT INTO accounts (user_id, amount) VALUES ($1, $2)", account.UserID, account.Amount)
	if err != nil {
		log.Printf("[PaymentRepository] Failed to create account: %v", err)
		return fmt.Errorf("failed to create account: %w", err)
	}
	log.Printf("[PaymentRepository] Account created: userID=%s", account.UserID)
	return nil
}
