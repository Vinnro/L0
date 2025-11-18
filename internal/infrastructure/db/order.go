package db

import (
	"database/sql"
	"log"
)

func CreateTable(dbConn *sql.DB) error {
	query := `	
		CREATE TABLE orders (
		order_uid TEXT PRIMARY KEY,
		track_number TEXT NOT NULL,
		entry TEXT NOT NULL,
		locale TEXT NOT NULL,
		customer_id TEXT NOT NULL,
		delivery_service TEXT NOT NULL,
		shardkey TEXT NOT NULL,
		sm_id INT NOT NULL,
		date_created TIMESTAMPTZ NOT NULL,
		oof_shard TEXT NOT NULL
	)`
	query1 := `
		CREATE TABLE delivery (
    	order_uid TEXT PRIMARY KEY REFERENCES orders(order_uid),
    	name TEXT NOT NULL,
    	phone TEXT NOT NULL,
    	zip TEXT NOT NULL,
    	city TEXT NOT NULL,
    	address TEXT NOT NULL,
    	region TEXT NOT NULL,
    	email TEXT NOT NULL
	)`
	query2 := `
		CREATE TABLE payment (
    	order_uid TEXT PRIMARY KEY REFERENCES orders(order_uid),
    	transaction TEXT NOT NULL,
    	request_id TEXT NOT NULL,
    	currency TEXT NOT NULL,
    	provider TEXT NOT NULL,
    	amount INT NOT NULL,
    	payment_dt BIGINT NOT NULL,
    	bank TEXT NOT NULL,
    	delivery_cost INT NOT NULL,
    	goods_total INT NOT NULL,
    	custom_fee INT NOT NULL
	)`
	query3 := `
		CREATE TABLE items (
    	id SERIAL PRIMARY KEY,
    	order_uid TEXT REFERENCES orders(order_uid),
    	chrt_id INT NOT NULL,
    	track_number TEXT NOT NULL,
    	price INT NOT NULL,
    	rid TEXT NOT NULL,
    	name TEXT NOT NULL,
    	sale INT NOT NULL,
    	size TEXT NOT NULL,
    	total_price INT NOT NULL,
    	nm_id INT NOT NULL,
    	brand TEXT NOT NULL,
    	status INT NOT NULL
	)`
	_, err := dbConn.Exec(query)
	if err != nil {
		log.Printf("Ошибка при создании таблицы orders: %v", err)
		return err
	}

	_, err = dbConn.Exec(query1)
	if err != nil {
		log.Printf("Ошибка при создании таблицы delivery: %v", err)
		return err
	}

	_, err = dbConn.Exec(query2)
	if err != nil {
		log.Printf("Ошибка при создании таблицы payment: %v", err)
		return err
	}

	_, err = dbConn.Exec(query3)
	if err != nil {
		log.Printf("Ошибка при создании таблицы items: %v", err)
		return err
	}

	log.Println("Все таблицы успешно созданы или уже существовали")
	return nil
}
