-- Установка pgcrypto для исп-я ф-ии gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Создание таблицы users, если она не существует
CREATE TABLE IF NOT EXISTS users (
    user_UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS orders (
    order_UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_UUID UUID NOT NULL,
    items VARCHAR(255)NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    FOREIGN KEY (user_UUID) REFERENCES users(user_UUID) ON DELETE CASCADE
);

-- Индекс для быстрого поиска 
CREATE INDEX IF NOT EXISTS idx_orders_order_UUID ON orders(order_UUID);
