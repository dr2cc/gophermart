package service

import (
	"gophermart"
	"gophermart/internal/repository"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

// Сервис аутентификации
type Authorization interface {
	// Функцонал:
	// Регистрация пользователей
	CreateUser(user gophermart.User) (int, error)
	// Генерация jwt токенов
	GenerateToken(username, password string) (string, error)
	// Валидация jwt токенов
	ParseToken(token string) (int, error)
}

// Сервис работы с заказами
type Orders interface {
}

// Сервис лояльности
type Loyalty interface {
}

// Здесь определены предметные области (доменные зоны).
// ❗Предметная область это круг задач (сферы реального мира) решаемых приложением.
// Получается тут три предметные области: аутентификация, работа со списками, работа с задачами.
type Service struct {
	// Сервис аутентификации, со своим функционалом.
	Authorization
	// Сервис работы с заказами, со своим функционалом.
	Orders
	// Сервис лояльности, со своим функционалом.
	Loyalty
}

// Вызываается из main
func NewService(repos *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(repos.Authorization),
		// Orders: ,
		// LoLoyalty: ,
	}
}
