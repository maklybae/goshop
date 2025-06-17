package domain

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Account, error)
	ChangeBalance(ctx context.Context, userID uuid.UUID, amount int64) error
	List(ctx context.Context) ([]*Account, error)
	Create(ctx context.Context, account *Account) error
}
