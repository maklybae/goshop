package producer

import (
	"log"

	"github.com/IBM/sarama"
)

type KafkaSyncProducer struct {
	producer sarama.SyncProducer
}

func NewKafkaSyncProducer(brokers []string) (*KafkaSyncProducer, error) {
	log.Printf("[KafkaSyncProducer] Initializing producer for brokers: %v", brokers)
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 10

	p, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		log.Printf("[KafkaSyncProducer] Failed to create producer: %v", err)
		return nil, err
	}
	log.Printf("[KafkaSyncProducer] Producer created successfully")
	return &KafkaSyncProducer{producer: p}, nil
}

func (k *KafkaSyncProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	for _, msg := range msgs {
		log.Printf("[KafkaSyncProducer] Sending message to topic=%s key=%v", msg.Topic, msg.Key)
		_, _, err := k.producer.SendMessage(msg)
		if err != nil {
			log.Printf("[KafkaSyncProducer] Failed to send message: %v", err)
			return err
		}
	}
	log.Printf("[KafkaSyncProducer] All messages sent successfully")
	return nil
}

func (k *KafkaSyncProducer) Close() error {
	log.Printf("[KafkaSyncProducer] Closing producer")
	return k.producer.Close()
}
