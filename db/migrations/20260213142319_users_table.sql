-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY, -- современный стиль (2026) вместо BIGSERIAL
    login VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
-- TODO: ИНДЕКС для login убрать TIME ZONE?
-- Оставляй WITH TIME ZONE. Это стандарт 2026 года для Go (time.Time),
-- так как он избавляет от проблем при переносе серверов между дата-центрами
-- Индекс для login: Postgres автоматически создает индекс для полей UNIQUE,
-- так что отдельный CREATE INDEX для login не нужен

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd