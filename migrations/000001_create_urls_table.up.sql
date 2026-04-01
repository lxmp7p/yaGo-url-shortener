-- migrations/000001_create_url_table.up.sql
-- Создание таблицы ссылок
CREATE TABLE urls (
    id SERIAL PRIMARY KEY,
    original VARCHAR(255) NOT NULL,
    short VARCHAR(255) NOT NULL
);

CREATE INDEX idx_url_original ON urls(original);
CREATE INDEX idx_url_short ON urls(short); 