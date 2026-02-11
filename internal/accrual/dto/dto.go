package dto

import (
	"errors"
)

var (
	ErrTooManyRequests    = errors.New("rate limit exceeded")
	ErrOrderNotRegistered = errors.New("order not registered")
)

type OrderResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"` // Используем указатель, чтобы поймать null/отсутствие
}
