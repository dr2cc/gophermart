package service

import (
	"gophermart/internal/models"
	"gophermart/internal/repository"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

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
	RecordOrder(n string) error
	// // Теоретический функционал получения данных из accrual. Не знаю какому сервису нужен.
	// ReceivingCalculationLoyaltyPointsAccrual(accrualResponse dto.OrderResponse) error
}

// Сервис лояльности
type Loyalty interface {
	// // Теоретический функционал получения данных из accrual. Не знаю какому сервису нужен.
	// ReceivingCalculationLoyaltyPointsAccrual(accrualResponse dto.OrderResponse) error
}

// Здесь определены предметные области (доменные зоны).
// ❗Предметная область это круг задач (сферы реального мира) решаемых приложением.
// Получается тут три предметные области: аутентификация, работа со списками, работа с задачами.
type Service struct {
	// Сервис аутентификации, со своим функционалом.
	Authorization
	// Сервис работы с заказами, со своим функционалом.
	Order
	// Сервис лояльности, со своим функционалом.
	Loyalty
}

// Вызываается из main
func NewService(repos *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(repos.Authorization),
		Order:         NewOrderService(repos.Order),
		// LoLoyalty: ,
	}
}
