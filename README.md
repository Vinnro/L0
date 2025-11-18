# L0 — сервис обработки заказов (Kafka + PostgreSQL + Cache)

Этот проект реализует обработку заказов в реальном времени через Apache Kafka, сохранение данных в PostgreSQL, кэширование в памяти и HTTP-API для получения заказов. Архитектура приближена к Clean Architecture.

----------------------------------------

## Возможности

- Приём заказов из Kafka (topic: orders)
- Сохранение заказов в PostgreSQL
- Кэширование заказов:
  - прогрев кэша при старте (загрузка всех заказов из БД)
  - пополнение при получении новых сообщений из Kafka
- HTTP-API:
  - POST /orders/create — загрузка заказа через JSON-файл
  - GET /orders/{id} — получение заказа по order_uid
- Graceful shutdown
- Unit-тесты (gomock)

----------------------------------------

## Структура проекта

cmd/
  main.go               — точка входа

internal/
  config/               — конфигурация
  domain/               — модели + валидация
  usecase/              — бизнес-логика сервиса
  infrastructure/
    db/                 — PostgreSQL-репозиторий
    cache/              — кэш (freecache)
    http/               — HTTP-хендлеры
    kafka/              — Kafka-consumer

web/
  static/
    order_lookup.html   — HTML-форма для поиска заказа

----------------------------------------

## Запуск Kafka через Docker

docker-compose up -d

Проверка:
docker ps

----------------------------------------

## Запуск Go-сервиса

cd cmd
go run main.go

Сервер по умолчанию:
http://localhost:8080/

----------------------------------------

## HTTP API

### Загрузка заказа

POST /orders/create

curl -X POST -F "order_file=@order.json" http://localhost:8080/orders/create

### Получение заказа

GET /orders/{order_uid}

curl http://localhost:8080/orders/<uid>

----------------------------------------

## Отправка заказа через Kafka

Скопировать JSON в контейнер:

docker cp ./order.json l0-kafka-1:/tmp/order.json

Отправить:

kafka-console-producer --bootstrap-server localhost:9092 --topic orders < /tmp/order.json

Consumer получит заказ, провалидирует, сохранит в БД и положит в кэш.

----------------------------------------

## Тестирование

go test ./... -v

Покрыто:
- валидный Insert + Cache.Set
- невалидный Insert — репозиторий не вызывается
- cache hit — БД не вызывается
- cache miss — загрузка из БД + запись в кэш

----------------------------------------

## Конфигурация

DB_DSN=postgres://user:password@localhost:5432/l0?sslmode=disable
KAFKA_BROKER=localhost:29092
KAFKA_TOPIC=orders

----------------------------------------

## Требования

- Go 1.21+
- PostgreSQL
- Docker + Docker Compose
- Kafka

----------------------------------------

## Автор

Vinnro
GitHub: https://github.com/Vinnro

