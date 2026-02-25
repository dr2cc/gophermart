package repository

import (
	"errors"
	"fmt"
	"gophermart/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type OrderPostgres struct {
	db *sqlx.DB
}

func NewOrderPostgres(db *sqlx.DB) *OrderPostgres {
	return &OrderPostgres{db: db}
}

func (r *OrderPostgres) RecordOrder(userId int, n string) error {
	// 1. Начинаем транзакцию
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Откатит всё, если где-то возникнет ошибка

	var id int
	// 1. Создаем заказ
	orderQuery := fmt.Sprintf("INSERT INTO %s (order_number,user_id) VALUES ($1,$2) RETURNING id", ordersTable)
	if err := tx.QueryRow(orderQuery, n, userId).Scan(&id); err != nil {
		var pgErr *pq.Error
		// Проверяем, является ли ошибка нарушением уникальности
		// 23505 — код unique_violation
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.ErrOrderAlreadyExists
		}
		return err
	}

	// 2. Фиксируем изменения (явный коммит)
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
