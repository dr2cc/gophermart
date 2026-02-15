package repository

import (
	"errors"
	"fmt"
	"gophermart/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// const usersTable = "users"
// const balanceTable = "balance"

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuthPostgres(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{db: db}
}

func (r *AuthPostgres) CreateUser(user models.User) (int, error) {
	// 1. Начинаем транзакцию
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() // Откатит всё, если где-то возникнет ошибка

	var id int
	// 1. Создаем пользователя
	userQuery := fmt.Sprintf("INSERT INTO %s (login, hash) VALUES ($1, $2) RETURNING id", usersTable)
	if err := tx.QueryRow(userQuery, user.Login, user.Password).Scan(&id); err != nil {
		var pgErr *pq.Error
		// Проверяем, является ли ошибка нарушением уникальности
		// 23505 — код unique_violation
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, ErrUserAlreadyExists
		}
		return 0, err
	}

	// 2. Создаем пустой счет для этого пользователя
	// Используем константу balanceTable
	balanceQuery := fmt.Sprintf("INSERT INTO %s (user_id) VALUES ($1)", balanceTable)
	if _, err := tx.Exec(balanceQuery, id); err != nil {
		return 0, fmt.Errorf("failed to init balance: %w", err)
	}

	// 3. Фиксируем изменения (явный коммит)
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return id, nil
}

// func (r *AuthPostgres) CreateUser(user models.User) (int, error) {
// 	// 1. Начинаем транзакцию
// 	tx, err := r.db.Begin()
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer tx.Rollback() // Откатит всё, если где-то возникнет ошибка

// 	var id int
// 	// Вставляем пользователя
// 	userQuery := fmt.Sprintf("INSERT INTO %s (login, hash) VALUES ($1, $2) RETURNING id", usersTable)
// 	err = tx.QueryRow(userQuery, user.Login, user.Password).Scan(&id)

// 	if err != nil {
// 		var pgErr *pq.Error
// 		// Проверяем, является ли ошибка нарушением уникальности
// 		// 23505 — код unique_violation
// 		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
// 			return 0, ErrUserAlreadyExists
// 		}
// 		return 0, err
// 	}

// 	// 2. Инициализируем запись в таблице balance (баланс и списания по умолчанию 0)
// 	// balanceTable — замените на константу с именем вашей таблицы
// 	balanceQuery := "INSERT INTO balance (user_id) VALUES ($1)"
// 	if _, err := tx.Exec(balanceQuery, id); err != nil {
// 		return 0, fmt.Errorf("failed to initialize balance: %w", err)
// 	}

// 	// 3. Фиксируем изменения
// 	if err := tx.Commit(); err != nil {
// 		return 0, err
// 	}

// 	return id, nil
// }

func (r *AuthPostgres) GetUser(login, password string) (models.User, error) {
	var user models.User
	query := fmt.Sprintf("SELECT id FROM %s WHERE login=$1 AND hash=$2", usersTable)
	err := r.db.Get(&user, query, login, password)

	return user, err
}
