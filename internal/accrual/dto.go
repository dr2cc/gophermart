package accrual

import (
	"errors"
	"gophermart/internal/repository"
)

var (
	ErrTooManyRequests    = errors.New("rate limit exceeded")
	ErrOrderNotRegistered = errors.New("order not registered")
)

type OrderResponse struct {
	Order   string                `json:"order"`
	Status  repository.TaskStatus `json:"status"`
	Accrual *float64              `json:"accrual,omitempty"` // Используем указатель, чтобы поймать null/отсутствие
}
