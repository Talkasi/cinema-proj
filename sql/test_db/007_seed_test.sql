
INSERT INTO users (name, email, password_hash, birth_date, is_admin) VALUES
('Ivan Ivanov', 'ivan@example.com', 'hashed_password_1', '1990-01-01', FALSE),
('Mariya Key', 'maria@example.com', 'hashed_password_2', '1985-05-15', FALSE),
('smirnov532', 'alexey@example.com', 'hashed_password_3', '1978-10-20', TRUE),
('Olga Kuznetsova', 'olga@example.com', 'hashed_password_4', '1995-03-30', FALSE);

INSERT INTO genres (name, description) VALUES
('Genre 1', 'Filmy, kotorye issleduyut chelovecheskie emotsii i konflikty.'),
('Komediya', 'Filmy, kotorye prednaznacheny dlya razvlecheniya i smekha.'),
('Genre 2', 'Filmy s vysokim urovnem napryazheniya i dinamichnymi stsenami.'),
('Uzhasy', 'Filmy, kotorye vyzyvayut strakh i napryazhenie.'),
('Nauchnaya fantastika', 'Filmy, osnovannye na nauchnykh kontseptsiyakh i tekhnologiyakh.'),
('Genre 3', 'Filmy, osnovannye na voobrazhaemykh mirakh i magii.'),
('Genre 4', 'Genre 5'),
('Romantika', 'Filmy, kotorye issleduyut lyubovnye otnosheniya mezhdu personazhami.'),
('Semeynyy', 'Filmy, podkhodyaschie dlya prosmotra vsey semey, chasto s pozitivnymi i vdokhnovlyayuschimi temami.'),
('Nauchnyy', 'Filmy, kotorye issleduyut nauchnye kontseptsii i idei.');

INSERT INTO screen_types (name, description, price_modifier) VALUES
('2D', 'Sample English text 1', 1.0),
('3D', 'Sample English text 2', 1.5),
('IMAX', 'Bolshoy ekran s vysokim razresheniem dlya pogruzhayuschego opyta.', 2.0);

INSERT INTO seat_types (name, description, price_modifier) VALUES
('Standart', 'Obychnye mesta s komfortnoy posadkoy.', 1),
('Premium', 'Mesta s uvelichennym prostranstvom dlya nog i luchshim komfortom.', 1.5),
('VIP', 'Mesta v otdelnom zale s povyshennym urovnem servisa i udobstvom.', 2.0),
('Detskoe', 'Mesta, prednaznachennye dlya detey, s bezopasnymi i udobnymi sidenyami.', 0.8),
('Lyuks', 'Mesta s maksimalnym komfortom, vklyuchaya vozmozhnost zakaza edy i napitkov.', 2.5);

INSERT INTO halls (screen_type_id, name, description) VALUES
((SELECT id FROM screen_types WHERE name = '2D'), 'Hall 1', 'Hall 2'),
((SELECT id FROM screen_types WHERE name = '3D'), 'Hall 3', 'Hall 4'),
((SELECT id FROM screen_types WHERE name = 'IMAX'), 'Hall 5', 'Hall 6');

DO $$
DECLARE
    hall_id UUID := (SELECT id FROM halls WHERE name = 'Hall 1');
    row_num INTEGER;
    seat_num INTEGER;
BEGIN
    FOR row_num IN 1..4 LOOP
        FOR seat_num IN 1..7 LOOP
            INSERT INTO seats (hall_id, seat_type_id, row_number, seat_number) VALUES
            (hall_id, (SELECT id FROM seat_types WHERE name = 'VIP'), row_num, seat_num);
        END LOOP;
    END LOOP;
END $$;

DO $$
DECLARE
    hall_id UUID := (SELECT id FROM halls WHERE name = 'Hall 3');
    row_num INTEGER;
    seat_num INTEGER;
BEGIN
    FOR row_num IN 1..10 LOOP
        FOR seat_num IN 1..10 LOOP
            INSERT INTO seats (hall_id, seat_type_id, row_number, seat_number) VALUES
            (hall_id, (SELECT id FROM seat_types WHERE name = 'Standart'), row_num, seat_num);
        END LOOP;
    END LOOP;
END $$;

DO $$
DECLARE
    hall_id UUID := (SELECT id FROM halls WHERE name = 'Hall 5');
    row_num INTEGER;
    seat_num INTEGER;
BEGIN
    FOR row_num IN 1..7 LOOP
        FOR seat_num IN 1..10 LOOP
            INSERT INTO seats (hall_id, seat_type_id, row_number, seat_number) VALUES
            (hall_id, (SELECT id FROM seat_types WHERE name = 'Premium'), row_num, seat_num);
        END LOOP;
    END LOOP;
END $$;

