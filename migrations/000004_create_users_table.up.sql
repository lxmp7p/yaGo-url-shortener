-- migrations/000004_create_users_table.up.sql
-- Создание таблицы юзеров
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL
);
