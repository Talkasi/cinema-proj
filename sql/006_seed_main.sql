-- Вставка пользователей
INSERT INTO users (name, email, password_hash, birth_date, is_admin) VALUES
('Иван Иванов', 'ivan@example.com', 'hashed_password_1', '1990-01-01', FALSE),
('Мария Key', 'maria@example.com', 'hashed_password_2', '1985-05-15', FALSE),
('smirnov532', 'alexey@example.com', 'hashed_password_3', '1978-10-20', TRUE),
('Ольга Кузнецова', 'olga@example.com', 'hashed_password_4', '1995-03-30', FALSE);

-- Вставка жанров
INSERT INTO genres (name, description) VALUES
('Драма', 'Фильмы, которые исследуют человеческие эмоции и конфликты.'),
('Комедия', 'Фильмы, которые предназначены для развлечения и смеха.'),
('Экшн', 'Фильмы с высоким уровнем напряжения и динамичными сценами.'),
('Ужасы', 'Фильмы, которые вызывают страх и напряжение.'),
('Научная фантастика', 'Фильмы, основанные на научных концепциях и технологиях.'),
('Фэнтези', 'Фильмы, основанные на воображаемых мирах и магии.'),
('Приключения', 'Фильмы, которые фокусируются на захватывающих путешествиях и приключениях.'),
('Романтика', 'Фильмы, которые исследуют любовные отношения между персонажами.'),
('Семейный', 'Фильмы, подходящие для просмотра всей семьей, часто с позитивными и вдохновляющими темами.'),
('Научный', 'Фильмы, которые исследуют научные концепции и идеи.');

-- Вставка типов экранов
INSERT INTO screen_types (name, description, price_modifier) VALUES
('2D', 'Стандартный экран для показа фильмов в 2D.', 1.0),
('3D', 'Экран для показа фильмов в 3D с использованием специальных очков.', 1.5),
('IMAX', 'Большой экран с высоким разрешением для погружающего опыта.', 2.0);

-- Вставка типов мест
INSERT INTO seat_types (name, description, price_modifier) VALUES
('Стандарт', 'Обычные места с комфортной посадкой.', 1),
('Премиум', 'Места с увеличенным пространством для ног и лучшим комфортом.', 1.5),
('VIP', 'Места в отдельном зале с повышенным уровнем сервиса и удобством.', 2.0),
('Детское', 'Места, предназначенные для детей, с безопасными и удобными сиденьями.', 0.8),
('Люкс', 'Места с максимальным комфортом, включая возможность заказа еды и напитков.', 2.5);

-- Вставка залов
INSERT INTO halls (screen_type_id, name, description) VALUES
((SELECT id FROM screen_types WHERE name = '2D'), 'Зал 1', 'Небольшой зал с VIP местами.'),
((SELECT id FROM screen_types WHERE name = '3D'), 'Зал 2', 'Зал для показа 3D фильмов.'),
((SELECT id FROM screen_types WHERE name = 'IMAX'), 'Зал 3', 'Зал с IMAX экраном.');

-- Вставка мест
-- Вставка мест в зал 1
DO $$
DECLARE
    hall_id UUID := (SELECT id FROM halls WHERE name = 'Зал 1');
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

-- Вставка мест в зал 2
DO $$
DECLARE
    hall_id UUID := (SELECT id FROM halls WHERE name = 'Зал 2');
    row_num INTEGER;
    seat_num INTEGER;
BEGIN
    FOR row_num IN 1..10 LOOP
        FOR seat_num IN 1..10 LOOP
            INSERT INTO seats (hall_id, seat_type_id, row_number, seat_number) VALUES
            (hall_id, (SELECT id FROM seat_types WHERE name = 'Стандарт'), row_num, seat_num);
        END LOOP;
    END LOOP;
END $$;

-- Вставка мест в зал 3
DO $$
DECLARE
    hall_id UUID := (SELECT id FROM halls WHERE name = 'Зал 3');
    row_num INTEGER;
    seat_num INTEGER;
BEGIN
    FOR row_num IN 1..7 LOOP
        FOR seat_num IN 1..10 LOOP
            INSERT INTO seats (hall_id, seat_type_id, row_number, seat_number) VALUES
            (hall_id, (SELECT id FROM seat_types WHERE name = 'Премиум'), row_num, seat_num);
        END LOOP;
    END LOOP;
END $$;

