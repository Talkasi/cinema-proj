-- Revoke all privileges from cinema_admin
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM cinema_admin;
REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM cinema_admin;
REVOKE CREATE, USAGE ON SCHEMA public FROM cinema_admin;

-- Revoke privileges from cinema_user
REVOKE UPDATE ON users FROM cinema_user;
REVOKE UPDATE ON tickets FROM cinema_user;
REVOKE INSERT, UPDATE, DELETE ON reviews FROM cinema_user;

-- Revoke cinema_guest role from cinema_user
REVOKE cinema_guest FROM cinema_user;

-- Revoke privileges from cinema_guest
REVOKE SELECT ON 
    movies, 
    movie_shows, 
    halls, 
    tickets,
    genres, 
    equipment_types, 
    seats,
    seat_types, 
    reviews,
    users
FROM cinema_guest;
REVOKE INSERT ON users FROM cinema_guest;

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

-- Drop all roles
DROP ROLE IF EXISTS cinema_admin;
DROP ROLE IF EXISTS cinema_user;
DROP ROLE IF EXISTS cinema_guest;

