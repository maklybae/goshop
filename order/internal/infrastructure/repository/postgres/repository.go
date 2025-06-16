package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maklybae/goshop/order/internal/domain"
	"github.com/maklybae/goshop/shared/txs"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (order *domain.Order, err error) {
	querier := txs.GetQuerier(ctx, r.db)

	order = &domain.Order{}

	var status string
	err = querier.
		QueryRow(ctx, "SELECT id, user_id, description, amount, status FROM orders WHERE id = $1", id).
		Scan(&order.ID, &order.UserID, &order.Description, &order.Amount, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found: %w", err)
		}

		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	order.Status = domain.OrderStatus(status)

	return order, err
}

func (r *Repository) List(ctx context.Context, userID uuid.UUID) (orders []domain.Order, err error) {
	querier := txs.GetQuerier(ctx, r.db)

	rows, err := querier.Query(ctx, "SELECT id, user_id, description, amount, status FROM orders WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		order := domain.Order{}

		var status string
		if err := rows.Scan(&order.ID, &order.UserID, &order.Description, &order.Amount, &status); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		order.Status = domain.OrderStatus(status)
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

func (r *Repository) Create(ctx context.Context, order *domain.Order) error {
	querier := txs.GetQuerier(ctx, r.db)

	_, err := querier.Exec(ctx,
		"INSERT INTO orders (id, user_id, description, amount, status) VALUES ($1, $2, $3, $4, $5)",
		order.ID, order.UserID, order.Description, order.Amount, order.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

func (r *Repository) Update(ctx context.Context, order *domain.Order) error {
	querier := txs.GetQuerier(ctx, r.db)

	_, err := querier.Exec(ctx,
		"UPDATE orders SET user_id = $1, description = $2, amount = $3, status = $4 WHERE id = $5",
		order.UserID, order.Description, order.Amount, order.Status, order.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}
