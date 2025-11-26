package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"L0_main/internal/domain"

	"github.com/segmentio/kafka-go"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *Postgres {
	return &Postgres{db: db}
}

func (p *Postgres) Insert(order *domain.Order) error {
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.Exec(`
        INSERT INTO orders (
            order_uid, track_number, entry, locale, 
            customer_id, delivery_service, shardkey, 
            sm_id, date_created, oof_shard
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		order.OrderUID,
		order.TrackNum,
		order.Entry,
		order.Locale,
		order.CustomerID,
		order.DeliveryService,
		order.ShardKey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	)
	if err != nil {
		return fmt.Errorf("insert into orders: %w", err)
	}

	_, err = tx.Exec(`
        INSERT INTO delivery (
            order_uid, name, phone, zip, city, 
            address, region, email
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		order.OrderUID,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("insert into delivery: %w", err)
	}

	_, err = tx.Exec(`
        INSERT INTO payment (
            order_uid, transaction, request_id, currency, 
            provider, amount, payment_dt, bank, 
            delivery_cost, goods_total, custom_fee
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		order.OrderUID,
		order.Payment.Transaction,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDt,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("insert into payment: %w", err)
	}

	for _, item := range order.Items {
		_, err = tx.Exec(`
            INSERT INTO items (
                order_uid, chrt_id, track_number, price, 
                rid, name, sale, size, total_price, 
                nm_id, brand, status
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			order.OrderUID,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.Rid,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmID,
			item.Brand,
			item.Status,
		)
		if err != nil {
			return fmt.Errorf("insert into items: %w", err)
		}

	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}
	return nil
}

func (p *Postgres) Get(UID string) (*domain.Order, error) {
	var order domain.Order
	log.Printf("Поиск по order: %s", UID)

	err := p.db.QueryRow(`
        SELECT 
            order_uid, track_number, entry, locale,
            customer_id, delivery_service, shardkey,
            sm_id, date_created, oof_shard
        FROM orders 
        WHERE order_uid = $1`,
		UID,
	).Scan(
		&order.OrderUID,
		&order.TrackNum,
		&order.Entry,
		&order.Locale,
		&order.CustomerID,
		&order.DeliveryService,
		&order.ShardKey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Printf("failed to get order: %v\n", err)
			return nil, err
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	var delivery domain.Delivery
	err = p.db.QueryRow(`
        SELECT 
            name, phone, zip, city, 
            address, region, email
        FROM delivery 
        WHERE order_uid = $1`,
		UID,
	).Scan(
		&delivery.Name,
		&delivery.Phone,
		&delivery.Zip,
		&delivery.City,
		&delivery.Address,
		&delivery.Region,
		&delivery.Email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Printf("failed to get delivery: %v\n", err)
		} else {
			return nil, fmt.Errorf("failed to get delivery: %w", err)
		}
	}
	order.Delivery = delivery

	var payment domain.Payment
	err = p.db.QueryRow(`
        SELECT 
            transaction, request_id, currency, provider,
            amount, payment_dt, bank, delivery_cost,
            goods_total, custom_fee
        FROM payment 
        WHERE order_uid = $1`,
		UID,
	).Scan(
		&payment.Transaction,
		&payment.RequestID,
		&payment.Currency,
		&payment.Provider,
		&payment.Amount,
		&payment.PaymentDt,
		&payment.Bank,
		&payment.DeliveryCost,
		&payment.GoodsTotal,
		&payment.CustomFee,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Printf("failed to get payment: %v\n", err)
		} else {
			return nil, fmt.Errorf("failed to get payment: %w", err)
		}
	}
	order.Payment = payment

	rows, err := p.db.Query(`
        SELECT 
            chrt_id, track_number, price, rid,
            name, sale, size, total_price,
            nm_id, brand, status
        FROM items 
        WHERE order_uid = $1`,
		UID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}(rows)

	var items []domain.Item
	for rows.Next() {
		var item domain.Item
		err := rows.Scan(
			&item.ChrtID,
			&item.TrackNumber,
			&item.Price,
			&item.Rid,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}
	order.Items = items

	return &order, nil

}

func (p *Postgres) GetAll() ([]domain.Order, error) {
	rows, err := p.db.Query(`SELECT order_uid FROM orders`)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}(rows)

	var result []domain.Order

	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}

		ord, err := p.Get(uid)
		if err != nil {
			return nil, err
		}

		result = append(result, *ord)
	}

	return result, nil
}
func Header(msg kafka.Message, key string) string {
	for _, h := range msg.Headers {
		if string(h.Key) == key {
			return string(h.Value)
		}
	}
	return ""
}

func (p *Postgres) InsertIntoDLQ(msg kafka.Message) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Printf("failed to rollback transaction: %v\n", err)
		}
	}()
	_, err = tx.Exec(`
	INSERT INTO dlq(
	                topic, key, value,
	                error_type, error_message,created_at)VALUES ($1, $2, $3, $4, $5, $6)`,
		msg.Topic,
		msg.Key,
		msg.Value,
		Header(msg, "error_type"),
		Header(msg, "error_message"),
		msg.Time,
	)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
