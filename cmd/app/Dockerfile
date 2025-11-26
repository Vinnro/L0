FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/app

FROM debian:stable-slim

WORKDIR /app
RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

COPY --from=builder /app/app /app/app
COPY --from=builder /app/web ./web
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY ./migrations /app/migrations

CMD ["./app"]
