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
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func Run(cfg *config.Config) error {
	log := setupLogger(cfg.Env)
	log.Info("init server", slog.String("address", cfg.ServerAddress))

	repository := repository.NewRepository(cfg)
	services := service.NewService(repository, cfg)
	handlers := handler.NewHandler(services)

	srv := new(server.Server)

	serverErrors := make(chan error, 1)
	// Сервер запускается в отдельной горутине (ListenAndServe() является блокирующим вызовом).
	go func() {
		log.Info("ShortenerApp is starting", slog.String("addr", cfg.ServerAddress))

		if err := srv.Run(cfg.ServerAddress, handlers.InitRoutes(log)); err != nil {

			if !errors.Is(err, http.ErrServerClosed) {
				serverErrors <- fmt.Errorf("server listener crashed: %w", err)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-serverErrors:
		return err

	case sig := <-quit:
		log.Info("ShortenerApp is shutting down", slog.String("signal", sig.String()))

		signal.Stop(quit)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown http server: %w", err)
		}
	}

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
