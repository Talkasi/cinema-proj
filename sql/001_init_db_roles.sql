CREATE ROLE cinema_admin WITH LOGIN PASSWORD 'cinema_admin_password';
CREATE ROLE cinema_user WITH LOGIN PASSWORD 'cinema_user_password';
CREATE ROLE cinema_guest WITH LOGIN PASSWORD 'cinema_guest_password';

GRANT SELECT ON
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
TO cinema_guest;
GRANT INSERT ON users TO cinema_guest;

-- Подумать насчет RLS

GRANT cinema_guest TO cinema_user;
GRANT UPDATE ON users TO cinema_user;
GRANT UPDATE ON tickets TO cinema_user;
GRANT INSERT, UPDATE, DELETE ON reviews TO cinema_user;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO cinema_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO cinema_admin;
GRANT CREATE, USAGE ON SCHEMA public TO cinema_admin;
