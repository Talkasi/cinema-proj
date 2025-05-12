-- Удаляем триггеры
DROP TRIGGER IF EXISTS update_movie_revenue_when_ticket_status_changed ON tickets;
DROP TRIGGER IF EXISTS check_movie_show_conflict_before_insert_or_update ON movie_shows;

-- Удаляем функции
DROP FUNCTION IF EXISTS update_box_office_revenue();
DROP FUNCTION IF EXISTS check_movie_show_conflict();

-- Удаляем таблицы
DROP TABLE IF EXISTS tickets CASCADE;
DROP TABLE IF EXISTS reviews CASCADE;
DROP TABLE IF EXISTS movie_shows CASCADE;
DROP TABLE IF EXISTS seats CASCADE;
DROP TABLE IF EXISTS movies_genres CASCADE;
DROP TABLE IF EXISTS halls CASCADE;
DROP TABLE IF EXISTS equipment_types CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS genres CASCADE;
DROP TABLE IF EXISTS movies CASCADE;
DROP TABLE IF EXISTS seat_types CASCADE;

-- Удаляем типы
DROP TYPE IF EXISTS ticket_status_enum;
DROP TYPE IF EXISTS language_enum;

-- Удаляем расширение
DROP EXTENSION IF EXISTS "uuid-ossp";