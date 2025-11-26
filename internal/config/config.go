package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DB_DSN             string
	KafkaBrokers       []string
	KafkaTopic         string
	KafkaGroupID       string
	CacheSizeMB        int
	CacheTTLSeconds    int
	HTTPAddr           string
	ConsumerDelaySec   int
	ProducerIntervalMS int
	RetryMax           int
	ProdDelay          int
	DlqTopic           string
	RetryTopic         string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		DB_DSN:             env("DB_DSN", ""),
		KafkaBrokers:       strings.Split(env("KAFKA_BROKERS", "kafka:9092"), ","),
		KafkaTopic:         env("KAFKA_TOPIC", "orders"),
		KafkaGroupID:       env("KAFKA_GROUP_ID", "orders-group"),
		CacheSizeMB:        envInt("CACHE_MB", 128),
		CacheTTLSeconds:    envInt("CACHE_TTL", 60),
		HTTPAddr:           env("HTTP_ADDR", ":8080"),
		ConsumerDelaySec:   envInt("CONSUMER_DELAY", 15),
		ProducerIntervalMS: envInt("PRODUCER_INTERVAL_MS", 5000),
		RetryMax:           envInt("RETRY_MAX", 10),
		DlqTopic:           env("DLQ_TOPIC", "orders_dlq"),
		RetryTopic:         env("RETRY_TOPIC", "retry"),
	}
}

func env(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func envInt(key string, def int) int {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return n
}
