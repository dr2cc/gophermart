package repository

import (
	"errors"
	"fmt"
	"gophermart/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const usersTable = "users"

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuthPostgres(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{db: db}
}

func (r *AuthPostgres) CreateUser(user models.User) (int, error) {
	var id int
	query := fmt.Sprintf("INSERT INTO %s (login, hash) values ($1, $2) RETURNING id", usersTable)

	row := r.db.QueryRow(query, user.Login, user.Password)
	if err := row.Scan(&id); err != nil {
		// Проверяем, является ли ошибка нарушением уникальности
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // 23505 — код unique_violation
				return 0, ErrUserAlreadyExists
			}
		}
		return 0, err
	}

	return id, nil
}

func (r *AuthPostgres) GetUser(login, password string) (models.User, error) {
	var user models.User
	query := fmt.Sprintf("SELECT id FROM %s WHERE login=$1 AND hash=$2", usersTable)
	err := r.db.Get(&user, query, login, password)

	return user, err
}
