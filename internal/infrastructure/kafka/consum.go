package kafka

import (
	"L0_main/internal/domain"
	"L0_main/internal/usecase"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	stop   chan struct{}
}

func NewConsumer(brokers []string, topic string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: "orders-group",
		}),
		stop: make(chan struct{}),
	}
}
func (c *Consumer) Start(service usecase.OrdService) error {
	time.Sleep(15 * time.Second)
	defer c.reader.Close()
	for {
		select {
		case <-c.stop:
			log.Println("consumer stopped")
			_ = c.reader.Close()
			return nil
		default:
			msg, err := c.reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("Ошибка чтения kafka: %v", err)
				continue
			}

			var order domain.Order
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				log.Printf("Ошибка чтения order: %v", err)
				continue
			}
			if err := service.InsertOrd(&order); err != nil {
				log.Printf("Ошибка сохранения в БД : %v", err)
				continue
			}
			log.Printf("Order processed: %s", order.OrderUID)
		}

	}
}

func (c *Consumer) Stop() {
	close(c.stop)
}
