package repository

import "github.com/jmoiron/sqlx"

type OrderPostgres struct {
	db *sqlx.DB
}

func NewOrderPostgres(db *sqlx.DB) *OrderPostgres {
	return &OrderPostgres{db: db}
}

func (r *OrderPostgres) RecordOrder(n string) error {

	return nil
}