INSERT INTO movies (title, duration, description, age_limit, box_office_revenue, release_date) VALUES
('Zvezdnye voyny: Epizod IV - Novaya nadezhda', '02:01:00', 'Genre 6', 12, 0, '1977-05-25'),
('Titanik', '03:14:00', 'Movie 1', 12, 0, '1997-12-19'),
('Vlastelin kolets: Bratstvo koltsa', '02:58:00', 'Sample English text 3', 12, 0, '2001-12-19'),
('Piraty Karibskogo morya: Proklyatie chernoy zhemchuzhiny', '02:23:00', 'Genre 7', 12, 0, '2003-07-09'),
('Temnyy rytsar', '02:32:00', 'Film o supergeroe Betmene i ego protivostoyanii s Dzhokerom.', 16, 0, '2008-07-18'),
('Nachalo', '02:28:00', 'Genre 8', 12, 0, '2010-07-16'),
('Garri Potter i filosofskiy kamen', '02:32:00', 'Movie 2', 6, 0, '2001-11-16'),
('Mstiteli: Final', '03:01:00', 'Zaklyuchitelnaya chast sagi o Mstitelyakh i ikh borbe s Tanosom.', 12, 0, '2019-04-26'),
('Krestnyy otets', '02:55:00', 'Genre 9', 18, 0, '1972-03-24'),
('Forrest Gamp', '02:22:00', 'Istoriya zhizni cheloveka s nizkim IQ, kotoryy stal svidetelem istoricheskikh sobytiy.', 12, 0, '1994-07-06');

INSERT INTO reviews (user_id, movie_id, rating, review_comment) VALUES
((SELECT id FROM users WHERE email = 'ivan@example.com'), (SELECT id FROM movies WHERE title = 'Zvezdnye voyny: Epizod IV - Novaya nadezhda'), 10, 'Message'),
((SELECT id FROM users WHERE email = 'maria@example.com'), (SELECT id FROM movies WHERE title = 'Titanik'), 9, 'Sample English text 4'),
((SELECT id FROM users WHERE email = 'alexey@example.com'), (SELECT id FROM movies WHERE title = 'Vlastelin kolets: Bratstvo koltsa'), 10, 'Sample English text 5'),
((SELECT id FROM users WHERE email = 'olga@example.com'), (SELECT id FROM movies WHERE title = 'Piraty Karibskogo morya: Proklyatie chernoy zhemchuzhiny'), 8, 'Genre 10'),
((SELECT id FROM users WHERE email = 'ivan@example.com'), (SELECT id FROM movies WHERE title = 'Temnyy rytsar'), 10, 'Luchshiy film o supergeroyakh! Dzhoker v ispolnenii Khita Ledzhera prosto potryasayuschiy.'),
((SELECT id FROM users WHERE email = 'maria@example.com'), (SELECT id FROM movies WHERE title = 'Nachalo'), 9, 'Umopomrachitelnyy film s zaputannym syuzhetom. Nuzhno smotret neskolko raz, chtoby ponyat vse detali.'),
((SELECT id FROM users WHERE email = 'alexey@example.com'), (SELECT id FROM movies WHERE title = 'Garri Potter i filosofskiy kamen'), 8, 'Prekrasnyy film dlya vsey semi. Volshebnyy mir, kotoryy zavorazhivaet.'),
((SELECT id FROM users WHERE email = 'olga@example.com'), (SELECT id FROM movies WHERE title = 'Mstiteli: Final'), 10, 'Sample English text 6'),
((SELECT id FROM users WHERE email = 'ivan@example.com'), (SELECT id FROM movies WHERE title = 'Krestnyy otets'), 10, 'Sample English text 7'),
((SELECT id FROM users WHERE email = 'maria@example.com'), (SELECT id FROM movies WHERE title = 'Forrest Gamp'), 9, 'Sample English text 8');

