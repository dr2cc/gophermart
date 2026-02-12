// Package app configures and runs application.
package app

import (
	"context"
	"errors"
	"fmt"
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

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func Run(cfg *config.Config) error {
	log := setupLogger(cfg.Env)
	log.Info("Init server", slog.String("address", cfg.ServerAddress))

	db, err := sqlx.Connect("postgres", cfg.DatabaseDSN)
	if err != nil {
		log.Error("failed to connect to db", "err", err)
		// До горутины. Сразу завершаем работу.
		os.Exit(1)
	}
	defer db.Close()
	//

	repository := repository.NewRepository(db)
	services := service.NewService(repository)
	handlers := handler.NewHandler(services)

	// Создаем корневой контекст, который отменится при сигналах завершения
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// // Инициализируем клиент accrual и запускаем фоновый процессор
	// accrualClient := accrual.NewClient(cfg.AccrualAddress)
	// // Передаем ctx внутрь. Когда в консоли нажмут Ctrl+C, в процессоре сработает <-ctx.Done()
	// processor.Run(ctx, repository, accrualClient)

	// 3. Настройка сервера
	srv := new(server.Server)
	serverErrors := make(chan error, 1)

	go func() {
		log.Info("App is starting")
		// Инициализируем роуты
		if err := srv.Run(cfg.ServerAddress, handlers.InitRoutes()); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				serverErrors <- fmt.Errorf("server listener crashed: %w", err)
			}
		}
	}()

	// 4. Ожидание завершения
	select {
	case err := <-serverErrors:
		return err // Если сервер сам упал

	case <-ctx.Done(): // Сработает при SIGTERM/SIGINT
		log.Info("Shutting down gracefully", slog.String("signal", "interrupt"))

		// Даем 5 секунд на то, чтобы сервер и воркеры завершили текущие дела
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

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
