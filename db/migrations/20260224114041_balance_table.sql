-- +goose Up
-- +goose StatementBegin
-- 2. Таблица баланса (связь 1-to-1 с пользователем)
CREATE TABLE balance (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    balance NUMERIC(10, 2) DEFAULT 0,
    debited NUMERIC(10, 2) DEFAULT 0
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS balance;
-- +goose StatementEnd