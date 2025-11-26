package kafka

import (
	"context"
	"log"
	"strconv"
	"time"

	"L0_main/internal/metrics"

	"github.com/Shopify/sarama"
	"github.com/segmentio/kafka-go"
)

type RetryConsumer struct {
	reader     *kafka.Reader
	stop       chan struct{}
	delay      time.Duration
	producer   sarama.SyncProducer
	kafkaTopic string
}

func NewRetryConsumer(brokers []string, GroupID string, topic string, kafkaTopic string, delay int) (*RetryConsumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.MaxVersion
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Ошибка продюссера retry: %v", err)
	}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: GroupID,
		Topic:   topic,
	})
	return &RetryConsumer{
		reader:     reader,
		delay:      time.Duration(delay),
		producer:   producer,
		stop:       make(chan struct{}),
		kafkaTopic: kafkaTopic,
	}, nil
}

func GetRetry(msg kafka.Message) int {
	for _, h := range msg.Headers {
		if h.Key == "retry" {
			n, err := strconv.Atoi(string(h.Value))
			if err == nil {
				return n
			}
		}
	}
	return 0
}

func (r *RetryConsumer) Start() error {
	defer func(reader *kafka.Reader) {
		err := reader.Close()
		if err != nil {
			log.Printf("failed to close reader: %v", err)
		}
	}(r.reader)

	for {
		select {
		case <-r.stop:
			log.Println("Retry stopped")
			return nil

		default:
			msg, err := r.reader.ReadMessage(context.Background())
			if err != nil {
				log.Println(err)
				continue
			}

			oldRetry := GetRetry(msg)
			newRetry := oldRetry + 1

			msgr := &sarama.ProducerMessage{
				Topic:   r.kafkaTopic,
				Value:   sarama.ByteEncoder(msg.Value),
				Key:     sarama.ByteEncoder(msg.Key),
				Headers: CopyHeaders(msg.Headers),
			}

			AddOrUpdateRetry(msgr, newRetry)

			_, _, err = r.producer.SendMessage(msgr)
			if err != nil {
				log.Println("retry→orders ошибка:", err)
				continue
			}
			metrics.RetryCount.Inc()
		}
	}
}

func (r *RetryConsumer) Stop() {
	close(r.stop)
}
