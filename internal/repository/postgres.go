package repository

import (
	"github.com/jmoiron/sqlx"
)

const (
	usersTable   = "users"
	balanceTable = "balance"
)

// Call from app
func NewPostgresDB(dsn string) (*sqlx.DB, error) {
	// sqlx.Connect вызывает Open, а затем сразу выполняет Ping для проверки соединения.
	db, err := sqlx.Connect("postgres", dsn)
	// db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// err = db.Ping()
	// if err != nil {
	// 	return nil, err
	// }

	return db, nil
}
