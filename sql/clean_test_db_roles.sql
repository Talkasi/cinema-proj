-- Revoke all privileges from cinema_test_admin
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM cinema_test_admin;
REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM cinema_test_admin;
REVOKE CREATE, USAGE ON SCHEMA public FROM cinema_test_admin;

-- Revoke privileges from cinema_test_user
REVOKE UPDATE ON users FROM cinema_test_user;
REVOKE UPDATE ON tickets FROM cinema_test_user;
REVOKE INSERT, UPDATE, DELETE ON reviews FROM cinema_test_user;

-- Revoke cinema_test_guest role from cinema_test_user
REVOKE cinema_test_guest FROM cinema_test_user;

-- Revoke privileges from cinema_test_guest
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
FROM cinema_test_guest;
REVOKE INSERT ON users FROM cinema_test_guest;

-- Drop all roles
DROP ROLE IF EXISTS cinema_test_admin;
DROP ROLE IF EXISTS cinema_test_user;
DROP ROLE IF EXISTS cinema_test_guest;