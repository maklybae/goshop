package outbox

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maklybae/goshop/shared/txs"
)

type Repository interface {
	Add(ctx context.Context, message *Message) error
	SetReserved(ctx context.Context, id uuid.UUID, reservedTo time.Time) error
	GetUnprocessed(ctx context.Context, limit int, reservedExpired time.Time) ([]*Message, error)
	MarkAsProcessed(ctx context.Context, id uuid.UUID) error
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Add(ctx context.Context, message *Message) error {
	querier := txs.GetQuerier(ctx, r.db)

	_, err := querier.Exec(ctx, "INSERT INTO outbox (id, event_type, payload) VALUES ($1, $2, $3)",
		message.ID, message.EventType, message.Payload)
	if err != nil {
		return fmt.Errorf("failed to add message to outbox: %w", err)
	}

	return nil
}

func (r *PostgresRepository) SetReserved(ctx context.Context, id uuid.UUID, reservedTo time.Time) error {
	querier := txs.GetQuerier(ctx, r.db)

	_, err := querier.Exec(ctx, "UPDATE outbox SET reserved_to = $1 WHERE id = $2", reservedTo, id)
	if err != nil {
		return fmt.Errorf("failed to set reserved time for message: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetUnprocessed(ctx context.Context, limit int, reservedExpired time.Time) ([]*Message, error) {
	querier := txs.GetQuerier(ctx, r.db)

	rows, err := querier.Query(
		ctx,
		`SELECT id, event_type, payload
		 FROM outbox
		 WHERE processed = false
		   AND (reserved_to IS NULL OR reserved_to < $1)
		 LIMIT $2
		 FOR UPDATE SKIP LOCKED`,
		reservedExpired, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get unprocessed messages: %w", err)
	}
	defer rows.Close()

	var messages []*Message

	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.EventType, &msg.Payload); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		messages = append(messages, &msg)
	}

	return messages, nil
}

func (r *PostgresRepository) MarkAsProcessed(ctx context.Context, id uuid.UUID) error {
	querier := txs.GetQuerier(ctx, r.db)

	_, err := querier.Exec(ctx, "UPDATE outbox SET processed = true WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to mark message as processed: %w", err)
	}

	return nil
}
