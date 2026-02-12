package repository

import (
	"context"
	"gophermart/internal/models"

	"github.com/jmoiron/sqlx"
)

type Authorization interface {
	CreateUser(user models.User) (int, error)
	GetUser(login, password string) (models.User, error)
}

// Интерфейс для взаимодействия БД c accrual
type OrderStore interface {
	// Функционал работы Заказов с db
	GetUnprocessedOrders(ctx context.Context) ([]string, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status string, accrual *float64) error
}

// type Orders interface {
// 	// Функционал работы Заказов с db
// }

type Loyalty interface {
	// Функционал работы Лояльности с db
}

type Repository struct {
	Authorization
	//Orders
	OrderStore
	Loyalty
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Authorization: NewAuthPostgres(db),
		// Orders: ,
		// LoLoyalty: ,
	}
}