-- Вставка фильмов
INSERT INTO movies (title, duration, description, age_limit, box_office_revenue, release_date) VALUES
('Звёздные войны: Эпизод IV - Новая надежда', '02:01:00', 'Классический научно-фантастический фильм о борьбе между Империей и Повстанцами.', 12, 0, '1977-05-25'),
('Титаник', '03:14:00', 'Романтическая драма о любви на фоне катастрофы Титаника.', 12, 0, '1997-12-19'),
('Властелин колец: Братство кольца', '02:58:00', 'Эпическая история о борьбе за уничтожение кольца власти.', 12, 0, '2001-12-19'),
('Пираты Карибского моря: Проклятие черной жемчужины', '02:23:00', 'Приключенческий фильм о пиратах и их поисках сокровищ.', 12, 0, '2003-07-09'),
('Темный рыцарь', '02:32:00', 'Фильм о супергерое Бэтмене и его противостоянии с Джокером.', 16, 0, '2008-07-18'),
('Начало', '02:28:00', 'Научно-фантастический триллер о мире снов и манипуляциях с сознанием.', 12, 0, '2010-07-16'),
('Гарри Поттер и философский камень', '02:32:00', 'Первый фильм о приключениях молодого волшебника Гарри Поттера.', 6, 0, '2001-11-16'),
('Мстители: Финал', '03:01:00', 'Заключительная часть саги о Мстителях и их борьбе с Таносом.', 12, 0, '2019-04-26'),
('Крестный отец', '02:55:00', 'Криминальная драма о семье Корлеоне и их бизнесе.', 18, 0, '1972-03-24'),
('Форрест Гамп', '02:22:00', 'История жизни человека с низким IQ, который стал свидетелем исторических событий.', 12, 0, '1994-07-06');

-- Вставка отзывов
INSERT INTO reviews (user_id, movie_id, rating, review_comment) VALUES
((SELECT id FROM users WHERE email = 'ivan@example.com'), (SELECT id FROM movies WHERE title = 'Звёздные войны: Эпизод IV - Новая надежда'), 10, 'Невероятный фильм! Классика жанра, которая вдохновила целое поколение.'),
((SELECT id FROM users WHERE email = 'maria@example.com'), (SELECT id FROM movies WHERE title = 'Титаник'), 9, 'Очень трогательная история любви. Грустный, но красивый фильм.'),
((SELECT id FROM users WHERE email = 'alexey@example.com'), (SELECT id FROM movies WHERE title = 'Властелин колец: Братство кольца'), 10, 'Эпическая сага, которая захватывает с первых минут. Великолепная работа!'),
((SELECT id FROM users WHERE email = 'olga@example.com'), (SELECT id FROM movies WHERE title = 'Пираты Карибского моря: Проклятие черной жемчужины'), 8, 'Забавный и увлекательный фильм с отличным юмором и приключениями.'),
((SELECT id FROM users WHERE email = 'ivan@example.com'), (SELECT id FROM movies WHERE title = 'Темный рыцарь'), 10, 'Лучший фильм о супергероях! Джокер в исполнении Хита Леджера просто потрясающий.'),
((SELECT id FROM users WHERE email = 'maria@example.com'), (SELECT id FROM movies WHERE title = 'Начало'), 9, 'Умопомрачительный фильм с запутанным сюжетом. Нужно смотреть несколько раз, чтобы понять все детали.'),
((SELECT id FROM users WHERE email = 'alexey@example.com'), (SELECT id FROM movies WHERE title = 'Гарри Поттер и философский камень'), 8, 'Прекрасный фильм для всей семьи. Волшебный мир, который завораживает.'),
((SELECT id FROM users WHERE email = 'olga@example.com'), (SELECT id FROM movies WHERE title = 'Мстители: Финал'), 10, 'Эпичное завершение саги о Мстителях. Все персонажи на своих местах, и финал просто потрясающий!'),
((SELECT id FROM users WHERE email = 'ivan@example.com'), (SELECT id FROM movies WHERE title = 'Крестный отец'), 10, 'Шедевр! Один из лучших фильмов всех времен. Великолепная игра актеров и сценарий.'),
((SELECT id FROM users WHERE email = 'maria@example.com'), (SELECT id FROM movies WHERE title = 'Форрест Гамп'), 9, 'Трогательная история о жизни и судьбе. Очень вдохновляющий фильм.');

