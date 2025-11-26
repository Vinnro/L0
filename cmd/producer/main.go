package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	config2 "L0_main/internal/config"
	"L0_main/internal/domain"

	"github.com/Shopify/sarama"
)

func generateOrder() domain.Order {
	uid := time.Now().Format("20060102150405")

	return domain.Order{
		OrderUID:          uid,
		TrackNum:          "TRACK_" + uid,
		Entry:             "WEB",
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "customer_" + uid,
		DeliveryService:   "meest",
		ShardKey:          "1",
		SmID:              rand.Intn(100),
		DateCreated:       time.Now().Format(time.RFC3339),
		OofShard:          "0",

		Delivery: domain.Delivery{
			Name:    "Test Testov",
			Phone:   "+1000222333",
			Zip:     "999999",
			City:    "Moscow",
			Address: "Pushkina 1",
			Region:  "MSK",
			Email:   "test@example.com",
		},

		Payment: domain.Payment{
			Transaction:  uid,
			Currency:     "RUB",
			Provider:     "bank",
			Amount:       1500,
			PaymentDt:    time.Now().Unix(),
			Bank:         "sber",
			DeliveryCost: 200,
			GoodsTotal:   1300,
			CustomFee:    0,
		},

		Items: []domain.Item{
			{
				ChrtID:      rand.Intn(1000000),
				TrackNumber: "TRACK_" + uid,
				Price:       1300,
				Rid:         "RID123",
				Name:        "Phone",
				Sale:        0,
				Size:        "M",
				TotalPrice:  1300,
				NmID:        555555,
				Brand:       "Xiaomi",
				Status:      200,
			},
		},
	}
}

func main() {
	cfg := config2.Load()
	brokers := cfg.KafkaBrokers
	topic := cfg.KafkaTopic

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = cfg.RetryMax
	config.Version = sarama.MaxVersion

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Ошибка создания продюсера: %v", err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Printf("Ошибка закрытия продюссера: %v", err)
		}
	}()
	interval := time.Duration(cfg.ProducerIntervalMS) * time.Millisecond
	log.Println("Kafka producer успешно запущен")
	for {
		order := generateOrder()

		data, err := json.Marshal(order)
		if err != nil {
			log.Printf("Ошибка сериализации: %v", err)
			continue
		}

		msg := &sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(order.OrderUID),
			Value: sarama.ByteEncoder(data),
		}

		partition, offset, err := producer.SendMessage(msg)
		if err != nil {
			log.Printf("Ошибка отправки сообщения в Kafka: %v", err)
			continue
		}

		log.Printf("Отправлен заказ UID=%s | partition=%d offset=%d",
			order.OrderUID, partition, offset)

		time.Sleep(interval)
	}
}
