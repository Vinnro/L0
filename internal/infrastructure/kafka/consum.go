package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"L0_main/internal/domain"
	"L0_main/internal/metrics"
	"L0_main/internal/usecase"

	"github.com/Shopify/sarama"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader    *kafka.Reader
	delay     time.Duration
	dlqtopic  string
	retrytoic string
	producer  sarama.SyncProducer
	stop      chan struct{}
}

func NewConsumer(brokers []string, topic string, groupID string, delay int, dlqtopic string, retrytoic string) (*Consumer, error) {

	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true

	baseProducer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})

	return &Consumer{
		reader:    reader,
		producer:  baseProducer,
		delay:     time.Duration(delay) * time.Second,
		dlqtopic:  dlqtopic,
		retrytoic: retrytoic,
		stop:      make(chan struct{}),
	}, nil
}

func CopyHeaders(src []kafka.Header) []sarama.RecordHeader {
	out := make([]sarama.RecordHeader, 0, len(src))
	for _, h := range src {
		out = append(out, sarama.RecordHeader{
			Key:   append([]byte(nil), h.Key...),
			Value: append([]byte(nil), h.Value...),
		})
	}
	return out
}

func AddOrUpdateRetry(msg *sarama.ProducerMessage, retry int) {
	rBytes := []byte(strconv.Itoa(retry))

	for i := range msg.Headers {
		if string(msg.Headers[i].Key) == "retry" {
			msg.Headers[i].Value = rBytes
			return
		}
	}

	msg.Headers = append(msg.Headers, sarama.RecordHeader{
		Key:   []byte("retry"),
		Value: rBytes,
	})
}

func GetR(msg kafka.Message) int {
	for _, h := range msg.Headers {
		if string(h.Key) == "retry" {
			n, err := strconv.Atoi(string(h.Value))
			if err == nil {
				return n
			}
		}
	}
	return 0
}

func (c *Consumer) Retry(msg kafka.Message) {
	retry := GetR(msg) + 1

	msgb := &sarama.ProducerMessage{
		Topic:   c.retrytoic,
		Key:     sarama.ByteEncoder(msg.Key),
		Value:   sarama.ByteEncoder(msg.Value),
		Headers: CopyHeaders(msg.Headers),
	}

	AddOrUpdateRetry(msgb, retry)

	_, _, err := c.producer.SendMessage(msgb)
	if err != nil {
		log.Printf("Ошибка отправки в retry: %v", err)
	}
}

func (c *Consumer) DLQ(msg kafka.Message, errMessage string, err error, ts time.Time) {

	msgb := &sarama.ProducerMessage{
		Topic:     c.dlqtopic,
		Key:       sarama.ByteEncoder(msg.Key),
		Value:     sarama.ByteEncoder(msg.Value),
		Timestamp: ts,
		Headers:   CopyHeaders(msg.Headers),
	}

	msgb.Headers = append(msgb.Headers,
		sarama.RecordHeader{
			Key:   []byte("error_message"),
			Value: []byte(errMessage),
		},
		sarama.RecordHeader{
			Key:   []byte("error_type"),
			Value: []byte(err.Error()),
		},
	)

	_, _, e := c.producer.SendMessage(msgb)
	if e != nil {
		log.Printf("Ошибка отправки в DLQ: %v", e)
	}
}

func (c *Consumer) Start(service usecase.OrdService) error {
	start := time.Now()
	time.Sleep(c.delay)
	defer func(reader *kafka.Reader) {
		err := reader.Close()
		if err != nil {
			log.Printf("Failed to close reader: %v", err)
		}
	}(c.reader)
	defer func(producer sarama.SyncProducer) {
		err := producer.Close()
		if err != nil {
			log.Printf("Failed to close producer: %v", err)
		}
	}(c.producer)
	for {
		select {
		case <-c.stop:
			log.Println("consumer stopped")
			return nil

		default:
			msg, err := c.reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("Ошибка чтения Kafka: %v", err)
				continue
			}

			retry := GetR(msg)

			var order domain.Order
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				metrics.OrdersErrors.Inc()
				metrics.DlqCount.Inc()
				log.Printf("Ошибка JSON: %v", err)
				if retry >= 3 {
					c.DLQ(msg, "Ошибка JSON", err, time.Now())
					continue
				}
				c.Retry(msg)
				continue
			}

			if err := order.Validate(); err != nil {
				metrics.OrdersErrors.Inc()
				metrics.DlqCount.Inc()
				log.Printf("Ошибка валидации: %v", err)
				if retry >= 3 {
					c.DLQ(msg, "Ошибка валидации", err, time.Now())
					continue
				}
				c.Retry(msg)
				continue
			}

			if err := service.InsertOrd(&order); err != nil {
				metrics.OrdersErrors.Inc()
				metrics.DlqCount.Inc()
				log.Printf("Ошибка БД: %v", err)
				if retry >= 3 {
					c.DLQ(msg, "Ошибка БД", err, time.Now())
					continue
				}
				c.Retry(msg)
				continue
			}
			metrics.OrdersProcessed.Inc()
			metrics.ProcessingTime.Observe(time.Since(start).Seconds())
			log.Printf("Order processed: %s", order.OrderUID)
		}
	}
}

func (c *Consumer) Stop() {
	close(c.stop)
}
