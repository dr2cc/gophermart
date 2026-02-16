package service

import "gophermart/internal/repository"

// "Общение" с репозиторием, сервиса работы с заказами
type OrderService struct {
	repo repository.Order
}

func NewOrderService(repo repository.Order) *OrderService {
	return &OrderService{repo: repo}
}

func (o *OrderService) RecordOrder(n string) error {

	return nil
}
