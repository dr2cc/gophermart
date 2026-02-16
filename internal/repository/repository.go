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

// Сервис работы с заказами
type Order interface {
	// ТЕОРИЯ! Запись нового заказа в таблицу orders
	RecordOrder(n string) error
}

// Функционал работы accrual с db
type OrderStore interface {
	GetUnprocessedOrders(ctx context.Context) ([]string, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status string, accrual *float64) error
}

type Loyalty interface {
	// Функционал работы Лояльности с db
}

type Repository struct {
	Authorization
	Order
	OrderStore
	Loyalty
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Authorization: NewAuthPostgres(db),
		Order:         NewOrderPostgres(db),
		OrderStore:    NewAccrualPostgres(db),
		// Loyalty: ,
	}
}
