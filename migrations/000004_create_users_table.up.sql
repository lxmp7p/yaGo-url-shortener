-- migrations/000004_create_users_table.down.sql
-- создание юзеров
ALTER TABLE urls
ADD COLUMN owner_id VARCHAR(255) NOT NULL;