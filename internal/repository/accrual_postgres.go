package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type AccrualPostgres struct {
	db *sqlx.DB
}

func NewAccrualPostgres(db *sqlx.DB) *AccrualPostgres {
	return &AccrualPostgres{db: db}
}

// Столбцы (columns) таблицы orders:
// number , user_id , status , accrual , uploaded_at (TIMESTAMP), updated_at (TIMESTAMP), attempts
// Столбцы (columns) таблицы users:
// id , login , hash , created_at (TIMESTAMP)

// Метод GetUnprocessedOrders - «глаза» вашего воркера.
// Он лезет в БД и ищет заказы, по которым мы еще не получили финальный ответ от системы лояльности.
// Логика: Выбираем все заказы, у которых статус NEW (только что создан),
// REGISTERED (принят системой начислений) или PROCESSING (в процессе расчета).
// Результат: Список строк (номеров заказов), которые воркер должен «прогнать» через HTTP-запрос к accrual.
func (r *AccrualPostgres) GetUnprocessedOrders(ctx context.Context) ([]string, error) {
	var orders []string

	// Запрос выбирает номера заказов, которые еще не в финальном статусе
	query := `SELECT number FROM orders WHERE status NOT IN ('PROCESSED', 'INVALID')`

	err := r.db.SelectContext(ctx, &orders, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch unprocessed orders: %w", err)
	}

	return orders, nil
}

// Метод UpdateOrderStatus - «руки» вашего воркера. Он записывает результат в базу.
// Логика: Обновить статус заказа.
// Важно: Если пришел статус PROCESSED, нужно начислить баллы (accrual) на баланс пользователя.
// Это должно быть атомарно (в одной транзакции), чтобы не вышло так, что статус заказа изменился,
// а баллы пользователю «не долетели».
func (r *AccrualPostgres) UpdateOrderStatus(ctx context.Context, orderID string, status string, accrual *float64) error {
	// Начинаем транзакцию с контекстом
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	// Откат транзакции при ошибке (в sqlx Rollback не требует контекста)
	defer tx.Rollback()

	var val float64
	if accrual != nil {
		val = *accrual
	}

	// 1. Обновляем статус и начисление в таблице заказов
	_, err = tx.ExecContext(ctx,
		"UPDATE orders SET status = $1, accrual = $2 WHERE number = $3",
		status, val, orderID,
	)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	// 2. Если расчет окончен, начисляем баллы на баланс пользователя
	if status == "PROCESSED" && val > 0 {
		// ВАЖНО: используем таблицу balance и прибавляем к колонке balance
		_, err = tx.ExecContext(ctx,
			`UPDATE balance 
         SET balance = balance + $1 
         WHERE user_id = (SELECT user_id FROM orders WHERE number = $2)`,
			val, orderID,
		)
		if err != nil {
			return fmt.Errorf("failed to update user balance: %w", err)
		}
	}

	// Коммит транзакции
	return tx.Commit()
}
