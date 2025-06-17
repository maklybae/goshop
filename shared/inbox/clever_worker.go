package inbox

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/maklybae/goshop/shared/txs"
)

type MessageHandler func(ctx context.Context, msg *Message) error

type CleverWorker struct {
	repo       Repository
	batchSize  int
	handler    MessageHandler
	transactor txs.Transactor
}

func NewCleverWorker(repo Repository, batchSize int, handler MessageHandler, transactor txs.Transactor) *CleverWorker {
	return &CleverWorker{
		repo:       repo,
		batchSize:  batchSize,
		handler:    handler,
		transactor: transactor,
	}
}

func (w *CleverWorker) Start(ctx context.Context, period time.Duration) {
	log.Printf("[InboxCleverWorker] Starting clever worker with batchSize=%d period=%v", w.batchSize, period)
	ticker := time.NewTicker(period)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Printf("[InboxCleverWorker] Context cancelled, stopping clever worker")
				return
			case <-ticker.C:
				log.Printf("[InboxCleverWorker] Ticker triggered, processing batch")
			}
			err := w.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
				log.Printf("[InboxCleverWorker] Fetching unprocessed messages")
				msgs, err := w.repo.GetUnprocessed(txCtx, w.batchSize)
				if err != nil {
					log.Printf("[InboxCleverWorker] Failed to get unprocessed messages: %v", err)
					return fmt.Errorf("failed to get unprocessed messages: %w", err)
				}
				if len(msgs) == 0 {
					log.Printf("[InboxCleverWorker] No messages to process")
					return nil
				}
				for _, msg := range msgs {
					log.Printf("[InboxCleverWorker] Handling message id=%s", msg.ID)
					if err := w.handler(txCtx, msg); err != nil {
						log.Printf("[InboxCleverWorker] Failed to handle message id=%s: %v", msg.ID, err)
						return fmt.Errorf("failed to handle message: %w", err)
					}
					if err := w.repo.MarkAsProcessed(txCtx, msg.ID); err != nil {
						log.Printf("[InboxCleverWorker] Failed to mark message as processed id=%s: %v", msg.ID, err)
						return fmt.Errorf("failed to mark message as processed: %w", err)
					}
					log.Printf("[InboxCleverWorker] Message processed and marked id=%s", msg.ID)
				}
				return nil
			})
			if err != nil {
				log.Printf("[InboxCleverWorker] Transaction error: %v", err)
			}
		}
	}()
}
