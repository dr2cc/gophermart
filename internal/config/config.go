// Package config contains configuration for application.
package config

import (
	"flag"
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env            string `env:"ENV" env-default:"local"`
	ServerAddress  string `env:"RUN_ADDRESS" env-default:":8080"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" env-default:"localhost:8090"`
	DatabaseDSN    string `env:"DATABASE_URI"`
	PollInterval   int    `env:"POLL_INTERVAL" env-default:"3"`
	DropDB         bool
}

func NewConfig() (*Config, error) {

	var cfg Config

	// 1. Загружаем переменные окружения и дефолты
	// Если ENV нет, подставятся значения из env-default
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read env: %w", err)
	}
	// 2. Настраиваем флаги.
	// В качестве дефолтов передаем уже заполненные значения из cfg (там сейчас ENV или дефолты)
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "host to listen on")
	flag.StringVar(&cfg.AccrualAddress, "r", cfg.AccrualAddress, "accrual is listening on")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "database dsn")
	flag.BoolVar(&cfg.DropDB, "drop", false, "drop db tables")

	// 3. Парсим флаги.
	// Если пользователь ввел флаг, он перезапишет значение из ENV.
	flag.Parse()

	return &cfg, nil
}
