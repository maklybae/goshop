package domain

import "github.com/google/uuid"

type OrderStatus string

const (
	StatusNew       OrderStatus = "new"
	StatusFinished  OrderStatus = "finished"
	StatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Description string
	Amount      int64
	Status      OrderStatus
}
