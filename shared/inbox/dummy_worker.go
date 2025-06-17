package inbox

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

type DummyWorker struct {
	consumer sarama.ConsumerGroup
	repo     Repository
	topic    string
	groupID  string
}

type paymentEvent struct {
	OrderID uuid.UUID `json:"order_id"`
	UserID  uuid.UUID `json:"user_id"`
	Amount  int64     `json:"amount"`
}

func NewDummyWorker(consumer sarama.ConsumerGroup, repo Repository, topic, groupID string) *DummyWorker {
	return &DummyWorker{
		consumer: consumer,
		repo:     repo,
		topic:    topic,
		groupID:  groupID,
	}
}

func (w *DummyWorker) Start(ctx context.Context) error {
	log.Printf("[InboxDummyWorker] Starting dummy worker for topic=%s groupID=%s", w.topic, w.groupID)
	handler := &dummyHandler{repo: w.repo}
	for {
		if err := w.consumer.Consume(ctx, []string{w.topic}, handler); err != nil {
			log.Printf("[InboxDummyWorker] Consume error: %v", err)
			return err
		}
		if ctx.Err() != nil {
			log.Printf("[InboxDummyWorker] Context cancelled, stopping dummy worker")
			return ctx.Err()
		}
	}
}

type dummyHandler struct {
	repo Repository
}

func (h *dummyHandler) Setup(_ sarama.ConsumerGroupSession) error {
	log.Printf("[InboxDummyWorker] Setup session")
	return nil
}
func (h *dummyHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	log.Printf("[InboxDummyWorker] Cleanup session")
	return nil
}

func (h *dummyHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.Printf("[InboxDummyWorker] Start consuming partition=%d topic=%s", claim.Partition(), claim.Topic())
	for msg := range claim.Messages() {
		log.Printf("[InboxDummyWorker] Received message: topic=%s partition=%d offset=%d key=%s", msg.Topic, msg.Partition, msg.Offset, string(msg.Key))
		var event paymentEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("[InboxDummyWorker] Failed to unmarshal kafka message: %v", err)
			sess.MarkMessage(msg, "")
			continue
		}

		inboxMsg := &Message{
			ID:        event.OrderID,
			EventType: msg.Topic,
			Payload:   msg.Value,
		}
		err := h.repo.Add(sess.Context(), inboxMsg)
		if err != nil {
			if !isUniqueViolation(err) {
				log.Printf("[InboxDummyWorker] Failed to add message to inbox: %v", err)
			}
		}
		log.Printf("[InboxDummyWorker] Message added to inbox: id=%s", event.OrderID)
		sess.MarkMessage(msg, "")
	}
	log.Printf("[InboxDummyWorker] Finished consuming partition=%d topic=%s", claim.Partition(), claim.Topic())
	return nil
}

// isUniqueViolation определяет, что ошибка — это дубликат PK (Postgres)
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate key value") || strings.Contains(errStr, "unique constraint")
}
