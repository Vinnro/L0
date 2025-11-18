package domain

import (
	"errors"
	"fmt"
)

func (o *Order) Validate() error {
	if o.OrderUID == "" {
		return errors.New("order_uid is required")
	}
	if o.TrackNum == "" {
		return errors.New("track_number is required")
	}
	if o.Entry == "" {
		return errors.New("entry is required")
	}
	if o.Delivery.Name == "" {
		return errors.New("delivery name is required")
	}
	if o.Delivery.Phone == "" {
		return errors.New("delivery phone is required")
	}
	if o.Delivery.Address == "" {
		return errors.New("delivery address is required")
	}
	if o.Payment.Transaction == "" {
		return errors.New("payment transaction is required")
	}
	if o.Payment.Amount < 0 {
		return errors.New("payment amount cannot be negative")
	}
	if o.Payment.GoodsTotal < 0 {
		return errors.New("payment goods_total cannot be negative")
	}
	if len(o.Items) == 0 {
		return errors.New("items cannot be empty")
	}
	for i, it := range o.Items {
		if it.ChrtID == 0 {
			return fmt.Errorf("items[%d].chrt_id is required", i)
		}
		if it.Price < 0 {
			return errors.New("item price cannot be negative")
		}
	}
	return nil
}
