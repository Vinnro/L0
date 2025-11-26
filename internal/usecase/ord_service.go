package usecase

import (
	"encoding/json"
	"log"
	"time"

	"L0_main/internal/domain"

	"github.com/segmentio/kafka-go"
)

type ICache interface {
	Set(key string, value []byte, expire time.Duration) error
	Get(key string) ([]byte, error)
	Del(key string) bool
}

type Order interface {
	Insert(order *domain.Order) error
	Get(Uid string) (*domain.Order, error)
	GetAll() ([]domain.Order, error)
	InsertIntoDLQ(msg kafka.Message) error
}

type OrdService struct {
	rep   Order
	cache ICache
	ttl   time.Duration
}

func NewSrvice(rep Order, cache ICache, TTL time.Duration) *OrdService {
	return &OrdService{rep: rep, cache: cache, ttl: TTL}
}

func (s *OrdService) InsertOrd(order *domain.Order) error {
	err := s.rep.Insert(order)
	if err != nil {
		return err
	}
	Data, err := json.Marshal(&order)
	if err == nil {
		if err := s.cache.Set(order.OrderUID, Data, s.ttl); err != nil {
			log.Printf("Ошибка записи в кэш: %v", err)
		}
		log.Printf("Заказ uid: %s успешно добавлен в кэш", order.OrderUID)
	}
	return nil
}

func (s *OrdService) GetOrd(uid string) (*domain.Order, error) {
	data, err := s.cache.Get(uid)
	if err == nil {
		var cached domain.Order
		if jsonErr := json.Unmarshal(data, &cached); jsonErr == nil {
			log.Printf("Успешная выгрузка заказа uid: %s , из кэша", uid)
			return &cached, nil
		}
	}
	ord, err := s.rep.Get(uid)
	if err != nil {
		return nil, err
	}
	bytes, _ := json.Marshal(ord)
	err = s.cache.Set(uid, bytes, time.Minute)
	if err != nil {
		return nil, err
	}

	return ord, nil
}

func (s *OrdService) WarmUpCache() error {
	orders, err := s.rep.GetAll()
	if err != nil {
		return err
	}
	for _, ord := range orders {
		data, _ := json.Marshal(ord)
		if err := s.cache.Set(ord.OrderUID, data, s.ttl); err != nil {
			return err
		}
		log.Printf("%s", ord.OrderUID)
	}
	log.Printf("Кэш успешно загружен (%d записей)", len(orders))
	return nil
}
func (s *OrdService) InsertDLQ(msg kafka.Message) error {
	err := s.rep.InsertIntoDLQ(msg)
	if err != nil {
		log.Printf("Ошибка вставки поврежденного заказа в БД: %v", err)
	}
	log.Printf("Успешная вставка поврежденного заказа в базу DLQ")
	return nil
}
