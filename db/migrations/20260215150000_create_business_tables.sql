-- +goose Up
-- +goose StatementBegin
-- 1. Таблица заказов
CREATE TABLE IF NOT EXISTS orders (
    number VARCHAR(32) PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    status VARCHAR(50) NOT NULL CHECK (
        status IN (
            'NEW',
            'PROCESSING',
            'INVALID',
            'PROCESSED'
        )
    ),
    accrual NUMERIC(10, 2) DEFAULT 0,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    attempts SMALLINT DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders (user_id);

-- 2. Таблица баланса (связь 1-to-1 с пользователем)
CREATE TABLE IF NOT EXISTS balance (
    user_id BIGINT PRIMARY KEY REFERENCES users (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    balance NUMERIC(10, 2) DEFAULT 0,
    debited NUMERIC(10, 2) DEFAULT 0
);

-- 3. Таблица списаний
CREATE TABLE IF NOT EXISTS withdrawals (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY, -- Современный стиль вместо BIGSERIAL
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    order_num VARCHAR(32) NOT NULL,
    sum NUMERIC(10, 2) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id ON withdrawals (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS withdrawals;

DROP TABLE IF EXISTS balance;

DROP TABLE IF EXISTS orders;
-- +goose StatementEnd