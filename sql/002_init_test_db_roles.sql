CREATE ROLE cinema_test_admin WITH LOGIN PASSWORD 'cinema_test_admin_password';
CREATE ROLE cinema_test_user WITH LOGIN PASSWORD 'cinema_test_user_password';
CREATE ROLE cinema_test_guest WITH LOGIN PASSWORD 'cinema_test_guest_password';

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
TO cinema_test_guest;
GRANT INSERT ON users TO cinema_test_guest;

-- Подумать насчет RLS

GRANT cinema_test_guest TO cinema_test_user;
GRANT UPDATE ON users TO cinema_test_user;
GRANT UPDATE ON tickets TO cinema_test_user;
GRANT INSERT, UPDATE, DELETE ON reviews TO cinema_test_user;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO cinema_test_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO cinema_test_admin;
GRANT CREATE, USAGE ON SCHEMA public TO cinema_test_admin;