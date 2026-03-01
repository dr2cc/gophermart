package worker

import (
	"context"
	"gophermart/internal/accrual"
	"gophermart/internal/accrual/dto"
	"gophermart/internal/repository"
	"log/slog"
	"time"
)

// Интерфейс для взаимодействия с БД (он реализован в repository).
// ❗ В Go существует правило: принимай интерфейсы, возвращай структуры.
// Интерфейс AccrualStorage (ex. OrderStore) объявлен в пакете processor. Это говорит о том, что
// ❗ пакету processor для работы нужно «что-то», что умеет:
// 1. давать необработанные заказы (GetUnprocessedOrders),
// 2. обновлять их статусы (UpdateOrderStatus).
// Ему всё равно, как именно это делает любая база данных.
type AccrualStorage interface {
	GetUnprocessedOrders(ctx context.Context) ([]string, error)
	// UpdateOrderStatus(ctx context.Context, orderID string, status string, accrual *float64) error
	UpdateOrderStatus(ctx context.Context, orderID string, status repository.TaskStatus, accrual *float64) error
}

// Получили ctx из app.Run. Когда в консоли нажмут Ctrl+C, в процессоре сработает <-ctx.Done()
func Run(ctx context.Context, repo AccrualStorage, client *accrual.Client, log *slog.Logger) {
	go func() {
		ticker := time.NewTicker(2 * time.Second) // Интервал проверки БД
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Info("Shutting down PROCESSOR")
				return
			case <-ticker.C:
				processOrders(ctx, repo, client, log)
			}
		}
	}()
}

func processOrders(ctx context.Context, repo AccrualStorage, client *accrual.Client, log *slog.Logger) {
	// Pattern operation name
	const op = "processor.processOrders"

	log = log.With(
		slog.String("op", op),
	)

	// 1. Берем из db заказы со статусом- NEW и PROCESSING
	orders, err := repo.GetUnprocessedOrders(ctx)
	if err != nil {
		log.Error("GetUnprocessedOrders returned an error:", "err", err)
		return
	}

	for _, orderNum := range orders {
		resp, retryAfter, err := client.GetAccrual(ctx, orderNum)

		if err != nil {
			if err == dto.ErrTooManyRequests {
				log.Info("rate limit hit", "retry_after", retryAfter)
				time.Sleep(retryAfter)
				return // Выходим из цикла обработки текущей пачки, чтобы подождать
			}
			if err == dto.ErrOrderNotRegistered {
				log.Info("order not registered in accrual system", "order_id", orderNum)
				continue
			}
			// не регламентные ошибки
			log.Error("non-regulatory", "order_number", orderNum, "err", err)
			continue
		}

		// 2. Если статус изменился на конечный, обновляем БД
		// ВАЖНО: Если статус PROCESSING, мы просто идем дальше и проверим его в следующей итерации
		if resp.Status == "PROCESSED" || resp.Status == "INVALID" {
			err := repo.UpdateOrderStatus(ctx, resp.Order, resp.Status, resp.Accrual)
			if err != nil {
				log.Error("failed to update order", "order", resp.Order, "err", err)
			}
		}
	}
}
