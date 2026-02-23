-- +goose Up
-- +goose StatementBegin
-- 1. Таблица заказов
-- ❌ПРОВЕРИТЬ КАК В ОРИГИНАЛАХ- ♊ и уже готовом
-- 1. Создаем пользовательский тип данных
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'task_status') THEN
        CREATE TYPE task_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
    END IF;
END $$;
-- 15:19
CREATE TABLE orders (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY, -- технический первичный ключ для связей внутри БД
    order_number TEXT NOT NULL UNIQUE,
    order_id UUID DEFAULT gen_random_uuid (), -- внутреннй, уникальный номер заказа (например, для публичного API)
    -- Если будет нужно получить этот сгенерированный UUID обратно в код Go, нужно использовать конструкцию RETURNING order_id в конце запроса.
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT ON UPDATE CASCADE, -- id пользователя, т.к. вводить номер могут только зарегистрированные пользователи
    status task_status DEFAULT 'NEW', -- предопределенные статусы заказа, их меняет внешний сервис
    accrual NUMERIC(10, 2) DEFAULT 0, -- начисление по заказу, из внешнего сервиса
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(), -- ex. uploaded_at 
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    attempts SMALLINT DEFAULT 0
);

CREATE INDEX idx_orders_user_id ON orders (user_id);

-- 2. Таблица баланса (связь 1-to-1 с пользователем)
CREATE TABLE balance (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT PRIMARY KEY REFERENCES users (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    balance NUMERIC(10, 2) DEFAULT 0,
    debited NUMERIC(10, 2) DEFAULT 0
);

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

DROP TABLE IF EXISTS balance;

DROP TABLE IF EXISTS orders;
-- +goose StatementEnd