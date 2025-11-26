-- +goose Up

CREATE TABLE IF NOT EXISTS orders (
                                      order_uid        TEXT PRIMARY KEY,
                                      track_number     TEXT NOT NULL,
                                      entry            TEXT NOT NULL,
                                      locale           TEXT NOT NULL,
                                      customer_id      TEXT NOT NULL,
                                      delivery_service TEXT NOT NULL,
                                      shardkey         TEXT NOT NULL,
                                      sm_id            INTEGER NOT NULL,
                                      date_created     TEXT NOT NULL,
                                      oof_shard        TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS delivery (
                                        order_uid TEXT PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
                                        name      TEXT NOT NULL,
                                        phone     TEXT NOT NULL,
                                        zip       TEXT NOT NULL,
                                        city      TEXT NOT NULL,
                                        address   TEXT NOT NULL,
                                        region    TEXT NOT NULL,
                                        email     TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS payment (
                                       order_uid     TEXT PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
                                       transaction   TEXT NOT NULL,
                                       request_id    TEXT,
                                       currency      TEXT NOT NULL,
                                       provider      TEXT NOT NULL,
                                       amount        INTEGER NOT NULL,
                                       payment_dt    BIGINT NOT NULL,
                                       bank          TEXT NOT NULL,
                                       delivery_cost INTEGER NOT NULL,
                                       goods_total   INTEGER NOT NULL,
                                       custom_fee    INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS items (
                                     id           SERIAL PRIMARY KEY,
                                     order_uid    TEXT NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
                                     chrt_id      INTEGER NOT NULL,
                                     track_number TEXT NOT NULL,
                                     price        INTEGER NOT NULL,
                                     rid          TEXT NOT NULL,
                                     name         TEXT NOT NULL,
                                     sale         INTEGER NOT NULL,
                                     size         TEXT NOT NULL,
                                     total_price  INTEGER NOT NULL,
                                     nm_id        INTEGER NOT NULL,
                                     brand        TEXT NOT NULL,
                                     status       INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS dlq (
                                    id BIGSERIAL PRIMARY KEY ,
                                    topic VARCHAR,
                                    key BYTEA,
                                    value BYTEA,
                                    error_type VARCHAR,
                                    error_message TEXT,
                                    created_at TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS payment;
DROP TABLE IF EXISTS delivery;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS dlq;
