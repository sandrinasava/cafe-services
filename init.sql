

-- Создание таблицы users, если она не существует
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    customer id TEXT NOT NULL,
    items TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для быстрого поиска по customer
CREATE INDEX idx_orders_customer ON orders(customer);

-- -- Пример вставки тестовых данных
-- INSERT INTO orders (customer, items, status) VALUES
-- ('John Doe', '["item1", "item2"]', 'received'),
-- ('Jane Smith', '["item3"]', 'received');