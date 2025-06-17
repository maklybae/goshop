package consumer

import (
	"context"
	"log"

	"github.com/IBM/sarama"
)

type MessageHandler func(ctx context.Context, msg *sarama.ConsumerMessage) error

type KafkaGroupConsumer struct {
	group   sarama.ConsumerGroup
	topic   string
	handler MessageHandler
}

func NewKafkaGroupConsumer(brokers []string, groupID, topic string, handler MessageHandler) (*KafkaGroupConsumer, error) {
	log.Printf("[KafkaGroupConsumer] Initializing consumer group '%s' for topic '%s' on brokers: %v", groupID, topic, brokers)
	cfg := sarama.NewConfig()
	cfg.Consumer.Return.Errors = true
	cfg.Version = sarama.V2_1_0_0

	group, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		log.Printf("[KafkaGroupConsumer] Failed to create consumer group: %v", err)
		return nil, err
	}
	log.Printf("[KafkaGroupConsumer] Consumer group created successfully")
	return &KafkaGroupConsumer{
		group:   group,
		topic:   topic,
		handler: handler,
	}, nil
}

type groupHandler struct {
	handler MessageHandler
}

func (h *groupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	log.Printf("[KafkaGroupConsumer] Setup session")
	return nil
}
func (h *groupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	log.Printf("[KafkaGroupConsumer] Cleanup session")
	return nil
}
func (h *groupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := sess.Context()
	log.Printf("[KafkaGroupConsumer] Start consuming partition=%d topic=%s", claim.Partition(), claim.Topic())
	for msg := range claim.Messages() {
		log.Printf("[KafkaGroupConsumer] Received message: topic=%s partition=%d offset=%d key=%s", msg.Topic, msg.Partition, msg.Offset, string(msg.Key))
		if err := h.handler(ctx, msg); err != nil {
			log.Printf("[KafkaGroupConsumer] failed to handle message: %v", err)
		}
		sess.MarkMessage(msg, "")
	}
	log.Printf("[KafkaGroupConsumer] Finished consuming partition=%d topic=%s", claim.Partition(), claim.Topic())
	return nil
}

func (k *KafkaGroupConsumer) Start(ctx context.Context) error {
	h := &groupHandler{handler: k.handler}
	log.Printf("[KafkaGroupConsumer] Starting consumer group for topic=%s", k.topic)
	for {
		if err := k.group.Consume(ctx, []string{k.topic}, h); err != nil {
			log.Printf("[KafkaGroupConsumer] Consume error: %v", err)
			return err
		}
		if ctx.Err() != nil {
			log.Printf("[KafkaGroupConsumer] Context cancelled, stopping consumer group")
			return ctx.Err()
		}
	}
}

func (k *KafkaGroupConsumer) Close() error {
	log.Printf("[KafkaGroupConsumer] Closing consumer group")
	return k.group.Close()
}
