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
    screen_types, 
    seats,
    seat_types, 
    reviews,
    users
FROM cinema_guest;
REVOKE INSERT ON users FROM cinema_guest;