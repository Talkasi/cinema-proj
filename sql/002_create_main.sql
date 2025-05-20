CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    birth_date DATE NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE,
    CONSTRAINT valid_name CHECK (name ~ '\S'),
    CONSTRAINT email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$'),
    CONSTRAINT valid_birth_date CHECK (birth_date <= CURRENT_DATE AND birth_date >= CURRENT_DATE - INTERVAL '100 years')
);

CREATE TABLE IF NOT EXISTS genres (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(64) NOT NULL UNIQUE,
    description VARCHAR(1000) NOT NULL,
    CONSTRAINT valid_name CHECK (name ~ '^[A-Za-zА-Яа-яЁё\s-]+$' AND name ~ '\S'),
    CONSTRAINT valid_description CHECK (description ~ '\S')
);

CREATE TABLE IF NOT EXISTS movies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(200) NOT NULL,
    duration TIME NOT NULL,
    description VARCHAR(1000) NOT NULL,
    age_limit INT NOT NULL DEFAULT 0,
    box_office_revenue DECIMAL(15,2) NOT NULL DEFAULT 0 CHECK (box_office_revenue >= 0),
    release_date DATE NOT NULL, -- can be IN the future
    CONSTRAINT valid_title CHECK (title ~ '\S'),
    CONSTRAINT valid_duration CHECK (duration > '00:00:00'),
    CONSTRAINT valid_description CHECK (description ~ '\S'),
    CONSTRAINT valid_age_limit CHECK (age_limit IN (0, 6, 12, 16, 18))
);

CREATE TABLE IF NOT EXISTS screen_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(1000) NOT NULL,
    CONSTRAINT valid_name CHECK (name ~ '\S'),
    CONSTRAINT valid_description CHECK (description ~ '\S')
);

ALTER TABLE screen_types ADD COLUMN price_modifier DECIMAL(3,2) 
DEFAULT 1.0 CHECK (price_modifier > 0);

CREATE TABLE IF NOT EXISTS halls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    screen_type_id UUID REFERENCES screen_types(id),
    name VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(1000),
    CONSTRAINT valid_name CHECK (
        name ~ '^[a-zA-Zа-яА-Я0-9\s\.\-_#№]+$' AND
        name ~ '\S' AND
        length(name) > 0 AND
        length(name) <= 100
    ),
    CONSTRAINT valid_description CHECK (description IS NULL OR description ~ '\S')
);

CREATE TABLE IF NOT EXISTS movies_genres (
    movie_id UUID REFERENCES movies(id) ON DELETE CASCADE,
    genre_id UUID REFERENCES genres(id),
    PRIMARY KEY (movie_id, genre_id)
);

CREATE TYPE language_enum AS ENUM (
    'English',
    'Spanish',
    'French',
    'German',
    'Italian',
    'Русский'
);

CREATE TABLE IF NOT EXISTS movie_shows ( -- Тут надо проверять при вставке, что не конфликтуют показы между собой
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    movie_id UUID REFERENCES movies(id),
    hall_id UUID REFERENCES halls(id),
    start_time TIMESTAMP NOT NULL CHECK (start_time > '1895-03-22'),
    language language_enum NOT NULL
);

CREATE OR REPLACE FUNCTION check_movie_show_conflict()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM movie_shows
        WHERE hall_id = NEW.hall_id
        AND id <> NEW.id
        AND (
            (start_time < NEW.start_time + (SELECT duration FROM movies WHERE id = NEW.movie_id) + INTERVAL '10 minutes' 
             AND 
             start_time + (SELECT duration FROM movies WHERE id = movie_shows.movie_id) + INTERVAL '10 minutes' > NEW.start_time)
            OR 
            (NEW.start_time < start_time + (SELECT duration FROM movies WHERE id = movie_shows.movie_id) + INTERVAL '10 minutes' 
             AND 
             NEW.start_time + (SELECT duration FROM movies WHERE id = NEW.movie_id) + INTERVAL '10 minutes' > start_time)
        )
    ) THEN
        RAISE EXCEPTION 'Невозможно запланировать показ, поскольку в это время кинозал будет занят показом другого фильма или будет проводиться уборка';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER check_movie_show_conflict_before_insert_or_update
BEFORE INSERT OR UPDATE ON movie_shows
FOR EACH ROW
EXECUTE FUNCTION check_movie_show_conflict();

