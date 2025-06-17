package outbox

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"

	"github.com/maklybae/goshop/shared/txs"
)

type EventSender interface {
	SendMessages(msgs []*sarama.ProducerMessage) error
}

type Worker struct {
	sender     EventSender
	transactor txs.Transactor
	repo       Repository
	config     *Config
}

func NewWorker(sender EventSender, transactor txs.Transactor, repo Repository, config *Config) *Worker {
	return &Worker{
		sender:     sender,
		transactor: transactor,
		repo:       repo,
		config:     config,
	}
}

func (w *Worker) StartProcessingEvents(ctx context.Context) {
	log.Printf("[OutboxWorker] Starting outbox worker with period=%v batchSize=%d", w.config.Period, w.config.BatchSize)
	ticker := time.NewTicker(w.config.Period)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Printf("[OutboxWorker] Context cancelled, stopping outbox worker")
				return
			case <-ticker.C:
				log.Printf("[OutboxWorker] Ticker triggered, processing events")
			}

			w.processEvents(ctx)
		}
	}()
}

func (w *Worker) processEvents(ctx context.Context) {
	log.Printf("[OutboxWorker] processEvents called")
	var events []*Message

	var err error
	err = w.transactor.WithTransaction(ctx, func(ctx context.Context) error {
		log.Printf("[OutboxWorker] Fetching unprocessed events")
		events, err = w.repo.GetUnprocessed(ctx, w.config.BatchSize, time.Now())
		if err != nil {
			log.Printf("[OutboxWorker] Failed to get unprocessed events: %v", err)
			return fmt.Errorf("failed to get unprocessed events: %w", err)
		}

		for _, event := range events {
			log.Printf("[OutboxWorker] Reserving event id=%s", event.ID)
			err = w.repo.SetReserved(ctx, event.ID, time.Now().Add(w.config.Period))
			if err != nil {
				log.Printf("[OutboxWorker] Failed to set reserved for event id=%s: %v", event.ID, err)
				return fmt.Errorf("failed to set event reserved: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("[OutboxWorker] Transaction error: %v", err)
		return
	}

	if len(events) == 0 {
		log.Printf("[OutboxWorker] No events to process")
		return
	}

	msgs := make([]*sarama.ProducerMessage, 0, len(events))

	for _, event := range events {
		log.Printf("[OutboxWorker] Preparing message for event id=%s topic=%s", event.ID, event.EventType)
		msg := &sarama.ProducerMessage{
			Topic: event.EventType,
			Key:   sarama.StringEncoder(event.ID.String()),
			Value: sarama.ByteEncoder(event.Payload),
		}

		msgs = append(msgs, msg)
	}

	err = w.sender.SendMessages(msgs)
	if err != nil {
		log.Printf("[OutboxWorker] Failed to send messages: %v", err)
		return
	}
	log.Printf("[OutboxWorker] Messages sent successfully")

	err = w.transactor.WithTransaction(ctx, func(ctx context.Context) error {
		for _, event := range events {
			log.Printf("[OutboxWorker] Marking event as processed id=%s", event.ID)
			err = w.repo.MarkAsProcessed(ctx, event.ID)
			if err != nil {
				log.Printf("[OutboxWorker] Failed to mark event as processed id=%s: %v", event.ID, err)
				return fmt.Errorf("failed to mark event as processed: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		log.Printf("[OutboxWorker] Error marking events as processed: %v", err)
	}
}
