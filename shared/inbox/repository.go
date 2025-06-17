package inbox

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maklybae/goshop/shared/txs"
)

type Repository interface {
	Add(ctx context.Context, message *Message) error
	GetUnprocessed(ctx context.Context, limit int) ([]*Message, error)
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

	_, err := querier.Exec(ctx, "INSERT INTO inbox (id, event_type, payload) VALUES ($1, $2, $3)",
		message.ID, message.EventType, message.Payload)
	if err != nil {
		return fmt.Errorf("failed to add message to inbox: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetUnprocessed(ctx context.Context, limit int) ([]*Message, error) {
	querier := txs.GetQuerier(ctx, r.db)

	rows, err := querier.Query(
		ctx,
		`SELECT id, event_type, payload
		 FROM inbox
		 WHERE processed = false
		 LIMIT $1
		 FOR UPDATE SKIP LOCKED`,
		limit,
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

	_, err := querier.Exec(ctx, "UPDATE inbox SET processed = true WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to mark message as processed: %w", err)
	}

	return nil
}