CREATE TABLE IF NOT EXISTS seat_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(1000) NOT NULL,
    CONSTRAINT valid_name CHECK (name ~ '\S'),
    CONSTRAINT valid_description CHECK (description ~ '\S')
);

ALTER TABLE seat_types ADD COLUMN price_modifier DECIMAL(3,2) 
DEFAULT 1.0 CHECK (price_modifier > 0);

CREATE TABLE IF NOT EXISTS seats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    hall_id UUID REFERENCES halls(id),
    seat_type_id UUID REFERENCES seat_types(id),
    row_number INTEGER NOT NULL CHECK (row_number > 0),
    seat_number INTEGER NOT NULL CHECK (seat_number > 0),
    CONSTRAINT unique_seat UNIQUE (hall_id, row_number, seat_number)
);

CREATE TYPE ticket_status_enum AS ENUM (
    'Purchased',
    'Reserved',
    'Available'
);

CREATE TABLE IF NOT EXISTS tickets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    movie_show_id UUID REFERENCES movie_shows(id),
    seat_id UUID REFERENCES seats(id),
    user_id UUID REFERENCES users(id),
    ticket_status ticket_status_enum NOT NULL,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    CONSTRAINT unique_ticket UNIQUE (movie_show_id, seat_id),
    CONSTRAINT user_id_status_check CHECK (
        (user_id IS NULL AND ticket_status = 'Available') OR
        (user_id IS NOT NULL)
    )
);

CREATE OR REPLACE FUNCTION update_box_office_revenue()
RETURNS TRIGGER AS $$
BEGIN
    -- Если статус билета изменился на "Купленный"
    IF NEW.ticket_status = 'Purchased' AND OLD.ticket_status <> 'Purchased' THEN
        UPDATE movies
        SET box_office_revenue = box_office_revenue + NEW.price
        WHERE id = (SELECT movie_id FROM movie_shows WHERE id = NEW.movie_show_id);
    
    -- Если статус билета изменился с "Купленного" на другой статус
    ELSIF OLD.ticket_status = 'Purchased' AND NEW.ticket_status <> 'Purchased' THEN
        UPDATE movies
        SET box_office_revenue = box_office_revenue - OLD.price
        WHERE id = (SELECT movie_id FROM movie_shows WHERE id = NEW.movie_show_id);
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_movie_revenue_when_ticket_status_changed
BEFORE UPDATE ON tickets
FOR EACH ROW
EXECUTE FUNCTION update_box_office_revenue();

CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    movie_id UUID REFERENCES movies(id) ON DELETE CASCADE,
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 10),
    review_comment TEXT,
    CONSTRAINT unique_review UNIQUE (user_id, movie_id),
    CONSTRAINT valid_review_comment CHECK (review_comment IS NULL OR review_comment ~ '\S')
);

CREATE OR REPLACE FUNCTION create_movie_show_with_tickets(
    p_movie_id UUID,
    p_hall_id UUID,
    p_start_time TIMESTAMP,
    p_language language_enum,
    p_base_price DECIMAL(10,2))
RETURNS UUID AS $$
DECLARE
    v_show_id UUID;
    v_screen_modifier DECIMAL(5,2);
    v_seat RECORD;
    v_price DECIMAL(10,2);
BEGIN
    INSERT INTO movie_shows (id, movie_id, hall_id, start_time, language)
    VALUES (uuid_generate_v4(), p_movie_id, p_hall_id, p_start_time, p_language)
    RETURNING id INTO v_show_id;

    SELECT st.price_modifier INTO v_screen_modifier
    FROM screen_types st
    JOIN halls h ON st.id = h.screen_type_id
    WHERE h.id = p_hall_id;

    FOR v_seat IN (
        SELECT s.id, st.price_modifier as seat_modifier
        FROM seats s
        JOIN seat_types st ON s.seat_type_id = st.id
        WHERE s.hall_id = p_hall_id
    ) LOOP
        v_price := ROUND(p_base_price * v_screen_modifier * v_seat.seat_modifier, 2);
        
        INSERT INTO tickets (id, movie_show_id, seat_id, ticket_status, price)
        VALUES (uuid_generate_v4(), v_show_id, v_seat.id, 'Available', v_price);
    END LOOP;

    RETURN v_show_id;
EXCEPTION
    WHEN OTHERS THEN
        RAISE EXCEPTION 'Ошибка при создании сеанса: %', SQLERRM;
END;
$$ LANGUAGE plpgsql;