
DROP TRIGGER IF EXISTS update_movie_revenue_when_ticket_Status_changed ON tickets;
DROP TRIGGER IF EXISTS check_movie_show_on_insert ON movie_shows;
DROP TRIGGER IF EXISTS check_movie_show_on_update ON movie_shows;

DROP INDEX IF EXISTS idx_users_email;

DROP FUNCTION IF EXISTS update_box_office_revenue();
DROP FUNCTION IF EXISTS check_movie_show_conflict();
DROP FUNCTION IF EXISTS create_movie_show_with_tickets;

DROP PROCEDURE update_movie(
    UUID,
    VARCHAR(200),
    TIME,
    VARCHAR(1000),
    INT,
    DATE,
    UUID[]
);

DROP TABLE IF EXISTS tickets CASCADE;
DROP TABLE IF EXISTS reviews CASCADE;
DROP TABLE IF EXISTS movie_shows CASCADE;
DROP TABLE IF EXISTS seats CASCADE;
DROP TABLE IF EXISTS movies_genres CASCADE;
DROP TABLE IF EXISTS halls CASCADE;
DROP TABLE IF EXISTS screen_types CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS genres CASCADE;
DROP TABLE IF EXISTS movies CASCADE;
DROP TABLE IF EXISTS seat_types CASCADE;

DROP TYPE IF EXISTS ticket_Status_enum;
DROP TYPE IF EXISTS language_enum;

DROP EXTENSION IF EXISTS "uuid-ossp";