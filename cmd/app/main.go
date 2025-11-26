package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"L0_main/internal/config"
	"L0_main/internal/infrastructure/cache"
	"L0_main/internal/infrastructure/db"
	http2 "L0_main/internal/infrastructure/http"
	"L0_main/internal/infrastructure/kafka"
	"L0_main/internal/metrics"
	"L0_main/internal/usecase"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func initTracer() (func(), error) {
	ctx := context.Background()

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("otel-collector:4317"),
	)
	if err != nil {
		return nil, err
	}
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewSchemaless(
			semconv.ServiceName("l0-service"),
			semconv.ServiceVersion("1.0.0"),
		)),
	)
	otel.SetTracerProvider(tp)
	return func() {
		_ = tp.Shutdown(context.Background())
	}, nil
}

func main() {
	shutdown, err := initTracer()
	if err != nil {
		log.Fatalf("Ошибка запуска трассировки: %v", err)
	}
	defer shutdown()
	cfg := config.Load()
	dbConn, err := sql.Open("postgres", cfg.DB_DSN)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer func(dbConn *sql.DB) {
		err := dbConn.Close()
		if err != nil {
			log.Printf("failed to close db connection: %v", err)
		}
	}(dbConn)

	repo := db.NewPostgres(dbConn)
	maincache := cache.NewCache(cfg.CacheSizeMB)
	ttl := time.Duration(cfg.CacheTTLSeconds) * time.Second
	serv := usecase.NewSrvice(repo, maincache, ttl)

	if err := serv.WarmUpCache(); err != nil {
		log.Fatalf("Ошибка прогрева кэша: %v", err)
	}
	consumer, err := kafka.NewConsumer(
		cfg.KafkaBrokers,
		cfg.KafkaTopic,
		cfg.KafkaGroupID,
		cfg.ConsumerDelaySec,
		cfg.DlqTopic,
		cfg.RetryTopic,
	)
	if err != nil {
		log.Fatal(err)
	}
	dlqconsumer := kafka.NewDlqConsumer(
		cfg.KafkaBrokers,
		cfg.DlqTopic,
		cfg.KafkaGroupID,
		cfg.ConsumerDelaySec,
	)
	go func() {
		err := consumer.Start(*serv)
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		err := dlqconsumer.Start(*serv)
		if err != nil {
			log.Fatal(err)
		}
	}()
	retry, _ := kafka.NewRetryConsumer(
		cfg.KafkaBrokers,
		cfg.KafkaGroupID,
		"retry",
		cfg.KafkaTopic,
		cfg.ConsumerDelaySec,
	)

	go func() {
		err := retry.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()
	metrics.Register()
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/static/order_lookup.html")
	})
	r.HandleFunc("/orders/{id}", http2.GetOrderHandler(serv)).Methods("GET")
	r.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Handler: r,
		Addr:    cfg.HTTPAddr,
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Сервер запущен: http://localhost:8080/")
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	<-stop
	log.Printf("Остановка сервиса...")

	consumer.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Printf("Сервис остановлен.")
}
