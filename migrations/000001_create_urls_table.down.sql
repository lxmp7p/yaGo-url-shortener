-- migrations/000001_create_urls_table.down.sql
-- Откат создания таблицы ссылок
DROP INDEX IF EXISTS idx_urls_original;
DROP INDEX IF EXISTS idx_urls_short;
DROP TABLE IF EXISTS urls; 