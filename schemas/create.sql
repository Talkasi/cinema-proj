CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    birth_date DATE,
    is_blocked BOOLEAN DEFAULT FALSE,
    is_admin BOOLEAN DEFAULT FALSE
);

-- Таблица жанров
CREATE TABLE IF NOT EXISTS genres (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(64) NOT NULL UNIQUE,
    description VARCHAR(1000)
);

-- Таблица фильмов
CREATE TABLE IF NOT EXISTS movies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(200) NOT NULL,
    duration TIME NOT NULL,
    rating DECIMAL(3,2),
    description VARCHAR(1000),
    age_limit INT,
    box_office_revenue DECIMAL(15,2) DEFAULT 0,
    release_date DATE
);

-- Таблица equipment_types
CREATE TABLE IF NOT EXISTS equipment_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(1000)
);

-- Таблица залов
CREATE TABLE IF NOT EXISTS halls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    equipment_type_id UUID REFERENCES equipment_types(id),
    name VARCHAR(100) NOT NULL UNIQUE,
    capacity INT,
    description VARCHAR(1000)
);

-- Таблица связей фильм-жанр
CREATE TABLE IF NOT EXISTS movies_genres (
    movie_id UUID REFERENCES movies(id),
    genre_id UUID REFERENCES genres(id),
    PRIMARY KEY (movie_id, genre_id)
);

-- Таблица сеансов
CREATE TABLE IF NOT EXISTS movie_shows (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    movie_id UUID REFERENCES movies(id),
    hall_id UUID REFERENCES halls(id),
    start_time TIME,
    language VARCHAR(50)
);

-- Типы мест
CREATE TABLE IF NOT EXISTS seat_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(1000)
);

-- Места
CREATE TABLE IF NOT EXISTS seats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    movie_show_id UUID REFERENCES movie_shows(id),
    seat_type_id UUID REFERENCES seat_types(id),
    row_number INTEGER,
    seat_number NUMERIC
);

-- Статусы билетов
CREATE TABLE IF NOT EXISTS ticket_status (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE
);

-- Билеты
CREATE TABLE IF NOT EXISTS tickets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    movie_show_id UUID REFERENCES movie_shows(id),
    seat_id UUID REFERENCES seats(id),
    ticket_status_id UUID REFERENCES ticket_status(id),
    price DECIMAL(10,2)
);

-- Отзывы
CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    movie_id UUID REFERENCES movies(id),
    rating DECIMAL(3,2),
    review_comment TEXT
);