-- Вставка сеансов
INSERT INTO movie_shows (movie_id, hall_id, start_time, language) VALUES
((SELECT id FROM movies WHERE title = 'Звёздные войны: Эпизод IV - Новая надежда'), (SELECT id FROM halls WHERE name = 'Зал 1'), '2025-05-21 18:00:00', 'English'),
((SELECT id FROM movies WHERE title = 'Титаник'), (SELECT id FROM halls WHERE name = 'Зал 2'), '2025-05-21 20:30:00', 'English'),
((SELECT id FROM movies WHERE title = 'Властелин колец: Братство кольца'), (SELECT id FROM halls WHERE name = 'Зал 3'), '2025-05-22 19:00:00', 'English'),
((SELECT id FROM movies WHERE title = 'Пираты Карибского моря: Проклятие черной жемчужины'), (SELECT id FROM halls WHERE name = 'Зал 1'), '2025-05-22 21:30:00', 'English'),
((SELECT id FROM movies WHERE title = 'Темный рыцарь'), (SELECT id FROM halls WHERE name = 'Зал 2'), '2025-05-23 18:00:00', 'English'),
((SELECT id FROM movies WHERE title = 'Начало'), (SELECT id FROM halls WHERE name = 'Зал 3'), '2025-05-23 20:30:00', 'English'),
((SELECT id FROM movies WHERE title = 'Гарри Поттер и философский камень'), (SELECT id FROM halls WHERE name = 'Зал 1'), '2025-05-24 16:00:00', 'English'),
((SELECT id FROM movies WHERE title = 'Мстители: Финал'), (SELECT id FROM halls WHERE name = 'Зал 2'), '2025-05-24 19:00:00', 'English'),
((SELECT id FROM movies WHERE title = 'Крестный отец'), (SELECT id FROM halls WHERE name = 'Зал 3'), '2025-05-25 18:00:00', 'English'),
((SELECT id FROM movies WHERE title = 'Форрест Гамп'), (SELECT id FROM halls WHERE name = 'Зал 1'), '2025-05-25 20:30:00', 'English');

-- Вставка данных о связи фильмов и жанров
INSERT INTO movies_genres (movie_id, genre_id) VALUES
((SELECT id FROM movies WHERE title = 'Звёздные войны: Эпизод IV - Новая надежда'), (SELECT id FROM genres WHERE name = 'Научная фантастика')),
((SELECT id FROM movies WHERE title = 'Звёздные войны: Эпизод IV - Новая надежда'), (SELECT id FROM genres WHERE name = 'Экшн')),
((SELECT id FROM movies WHERE title = 'Титаник'), (SELECT id FROM genres WHERE name = 'Драма')),
((SELECT id FROM movies WHERE title = 'Титаник'), (SELECT id FROM genres WHERE name = 'Романтика')),
((SELECT id FROM movies WHERE title = 'Властелин колец: Братство кольца'), (SELECT id FROM genres WHERE name = 'Фэнтези')),
((SELECT id FROM movies WHERE title = 'Властелин колец: Братство кольца'), (SELECT id FROM genres WHERE name = 'Приключения')),
((SELECT id FROM movies WHERE title = 'Пираты Карибского моря: Проклятие черной жемчужины'), (SELECT id FROM genres WHERE name = 'Приключения')),
((SELECT id FROM movies WHERE title = 'Пираты Карибского моря: Проклятие черной жемчужины'), (SELECT id FROM genres WHERE name = 'Комедия')),
((SELECT id FROM movies WHERE title = 'Темный рыцарь'), (SELECT id FROM genres WHERE name = 'Экшн')),
((SELECT id FROM movies WHERE title = 'Начало'), (SELECT id FROM genres WHERE name = 'Научная фантастика')),
((SELECT id FROM movies WHERE title = 'Гарри Поттер и философский камень'), (SELECT id FROM genres WHERE name = 'Фэнтези')),
((SELECT id FROM movies WHERE title = 'Гарри Поттер и философский камень'), (SELECT id FROM genres WHERE name = 'Приключения')),
((SELECT id FROM movies WHERE title = 'Мстители: Финал'), (SELECT id FROM genres WHERE name = 'Экшн')),
((SELECT id FROM movies WHERE title = 'Крестный отец'), (SELECT id FROM genres WHERE name = 'Драма')),
((SELECT id FROM movies WHERE title = 'Форрест Гамп'), (SELECT id FROM genres WHERE name = 'Драма')),
((SELECT id FROM movies WHERE title = 'Форрест Гамп'), (SELECT id FROM genres WHERE name = 'Комедия'));

-- Вставка данных о билетах
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
            
            INSERT INTO tickets (movie_show_id, seat_id, user_id, ticket_status, price)
            VALUES (show_id, seat_id, NULL, 'Available', ticket_price);
        END LOOP;
    END LOOP;
END $$;

