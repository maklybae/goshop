package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maklybae/goshop/order/internal/domain"
	"github.com/maklybae/goshop/shared/txs"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	log.Printf("[OrderRepository] Initializing repository")
	return &Repository{db: db}
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (order *domain.Order, err error) {
	log.Printf("[OrderRepository] Get called: id=%s", id)
	querier := txs.GetQuerier(ctx, r.db)

	order = &domain.Order{}

	var status string
	err = querier.
		QueryRow(ctx, "SELECT id, user_id, description, amount, status FROM orders WHERE id = $1", id).
		Scan(&order.ID, &order.UserID, &order.Description, &order.Amount, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[OrderRepository] Order not found: %v", err)
			return nil, fmt.Errorf("order not found: %w", err)
		}
		log.Printf("[OrderRepository] Failed to get order: %v", err)
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	order.Status = domain.OrderStatus(status)
	log.Printf("[OrderRepository] Order found: %+v", order)
	return order, err
}

func (r *Repository) List(ctx context.Context, userID uuid.UUID) (orders []domain.Order, err error) {
	log.Printf("[OrderRepository] List called: userID=%s", userID)
	querier := txs.GetQuerier(ctx, r.db)

	rows, err := querier.Query(ctx, "SELECT id, user_id, description, amount, status FROM orders WHERE user_id = $1", userID)
	if err != nil {
		log.Printf("[OrderRepository] Failed to list orders: %v", err)
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		order := domain.Order{}
		var status string
		if err := rows.Scan(&order.ID, &order.UserID, &order.Description, &order.Amount, &status); err != nil {
			log.Printf("[OrderRepository] Failed to scan order: %v", err)
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		order.Status = domain.OrderStatus(status)
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		log.Printf("[OrderRepository] Rows error: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	log.Printf("[OrderRepository] Orders found: %d", len(orders))
	return orders, nil
}

func (r *Repository) Create(ctx context.Context, order *domain.Order) error {
	log.Printf("[OrderRepository] Create called: %+v", order)
	querier := txs.GetQuerier(ctx, r.db)

	_, err := querier.Exec(ctx,
		"INSERT INTO orders (id, user_id, description, amount, status) VALUES ($1, $2, $3, $4, $5)",
		order.ID, order.UserID, order.Description, order.Amount, order.Status,
	)
	if err != nil {
		log.Printf("[OrderRepository] Failed to create order: %v", err)
		return fmt.Errorf("failed to create order: %w", err)
	}
	log.Printf("[OrderRepository] Order created successfully: %s", order.ID)
	return nil
}

func (r *Repository) Update(ctx context.Context, order *domain.Order) error {
	log.Printf("[OrderRepository] Update called: %+v", order)
	querier := txs.GetQuerier(ctx, r.db)

	_, err := querier.Exec(ctx,
		"UPDATE orders SET user_id = $1, description = $2, amount = $3, status = $4 WHERE id = $5",
		order.UserID, order.Description, order.Amount, order.Status, order.ID,
	)
	if err != nil {
		log.Printf("[OrderRepository] Failed to update order: %v", err)
		return fmt.Errorf("failed to update order: %w", err)
	}
	log.Printf("[OrderRepository] Order updated successfully: %s", order.ID)
	return nil
}
