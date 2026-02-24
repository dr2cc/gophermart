-- +goose Up
-- +goose StatementBegin
-- 2. Таблица баланса (связь 1-to-1 с пользователем)
-- 3. Таблица движений по балансу!!❗
CREATE TABLE balance_history (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    order_num VARCHAR(32) NOT NULL,
    sum NUMERIC(10, 2) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_balance_history_user_id ON balance_history (user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS balance_history;
-- +goose StatementEnd