package main

import (
	"L0_main/internal/config"
	"L0_main/internal/infrastructure/cache"
	"L0_main/internal/infrastructure/db"
	http2 "L0_main/internal/infrastructure/http"
	"L0_main/internal/infrastructure/kafka"
	"L0_main/internal/usecase"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.NewCon()
	dbConn, err := sql.Open("postgres", cfg.DBconfig)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer dbConn.Close()
	repo := db.NewPostgres(dbConn)
	maincache := cache.NewCache(120)
	serv := usecase.NewSrvice(repo, maincache)
	if err := serv.WarmUpCache(); err != nil {
		log.Fatalf("Ошибка прогрева кэша: %v", err)
	}
	consumer := kafka.NewConsumer(
		[]string{"localhost:29092"},
		"orders",
	)
	go consumer.Start(*serv)

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/static/order_lookup.html")
	})
	r.HandleFunc("/orders/create", http2.CreateOrderHandler(serv)).Methods("POST")
	r.HandleFunc("/orders/{id}", http2.GetOrderHandler(serv)).Methods("GET")

	srv := &http.Server{
		Handler: r,
		Addr:    ":8080",
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Сервер запущен на :http://localhost:8080/")
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	<-stop
	log.Printf("Остановка сервиса.....")

	consumer.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	log.Printf("Сервис остановлен...")
}
