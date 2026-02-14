// Package app configures and runs application.
package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"gophermart/db/dbgen"
	"gophermart/db/migrations" // импорт вашего пакета с FS
	"gophermart/internal/accrual"
	"gophermart/internal/accrual/processor"
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
	"github.com/pressly/goose/v3"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func Run(cfg *config.Config) error {
	log := setupLogger(cfg.Env)
	log.Info("Init server", slog.String("address", cfg.ServerAddress))

	// 1. Добавляем флаг (нужно импортировать пакет "flag")
	dropDB := flag.Bool("drop", false, "снести все таблицы перед стартом")
	flag.Parse()

	// 2. Коннект
	db, err := sqlx.Connect("postgres", cfg.DatabaseDSN)
	if err != nil {
		// это примерный аналог log.Fatal(err) для slog
		log.Error("failed to connect to db", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// 3. Запуск миграций Goose перед стартом логики (проверка наличия таблиц и их создание)
	// Передаем стандартный *sql.DB через db.DB
	goose.SetBaseFS(migrations.FS) // migrations.FS — переменная с go:embed
	if err := goose.SetDialect("postgres"); err != nil {
		log.Error("Goose error", "err", err)
		os.Exit(1)
	}
	log.Info("Applying migrations...")

	// 4. МАГИЯ: Если запустили с флагом -drop
	if *dropDB {
		log.Info("Cleaning up the database...")
		// Ресет выполнит все Down блоки из твоей FS
		if err := goose.Reset(db.DB, "."); err != nil {
			log.Error("Goose reset error", "err", err)
			os.Exit(1)
		}
		log.Info("Database is clean!")
	}

	// 5. Обычный запуск миграций (накатка)
	if err := goose.Up(db.DB, "."); err != nil {
		// Логируем ошибку миграции
		log.Error("Migration failed:", "err", err)
		os.Exit(1)
	}

	log.Info("Database is up to date!")

	// 6. Инициализация sqlc (dbgen)
	// sqlx.DB отлично подходит, так как реализует интерфейс DBTX
	store := dbgen.New(db)

	// 7. Пример использования типизированного метода
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := store.GetUserByLogin(ctx, "admin")
	if err != nil {
		// Если пользователь не найден, sqlc вернет sql.ErrNoRows
	}

	// Теперь у вас есть доступ к полям через точку с правильными типами
	println("User ID:", user.ID)
	println("User Login:", user.Login)
	//

	repository := repository.NewRepository(db)
	services := service.NewService(repository)
	handlers := handler.NewHandler(services)

	// Создаем корневой контекст, который отменится при сигналах завершения
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Инициализируем клиент accrual и запускаем фоновый процессор
	accrualClient := accrual.NewClient(cfg.AccrualAddress)
	// Передаем ctx внутрь. Когда в консоли нажмут Ctrl+C, в процессоре сработает <-ctx.Done()
	processor.Run(ctx, repository.OrderStore, accrualClient, log)

	// Настройка сервера
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

	// Ожидание завершения
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
