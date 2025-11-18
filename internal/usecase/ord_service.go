package usecase

import (
	"L0_main/internal/domain"
	"encoding/json"
	"log"
	"time"
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
}

type OrdService struct {
	rep   Order
	cache ICache
}

func NewSrvice(rep Order, cache ICache) *OrdService {
	return &OrdService{rep: rep, cache: cache}
}

func (s *OrdService) InsertOrd(order *domain.Order) error {
	if err := order.Validate(); err != nil {
		log.Printf("Данные не валидны")
		return err
	}
	err := s.rep.Insert(order)
	if err != nil {
		return err
	}

	Data, err := json.Marshal(&order)
	if err == nil {
		_ = s.cache.Set(order.OrderUID, Data, 0)
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
	s.cache.Set(uid, bytes, time.Minute)

	return ord, nil
}

func (s *OrdService) WarmUpCache() error {
	orders, err := s.rep.GetAll()
	if err != nil {
		return err
	}
	for _, ord := range orders {
		data, _ := json.Marshal(ord)
		s.cache.Set(ord.OrderUID, data, 60*time.Second)
		log.Printf("%s", ord.OrderUID)
	}
	log.Printf("Кэш успешно загружен (%d записей)", len(orders))
	return nil
}
