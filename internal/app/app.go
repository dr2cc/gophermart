// Package app configures and runs application.
package app

import (
	"context"
	"errors"
	"fmt"
	"gophermart/db/migrations" // импорт вашего пакета с FS
	"gophermart/internal/accrual"
	worker "gophermart/internal/accrual/processor"
	"gophermart/internal/config"
	"gophermart/internal/handler"
	"gophermart/internal/repository"
	"gophermart/internal/server"
	"gophermart/internal/service"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

// "Точка сборки" (композитор) приложения.
func Run(cfg *config.Config) error {
	// 1. Создаем logger
	log := setupLogger(cfg.Env)
	log.Info("Init server", slog.String("address", cfg.ServerAddress))

	// 2. Инициализация подключения к БД
	// Используем кастомный конструктор для настройки пула соединений Postgres
	db, err := repository.NewPostgresDB(cfg.DatabaseDSN)
	if err != nil {
		// эти две строки- примерный аналог для slog log.Fatal(err)
		log.Error("failed to connect to db", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// 3. Настройка системы миграций Goose
	// Используем встроенную файловую систему (go:embed) для работы с SQL-файлами
	goose.SetBaseFS(migrations.FS) // migrations.FS — переменная с go:embed
	if err := goose.SetDialect("postgres"); err != nil {
		log.Error("Goose error", "err", err)
		os.Exit(1)
	}
	log.Info("Applying migrations...")

	// 4. Режим очистки БД (Drop/Reset, запуск с флагом -drop)
	// Если передан флаг очистки, откатываем все миграции перед стартом
	if cfg.DropDB {
		log.Info("Cleaning up the database...")
		// Ресет выполнит все Down блоки
		if err := goose.Reset(db.DB, "."); err != nil {
			log.Error("Goose reset error", "err", err)
			os.Exit(1)
		}
		log.Info("DB IS CLEAN.")
	}

	// 5. Применение актуальных миграций
	// Приводим схему БД к актуальному состоянию (команда Up)
	if err := goose.Up(db.DB, "."); err != nil {
		// Логируем ошибку миграции
		log.Error("Migration failed:", "err", err)
		os.Exit(1)
	}
	log.Info("Database is up to date!")

	// 6. Сборка зависимостей (Dependency Injection)
	// Инициализируем цепочку Репозиторий -> Сервис -> Хендлер
	repo := repository.NewRepository(db)
	services := service.NewService(repo)
	handlers := handler.NewHandler(services)

	// 7. Создаем корневой контекст, который отменится при сигналах завершения.
	// Слушаем сигналы ОС (SIGINT(Ctrl+C), SIGTERM, порядок перечисления НЕ ВАЖЕН) для корректной остановки. SIGTERM (Signal Terminate): стандартный сигнал для завершения процесса.
	// Его используют Docker, Kubernetes и системные менеджеры (типа systemd), когда производят Graceful Shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// 8. Запуск фоновых процессов. Инициализируем внешние клиенты и запускаем воркеры в неблокирующем режиме.

	// - Инициализируем клиент accrual
	accrualClient := accrual.NewClient(cfg.AccrualAddress)
	// -  Запускаем фоновый процесс
	worker.Run(ctx, repo, accrualClient, log)

	// 9. Запуск HTTP-сервера
	srv := new(server.Server)
	serverErrors := make(chan error, 1)

	go func() {
		log.Info("App is starting")
		// Запуск прослушивания порта, инициализируем роуты
		if err := srv.Run(cfg.ServerAddress, handlers.InitRoutes()); err != nil {
			// ErrServerClosed игнорируем, так как это штатная остановка.
			if !errors.Is(err, http.ErrServerClosed) {
				serverErrors <- fmt.Errorf("server listener crashed: %w", err)
			}
		}
	}()

	// 10. Ожидание событий завершения (блокировка основной горутины)
	select {
	case err := <-serverErrors:
		return err // Если сервер сам упал

	case <-ctx.Done(): // Сработает при SIGTERM/SIGINT
		log.Info("Shutting down gracefully", slog.String("signal", "interrupt"))

		// Даем 5 секунд на то, чтобы сервер и воркеры завершили активные запросы (timeout)
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("failed to shutdown http server: %w", err)
		}
	}

	log.Info("App exited cleanly")
	return nil
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
