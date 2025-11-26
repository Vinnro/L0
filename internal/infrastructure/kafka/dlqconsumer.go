package kafka

import (
	"context"
	"log"
	"time"

	"L0_main/internal/metrics"
	"L0_main/internal/usecase"

	"github.com/segmentio/kafka-go"
)

type DlqConsumer struct {
	reader *kafka.Reader
	delay  time.Duration
	stop   chan struct{}
}

func NewDlqConsumer(brokers []string, topic string, groupid string, delay int) *DlqConsumer {
	return &DlqConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupid,
		}),
		delay: time.Duration(delay) * time.Second,
		stop:  make(chan struct{}),
	}
}

func (c *DlqConsumer) Start(service usecase.OrdService) error {
	time.Sleep(c.delay)
	defer func() {
		if err := c.reader.Close(); err != nil {
			log.Printf("Ошибка закрытия reader DLQ: %v", err)
		}
	}()
	for {
		select {
		case <-c.stop:
			log.Println("Остановка DLQ")
			if err := c.reader.Close(); err != nil {
				return err
			}
			return nil
		default:
			msg, err := c.reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("Ошибка DLQ: %v", err)
				continue
			}
			if err = service.InsertDLQ(msg); err != nil {
				log.Printf("Ошибка вставки в базу DLQ: %v", err)
				continue
			}
			metrics.DlqCount.Inc()
		}
	}
}
func (c *DlqConsumer) Stop() {
	close(c.stop)
}
