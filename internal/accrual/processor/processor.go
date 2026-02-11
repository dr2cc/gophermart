package processor

import (
	"context"
	"gophermart/internal/accrual"
	"gophermart/internal/accrual/dto"
	"log"
	"time"
)

// Интерфейс для взаимодействия с БД (реализуйте его в своем storage)
type OrderStore interface {
	GetUnprocessedOrders(ctx context.Context) ([]string, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status string, accrual *float64) error
}

func Run(ctx context.Context, store OrderStore, client *accrual.Client) {
	go func() {
		ticker := time.NewTicker(2 * time.Second) // Интервал проверки БД
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// TODO: логируем выход, чтобы в консоли было видно завершение
				return
			case <-ticker.C:
				processOrders(ctx, store, client)
			}
		}
	}()
}

func processOrders(ctx context.Context, store OrderStore, client *accrual.Client) {
	// 1. Берем заказы, которые еще не завершены (NEW, PROCESSING)
	orders, err := store.GetUnprocessedOrders(ctx)
	if err != nil {
		log.Printf("failed to fetch orders: %v", err)
		return
	}

	for _, orderNum := range orders {
		resp, retryAfter, err := client.GetAccrual(ctx, orderNum)

		if err != nil {
			if err == dto.ErrTooManyRequests {
				log.Printf("Rate limit hit, sleeping for %v", retryAfter)
				time.Sleep(retryAfter)
				return // Выходим из цикла обработки текущей пачки, чтобы подождать
			}
			if err == dto.ErrOrderNotRegistered {
				log.Printf("Order %s not registered in accrual system", orderNum)
				continue
			}
			log.Printf("Error fetching accrual for %s: %v", orderNum, err)
			continue
		}

		// 2. Если статус изменился на конечный, обновляем БД
		// ВАЖНО: Если статус PROCESSING, мы просто идем дальше и проверим его в следующей итерации
		if resp.Status == "PROCESSED" || resp.Status == "INVALID" {
			err := store.UpdateOrderStatus(ctx, resp.Order, resp.Status, resp.Accrual)
			if err != nil {
				log.Printf("Failed to update order %s: %v", resp.Order, err)
			}
		}
	}
}
