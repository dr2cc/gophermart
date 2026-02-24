package processor

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
// Интерфейс OrderStore объявлен в пакете processor. Это говорит о том, что
// ❗ пакету processor для работы нужно «что-то», что умеет:
// 1. давать необработанные заказы (GetUnprocessedOrders),
// 2. обновлять их статусы (UpdateOrderStatus).
// Ему всё равно, как именно это делает любая база данных.
type OrderStore interface {
	GetUnprocessedOrders(ctx context.Context) ([]string, error)
	// UpdateOrderStatus(ctx context.Context, orderID string, status string, accrual *float64) error
	UpdateOrderStatus(ctx context.Context, orderID string, status repository.TaskStatus, accrual *float64) error
}

// Получили ctx из app.Run. Когда в консоли нажмут Ctrl+C, в процессоре сработает <-ctx.Done()
func Run(ctx context.Context, store OrderStore, client *accrual.Client, log *slog.Logger) {
	go func() {
		ticker := time.NewTicker(2 * time.Second) // Интервал проверки БД
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Info("Shutting down PROCESSOR")
				return
			case <-ticker.C:
				processOrders(ctx, store, client, log)
			}
		}
	}()
}

func processOrders(ctx context.Context, store OrderStore, client *accrual.Client, log *slog.Logger) {
	const op = "processor.processOrders"

	log = log.With(
		slog.String("op", op),
	)

	// 1. Берем заказы, которые еще не завершены (NEW, PROCESSING)
	orders, err := store.GetUnprocessedOrders(ctx)
	if err != nil {
		log.Error("failed to fetch orders:", "err", err)
		return
	}

	for _, orderNum := range orders {
		resp, retryAfter, err := client.GetAccrual(ctx, orderNum)

		if err != nil {
			if err == dto.ErrTooManyRequests {
				slog.Info("rate limit hit", "retry_after", retryAfter)
				time.Sleep(retryAfter)
				return // Выходим из цикла обработки текущей пачки, чтобы подождать
			}
			if err == dto.ErrOrderNotRegistered {
				slog.Info("order not registered in accrual system", "order_id", orderNum)
				continue
			}
			slog.Error("error fetching accrual", "order", orderNum, "err", err)
			continue
		}

		// 2. Если статус изменился на конечный, обновляем БД
		// ВАЖНО: Если статус PROCESSING, мы просто идем дальше и проверим его в следующей итерации
		if resp.Status == "PROCESSED" || resp.Status == "INVALID" {
			err := store.UpdateOrderStatus(ctx, resp.Order, resp.Status, resp.Accrual)
			if err != nil {
				log.Error("failed to update order", "order", resp.Order, "err", err)
			}
		}
	}
}
