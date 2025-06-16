package domain

import (
	"context"

	"github.com/google/uuid"
)

type OrderRepository interface {
	Get(ctx context.Context, id uuid.UUID) (order *Order, err error)
	List(ctx context.Context, userID uuid.UUID) (orders []Order, err error)
	Create(ctx context.Context, order *Order) (err error)
	Update(ctx context.Context, order *Order) (err error)
}
