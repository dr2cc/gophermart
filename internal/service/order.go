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
	// Номер заказа может быть проверен на корректность ввода с помощью [алгоритма Луна]
	// Отдаем номер в db
	if n != "0" {
		return ErrOrderAlreadyExists
	}
	return nil
}
