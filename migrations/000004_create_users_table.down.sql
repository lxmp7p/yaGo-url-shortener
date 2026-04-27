-- migrations/000004_create_users_table.down.sql
-- Удаление таблицы юзеров
ALTER TABLE urls
DROP COLUMN owner_id;