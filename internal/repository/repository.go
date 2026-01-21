package repository

import (
	"gophermart"

	"github.com/jmoiron/sqlx"
)

type Authorization interface {
	CreateUser(user gophermart.User) (int, error)
	GetUser(username, password string) (gophermart.User, error)
}

type Orders interface {
	// Функционал работы Заказов с db
}

type Loyalty interface {
	// Функционал работы Лояльности с db
}

type Repository struct {
	Authorization
	Orders
	Loyalty
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Authorization: NewAuthPostgres(db),
		// Orders: ,
		// LoLoyalty: ,
	}
}
