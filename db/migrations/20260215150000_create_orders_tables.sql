-- +goose Up
-- +goose StatementBegin
-- 1. Таблица заказов
-- ❌ПРОВЕРИТЬ КАК В ОРИГИНАЛАХ- ♊ и уже готовом
CREATE TABLE orders (
    -- number
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    order_id UUID, -- новый, с 9:26. Разобраться!
    status VARCHAR(50) NOT NULL CHECK (
        status IN (
            'NEW',
            'PROCESSING',
            'INVALID',
            'PROCESSED'
        )
    ),
    accrual NUMERIC(10, 2) DEFAULT 0, -- видимо сюда записывает свои данные accrual
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(), -- ex. uploaded_at 
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    attempts SMALLINT DEFAULT 0
);
-- TODO: ENUM для status - VARCHAR нет!
CREATE INDEX idx_orders_user_id ON orders (user_id);

-- 2. Таблица баланса (связь 1-to-1 с пользователем)
CREATE TABLE balance (
    user_id BIGINT PRIMARY KEY REFERENCES users (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    balance NUMERIC(10, 2) DEFAULT 0,
    debited NUMERIC(10, 2) DEFAULT 0
);

-- 3. Таблица движений по балансу!!❗
CREATE TABLE balance_history (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY, -- Современный стиль вместо BIGSERIAL
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    order_num VARCHAR(32) NOT NULL,
    sum NUMERIC(10, 2) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_balance_history_user_id ON balance_history (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE balance_history;

DROP TABLE balance;

DROP TABLE orders;
-- +goose StatementEnd