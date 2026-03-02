package accrual

import (
	"context"
	"gophermart/internal/repository"
	"log/slog"
	"sync"
	"time"
)

// Интерфейс для взаимодействия с БД (он реализован в repository).
// ❗ В Go существует правило: принимай интерфейсы, возвращай структуры.
// Интерфейс AccrualStorage (ex. OrderStore) объявлен в пакете processor. Это говорит о том, что
// ❗ пакету processor для работы нужно «что-то», что умеет:
// 1. давать необработанные заказы (GetUnprocessedOrders),
// 2. обновлять их статусы (UpdateOrderStatus).
// Ему всё равно, как именно это делает любая база данных.

// AccrualStorage — интерфейс зависимостей
type AccrualStorage interface {
	GetUnprocessedOrders(ctx context.Context) ([]string, error)
	// UpdateOrderStatus(ctx context.Context, orderID string, status string, accrual *float64) error
	UpdateOrderStatus(ctx context.Context, orderID string, status repository.TaskStatus, accrual *float64) error
}

type Worker struct {
	repo     AccrualStorage
	client   *Client
	log      *slog.Logger
	interval time.Duration

	// Поля для подавления дублей логов
	mu              sync.Mutex
	reportedMissing map[string]bool
}

// Конструктор, собирающий воркер
// New принимает интервал как аргумент (например, из cfg.PollInterval)
func New(repo AccrualStorage, client *Client, interval int, log *slog.Logger) *Worker {
	return &Worker{
		repo:            repo,
		client:          client,
		log:             log,
		interval:        time.Duration(interval) * time.Second,
		reportedMissing: make(map[string]bool),
	}
}

// Run — метод структуры Worker. Он принимает только ctx из app.Run.
// Когда в консоли нажмут Ctrl+C, в процессоре сработает <-ctx.Done()
func (w *Worker) Run(ctx context.Context) {
	w.log.Info("Starting accrual worker")

	go func() {
		ticker := time.NewTicker(w.interval) // (2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				w.log.Info("Shutting down PROCESSOR")
				return
			case <-ticker.C:
				w.processOrders(ctx)
			}
		}
	}()
}

func (w *Worker) processOrders(ctx context.Context) {
	// Pattern operation name
	const op = "processor.processOrders"

	// Используем логгер из структуры, добавляя контекст операции
	logger := w.log.With(slog.String("op", op))

	// 1. Берем из db заказы со статусом- NEW и PROCESSING
	orders, err := w.repo.GetUnprocessedOrders(ctx)
	if err != nil {
		logger.Error("GetUnprocessedOrders error", "err", err)
		return
	}

	for _, orderNum := range orders {
		resp, retryAfter, err := w.client.GetAccrual(ctx, orderNum)

		if err != nil {
			if err == ErrTooManyRequests {
				// Rate limit логируем всегда (или редко),
				// так как это критично для всей системы, а не для заказа
				logger.Info("rate limit hit", "retry_after", retryAfter)
				time.Sleep(retryAfter)
				return // Выходим из цикла обработки текущей пачки, чтобы подождать
			}
			if err == ErrOrderNotRegistered {
				// Проверяем, логировали ли мы это уже
				if !w.reportedMissing[orderNum] {
					logger.Info("order not registered", "order_id", orderNum)
					w.reportedMissing[orderNum] = true
				}
				continue
			}
			// не регламентные ошибки
			logger.Error("abnormal error or accrual does not respond", "err", err)
			continue
		}
		// Если запрос прошел успешно — удаляем заказ из карты "проблемных",
		// чтобы если он снова пропадет (вдруг), мы снова получили лог.
		delete(w.reportedMissing, orderNum)

		// 2. Если статус изменился на конечный, обновляем БД
		// ВАЖНО: Если статус PROCESSING, мы просто идем дальше и проверим его в следующей итерации
		if resp.Status == "PROCESSED" || resp.Status == "INVALID" {
			err := w.repo.UpdateOrderStatus(ctx, resp.Order, repository.TaskStatus(resp.Status), resp.Accrual)
			if err != nil {
				logger.Error("failed to update order", "order", resp.Order, "err", err)
			}
		}
	}
}
