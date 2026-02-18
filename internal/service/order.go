package service

import "gophermart/internal/repository"

// "Общение" с репозиторием, сервиса работы с заказами
type OrderService struct {
	repo repository.Order
}

func NewOrderService(repo repository.Order) *OrderService {
	return &OrderService{repo: repo}
}

// 4️⃣ Возвращаем клиенту response.
// - `200` — номер заказа уже был загружен этим пользователем;
// - `202` StatusAccepted — новый номер заказа принят в обработку;
// - `400` — неверный формат запроса;
// - `401` — пользователь не аутентифицирован;
// - `409` — номер заказа уже был загружен другим пользователем;
// - `422` StatusUnprocessableEntity  — неверный формат номера заказа;
// - `500` — внутренняя ошибка сервера.

func (s *OrderService) RecordOrder(n string) error {
	// Номер заказа может быть проверен на корректность ввода с помощью [алгоритма Луна]
	// Отдаем номер в db
	if n != "0" {
		return ErrOrderAlreadyExists
	}
	return s.repo.RecordOrder(n)
}
