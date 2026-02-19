package service

import "gophermart/internal/repository"

// Service implementation struct.
// Отдадим, в конструкторе ниже, Структуру в которую там же
// Приняли Интерфейс репозитория (для "общения" с базой).
type orderService struct {
	repo repository.Order
}

func NewOrderService(repo repository.Order) *orderService {
	return &orderService{repo: repo}
}

// 4️⃣ Возвращаем клиенту response.
// - `200` — номер заказа уже был загружен этим пользователем;
// - `202` StatusAccepted — новый номер заказа принят в обработку;
// - `400` — неверный формат запроса;
// - `401` — пользователь не аутентифицирован;
// - `409` — номер заказа уже был загружен другим пользователем;
// - `422` StatusUnprocessableEntity  — неверный формат номера заказа;
// - `500` — внутренняя ошибка сервера.

func (s *orderService) RecordOrder(n string) error {
	// Номер заказа может быть проверен на корректность ввода с помощью [алгоритма Луна]
	// Отдаем номер в db
	if n != "0" {
		return ErrOrderAlreadyExists
	}
	return s.repo.RecordOrder(n)
}
