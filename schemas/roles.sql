CREATE ROLE guest;
CREATE ROLE ruser;
CREATE ROLE admin;

GRANT SELECT ON 
    movies, 
    movie_shows, 
    halls, 
    tickets, 
    ticket_status,
    genres,
    equipment_types,
    seats,
    seat_types,
    reviews,
    users
TO guest;

GRANT INSERT ON users TO guest;

GRANT guest TO ruser;
GRANT 
    SELECT, INSERT, UPDATE 
ON 
    tickets,
    reviews,
    seats,
    users
TO ruser;

GRANT ALL PRIVILEGES ON 
    ALL TABLES IN SCHEMA public 
TO admin;

GRANT CREATE ON DATABASE cinema_test TO admin;
GRANT USAGE ON SCHEMA public TO admin;
GRANT CREATE ON SCHEMA public TO admin;


ALTER ROLE admin WITH LOGIN PASSWORD 'admin555';
ALTER ROLE ruser WITH LOGIN PASSWORD 'user111';
ALTER ROLE guest WITH LOGIN PASSWORD 'guest111';





CREATE ROLE guest_test;
CREATE ROLE ruser_test;
CREATE ROLE admin_test;

GRANT SELECT ON 
    movies, 
    movie_shows, 
    halls, 
    tickets, 
    ticket_status,
    genres,
    equipment_types,
    seats,
    seat_types,
    reviews,
    users
TO guest_test;

GRANT INSERT ON users TO guest_test;

GRANT guest_test TO ruser_test;
GRANT 
    SELECT, INSERT, UPDATE 
ON 
    tickets,
    reviews,
    seats,
    users
TO ruser_test;

GRANT ALL PRIVILEGES ON 
    ALL TABLES IN SCHEMA public 
TO admin_test;

GRANT CREATE ON DATABASE cinema_test TO admin_test;
GRANT USAGE ON SCHEMA public TO admin;
GRANT CREATE ON SCHEMA public TO admin;

GRANT guest_test TO ruser_test;
GRANT 
    SELECT, INSERT, UPDATE 
ON 
    tickets,
    reviews,
    seats,
    users,
    halls
TO ruser_test;

GRANT ALL PRIVILEGES ON 
    ALL TABLES IN SCHEMA public 
TO admin_test;

GRANT CREATE ON DATABASE cinema_test TO admin_test;
GRANT USAGE ON SCHEMA public TO admin;
GRANT CREATE ON SCHEMA public TO admin;

-- Предоставьте права на использование схемы
GRANT USAGE ON SCHEMA public TO admin;

-- Предоставьте права на создание объектов в схеме
GRANT CREATE ON SCHEMA public TO admin;

-- Если необходимо, предоставьте права на все существующие таблицы в схеме
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO admin;

-- Если необходимо, предоставьте права на все существующие последовательности в схеме
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO admin;


ALTER ROLE admin_test WITH LOGIN PASSWORD 'admin555';
ALTER ROLE ruser_test WITH LOGIN PASSWORD 'user111';
ALTER ROLE guest_test WITH LOGIN PASSWORD 'guest111';


GRANT ALL PRIVILEGES ON DATABASE cinema_test TO admin_test;
GRANT ALL ON SCHEMA public TO admin_test;