package service

import (
	"gophermart/internal/models"
	"gophermart/internal/repository"
)

// Здесь в интерфейсах определены доменные зоны (предметные области).
// ❗Предметная область это круг задач (сферы реального мира) решаемых приложением.
// Получается тут три предметные области: аутентификация, работа с заказами, сервис лояльности.

//go:generate mockgen -source=$GOFILE -destination=mocks/mock.go

// Сервис аутентификации
type Authorization interface {
	// Функцонал:
	// Регистрация пользователей
	CreateUser(user models.User) (int, error)
	// Генерация jwt токенов
	GenerateToken(login, password string) (string, error)
	// Валидация jwt токенов
	ParseToken(token string) (int, error)
}

// Сервис работы с заказами
type Order interface {
	// Запись нового заказа в таблицу orders
	RecordOrder(id int, n string) error
	// // Теоретический функционал получения данных из accrual. Не знаю какому сервису нужен.
	// ReceivingCalculationLoyaltyPointsAccrual(accrualResponse dto.OrderResponse) error
}

// Сервис лояльности
type Loyalty interface {
	// // Теоретический функционал получения данных из accrual. Не знаю какому сервису нужен.
	// ReceivingCalculationLoyaltyPointsAccrual(accrualResponse dto.OrderResponse) error
}

type Service struct {
	// Интерфейс сервиса аутентификации.
	Authorization
	// Интерфейс сервиса работы с заказами.
	Order
	// Интерфейс сервиса лояльности.
	Loyalty
}

// Вызываается из main
func NewService(repo *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(repo.Authorization),
		Order:         NewOrderService(repo.Order),
		// LoLoyalty: ,
	}
}
