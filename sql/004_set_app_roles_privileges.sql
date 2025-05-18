GRANT SELECT ON
    movies, 
    movie_shows, 
    halls, 
    tickets,
    genres, 
    screen_types, 
    seats,
    seat_types, 
    reviews,
    users,
    movies_genres
TO cinema_guest;
GRANT INSERT ON users TO cinema_guest;

-- Подумать насчет RLS

GRANT cinema_guest TO cinema_user;
GRANT UPDATE ON users TO cinema_user;
GRANT UPDATE ON tickets TO cinema_user;
GRANT UPDATE ON movies TO cinema_test_user;
GRANT INSERT, UPDATE, DELETE ON reviews TO cinema_user;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO cinema_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO cinema_admin;
GRANT CREATE, USAGE ON SCHEMA public TO cinema_admin;