INSERT INTO movie_shows (movie_id, hall_id, start_time, language) VALUES
((SELECT id FROM movies WHERE title = 'Zvezdnye voyny: Epizod IV - Novaya nadezhda'), (SELECT id FROM halls WHERE name = 'Hall 1'), '2025-05-21 18:00:00', 'English'),
((SELECT id FROM movies WHERE title = 'Titanik'), (SELECT id FROM halls WHERE name = 'Hall 3'), '2025-05-21 20:30:00', 'English'),
((SELECT id FROM movies WHERE title = 'Vlastelin kolets: Bratstvo koltsa'), (SELECT id FROM halls WHERE name = 'Hall 5'), '2025-05-22 19:00:00', 'English'),
((SELECT id FROM movies WHERE title = 'Piraty Karibskogo morya: Proklyatie chernoy zhemchuzhiny'), (SELECT id FROM halls WHERE name = 'Hall 1'), '2025-05-22 21:30:00', 'English'),
((SELECT id FROM movies WHERE title = 'Temnyy rytsar'), (SELECT id FROM halls WHERE name = 'Hall 3'), '2025-05-23 18:00:00', 'English'),
((SELECT id FROM movies WHERE title = 'Nachalo'), (SELECT id FROM halls WHERE name = 'Hall 5'), '2025-05-23 20:30:00', 'English'),
((SELECT id FROM movies WHERE title = 'Garri Potter i filosofskiy kamen'), (SELECT id FROM halls WHERE name = 'Hall 1'), '2025-05-24 16:00:00', 'English'),
((SELECT id FROM movies WHERE title = 'Mstiteli: Final'), (SELECT id FROM halls WHERE name = 'Hall 3'), '2025-05-24 19:00:00', 'English'),
((SELECT id FROM movies WHERE title = 'Krestnyy otets'), (SELECT id FROM halls WHERE name = 'Hall 5'), '2025-05-25 18:00:00', 'English'),
((SELECT id FROM movies WHERE title = 'Forrest Gamp'), (SELECT id FROM halls WHERE name = 'Hall 1'), '2025-05-25 20:30:00', 'English');

INSERT INTO movies_genres (movie_id, genre_id) VALUES
((SELECT id FROM movies WHERE title = 'Zvezdnye voyny: Epizod IV - Novaya nadezhda'), (SELECT id FROM genres WHERE name = 'Nauchnaya fantastika')),
((SELECT id FROM movies WHERE title = 'Zvezdnye voyny: Epizod IV - Novaya nadezhda'), (SELECT id FROM genres WHERE name = 'Genre 2')),
((SELECT id FROM movies WHERE title = 'Titanik'), (SELECT id FROM genres WHERE name = 'Genre 1')),
((SELECT id FROM movies WHERE title = 'Titanik'), (SELECT id FROM genres WHERE name = 'Romantika')),
((SELECT id FROM movies WHERE title = 'Vlastelin kolets: Bratstvo koltsa'), (SELECT id FROM genres WHERE name = 'Genre 3')),
((SELECT id FROM movies WHERE title = 'Vlastelin kolets: Bratstvo koltsa'), (SELECT id FROM genres WHERE name = 'Genre 4')),
((SELECT id FROM movies WHERE title = 'Piraty Karibskogo morya: Proklyatie chernoy zhemchuzhiny'), (SELECT id FROM genres WHERE name = 'Genre 4')),
((SELECT id FROM movies WHERE title = 'Piraty Karibskogo morya: Proklyatie chernoy zhemchuzhiny'), (SELECT id FROM genres WHERE name = 'Komediya')),
((SELECT id FROM movies WHERE title = 'Temnyy rytsar'), (SELECT id FROM genres WHERE name = 'Genre 2')),
((SELECT id FROM movies WHERE title = 'Nachalo'), (SELECT id FROM genres WHERE name = 'Nauchnaya fantastika')),
((SELECT id FROM movies WHERE title = 'Garri Potter i filosofskiy kamen'), (SELECT id FROM genres WHERE name = 'Genre 3')),
((SELECT id FROM movies WHERE title = 'Garri Potter i filosofskiy kamen'), (SELECT id FROM genres WHERE name = 'Genre 4')),
((SELECT id FROM movies WHERE title = 'Mstiteli: Final'), (SELECT id FROM genres WHERE name = 'Genre 2')),
((SELECT id FROM movies WHERE title = 'Krestnyy otets'), (SELECT id FROM genres WHERE name = 'Genre 1')),
((SELECT id FROM movies WHERE title = 'Forrest Gamp'), (SELECT id FROM genres WHERE name = 'Genre 1')),
((SELECT id FROM movies WHERE title = 'Forrest Gamp'), (SELECT id FROM genres WHERE name = 'Komediya'));

DO $$
DECLARE
    show_id UUID;
    seat_id UUID;
    price_modifier_seat DECIMAL;
    price_modifier_screen DECIMAL;
    ticket_price DECIMAL;
BEGIN
    FOR show_id IN SELECT id FROM movie_shows LOOP
        FOR seat_id IN SELECT id FROM seats LOOP
            SELECT price_modifier INTO price_modifier_seat
            FROM seat_types
            WHERE id = (SELECT seat_type_id FROM seats WHERE id = seat_id);
            
            SELECT price_modifier INTO price_modifier_screen
            FROM screen_types
            WHERE id = (SELECT screen_type_id FROM halls WHERE id = (SELECT hall_id FROM movie_shows WHERE id = show_id));
            
            ticket_price := 300 * price_modifier_seat * price_modifier_screen;
            
            INSERT INTO tickets (movie_show_id, seat_id, user_id, ticket_Status, price)
            VALUES (show_id, seat_id, NULL, 'Available', ticket_price);
        END LOOP;
    END LOOP;
END $$;

