\copy genres(id, name, description) FROM '/docker-entrypoint-initdb.d/data/genres.csv' DELIMITER ',' CSV HEADER;

\copy screen_types(id, name, description, price_modifier) FROM '/docker-entrypoint-initdb.d/data/screen_types.csv' DELIMITER ',' CSV HEADER;

\copy seat_types(id, name, description, price_modifier) FROM '/docker-entrypoint-initdb.d/data/seat_types.csv' DELIMITER ',' CSV HEADER;

\copy users(id, name, email, password_hash, birth_date, is_admin) FROM '/docker-entrypoint-initdb.d/data/users.csv' DELIMITER ',' CSV HEADER;

\copy halls(id, screen_type_id, name, description) FROM '/docker-entrypoint-initdb.d/data/halls.csv' DELIMITER ',' CSV HEADER;

\copy seats(id, hall_id, seat_type_id, row_number, seat_number) FROM '/docker-entrypoint-initdb.d/data/seats.csv' DELIMITER ',' CSV HEADER;

\copy movies(id, title, duration, description, age_limit, box_office_revenue, release_date) FROM '/docker-entrypoint-initdb.d/data/movies.csv' DELIMITER ',' CSV HEADER;

\copy movies_genres(movie_id, genre_id) FROM '/docker-entrypoint-initdb.d/data/movies_genres.csv' DELIMITER ',' CSV HEADER;

\copy movie_shows(id, movie_id, hall_id, start_time, language) FROM '/docker-entrypoint-initdb.d/data/movie_shows.csv' DELIMITER ',' CSV HEADER;

\copy tickets(id, movie_show_id, seat_id, user_id, ticket_Status, price) FROM '/docker-entrypoint-initdb.d/data/tickets.csv' DELIMITER ',' CSV HEADER;

\copy reviews(id, user_id, movie_id, rating, review_comment) FROM '/docker-entrypoint-initdb.d/data/reviews.csv' DELIMITER ',' CSV HEADER;

\! touch /var/lib/postgresql/data/.initialized