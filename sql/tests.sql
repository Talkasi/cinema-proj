SET ROLE cinema_admin;

DO $$
DECLARE
    new_user_id UUID;
    show_id UUID;
    ticket_id UUID;
    ticket_price DECIMAL(10,2);
    initial_revenue DECIMAL(10,2);
    new_revenue DECIMAL(10,2);
BEGIN
    SELECT id INTO show_id FROM movie_shows LIMIT 1;
    SELECT id INTO new_user_id FROM users LIMIT 1;
    SELECT id, price INTO ticket_id, ticket_price FROM tickets WHERE movie_show_id = show_id LIMIT 1;
    
    -- Запоминаем начальные сборы
    SELECT box_office_revenue INTO initial_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    
    -- Тест: Available → Purchased
    UPDATE tickets SET ticket_status = 'Purchased'::ticket_status_enum, user_id = new_user_id WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Тест 1: Available → Purchased, Revenue изменился с % на % (ожидалось +%)', initial_revenue, new_revenue, ticket_price;

    -- Запоминаем новый доход
    initial_revenue := new_revenue;

    -- Тест: Purchased → Available
    UPDATE tickets SET ticket_status = 'Available'::ticket_status_enum, user_id = NULL WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Тест 2: Purchased → Available, Revenue изменился с % на % (ожидалось -%)', initial_revenue, new_revenue, ticket_price;

    -- Запоминаем новый доход
    initial_revenue := new_revenue;

    -- Тест: Available → Reserved
    UPDATE tickets SET ticket_status = 'Reserved'::ticket_status_enum, user_id = new_user_id WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Тест 3: Available → Reserved, Revenue изменился с % на % (ожидалось 0)', initial_revenue, new_revenue;

    -- Запоминаем новый доход
    initial_revenue := new_revenue;

    -- Тест: Reserved → Purchased
    UPDATE tickets SET ticket_status = 'Purchased'::ticket_status_enum, user_id = new_user_id WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Тест 4: Reserved → Purchased, Revenue изменился с % на % (ожидалось +%)', initial_revenue, new_revenue, ticket_price;

    -- Запоминаем новый доход
    initial_revenue := new_revenue;

    -- Тест: Purchased → Reserved
    UPDATE tickets SET ticket_status = 'Reserved'::ticket_status_enum, user_id = new_user_id WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Тест 5: Purchased → Reserved, Revenue изменился с % на % (ожидалось -%)', initial_revenue, new_revenue, ticket_price;

    -- Запоминаем новый доход
    initial_revenue := new_revenue;

    -- Тест: Reserved → Available
    UPDATE tickets SET ticket_status = 'Available'::ticket_status_enum, user_id = NULL WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Тест 6: Reserved → Available, Revenue изменился с % на % (ожидалось 0)', initial_revenue, new_revenue;

	-- Запоминаем новый доход
    initial_revenue := new_revenue;

    -- Тест: price + 100
    UPDATE tickets SET price = price + 100 WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Тест 7: price + 100, Revenue изменился с % на % (ожидалось 0)', initial_revenue, new_revenue;

END $$;

RESET ROLE;
SET ROLE cinema_user;

DO $$
DECLARE
    new_user_id UUID;
    show_id UUID;
    ticket_id UUID;
    ticket_price DECIMAL(10,2);
    initial_revenue DECIMAL(10,2);
    new_revenue DECIMAL(10,2);
BEGIN
    SELECT id INTO show_id FROM movie_shows LIMIT 1;
    SELECT id INTO new_user_id FROM users LIMIT 1;
    SELECT id, price INTO ticket_id, ticket_price FROM tickets WHERE movie_show_id = show_id LIMIT 1;
    
    -- Запоминаем начальные сборы
    SELECT box_office_revenue INTO initial_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);

    -- Тест: Available → Reserved
    UPDATE tickets SET ticket_status = 'Reserved'::ticket_status_enum, user_id = new_user_id WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Тест 1: Available → Reserved, Revenue изменился с % на % (ожидалось 0)', initial_revenue, new_revenue;

    -- Запоминаем новый доход
    initial_revenue := new_revenue;

    -- Тест: Purchased → Reserved
    UPDATE tickets SET ticket_status = 'Reserved'::ticket_status_enum, user_id = new_user_id WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Тест 2: Purchased → Reserved, Revenue изменился с % на % (ожидалось 0)', initial_revenue, new_revenue;

	-- Запоминаем новый доход
    initial_revenue := new_revenue;

    -- Тест: price + 100
    UPDATE tickets SET price = price + 100 WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Тест 3: price + 100, Revenue изменился с % на % (ожидалось 0)', initial_revenue, new_revenue;
END $$;

SET ROLE cinema_admin;

DO $$
DECLARE
    v_hall_id UUID;
    v_movie_id UUID;
    v_existing_show_id UUID;
    v_new_show_id UUID;
    v_cleanup_time INTERVAL := INTERVAL '10 minutes';
    v_show_duration INTERVAL := INTERVAL '120 minutes';
    v_test_start_time TIMESTAMP;
    v_test_end_time TIMESTAMP;
    v_screen_type_id UUID;
    v_seat_type_id UUID;
    v_base_user_id UUID;
BEGIN
    -- Создаем тестовые данные
    INSERT INTO screen_types (name, description, price_modifier)
    VALUES ('Standard', 'Standard screen', 1.0)
    RETURNING id INTO v_screen_type_id;
    
    INSERT INTO halls (screen_type_id, name, description)
    VALUES (v_screen_type_id, 'Test Hall', 'Test Hall Description')
    RETURNING id INTO v_hall_id;
    
    INSERT INTO seat_types (name, description, price_modifier)
    VALUES ('Standard', 'Standard seat', 1.0)
    RETURNING id INTO v_seat_type_id;
    
    -- Добавляем тестовые места
    INSERT INTO seats (hall_id, seat_type_id, row_number, seat_number)
    VALUES 
        (v_hall_id, v_seat_type_id, 1, 1),
        (v_hall_id, v_seat_type_id, 1, 2);
    
    INSERT INTO users (name, email, password_hash, birth_date)
    VALUES ('Test User', 'test@example.com', 'hash', '1990-01-01')
    RETURNING id INTO v_base_user_id;
    
    INSERT INTO movies (title, duration, description, age_limit, release_date)
    VALUES ('Test Movie', '02:00:00', 'Test Description', 12, '2023-01-01')
    RETURNING id INTO v_movie_id;

    -- Создаем тестовый сеанс и сохраняем его ID
    SELECT create_movie_show_with_tickets(
        v_movie_id, 
        v_hall_id, 
         '2023-01-01:12:12:12'::timestamp, 
        'English'::language_enum, 
        300.00
    ) INTO v_existing_show_id;
    
    RAISE NOTICE 'Создан тестовый сеанс с ID: %', v_existing_show_id;
    
    RAISE NOTICE '=== Тестирование граничных условий ===';
    
    -- 1. Начало сразу после уборки (успех)
    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_existing_show_id) + 
                          (SELECT duration FROM movies WHERE id = v_movie_id) + v_cleanup_time;
        
        INSERT INTO movie_shows (movie_id, hall_id, start_time, language)
        VALUES (v_movie_id, v_hall_id, v_test_start_time, 'English'::language_enum)
        RETURNING id INTO v_new_show_id;
        
        RAISE NOTICE 'Тест 1: Начало сразу после уборки - УСПЕХ (сеанс %)', v_new_show_id;
        DELETE FROM movie_shows WHERE id = v_new_show_id;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 1: Начало сразу после уборки - ОШИБКА: %', SQLERRM;
    END;
    
    -- 2. Конец уборки = начало другого сеанса (успех)
    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_existing_show_id) + 
                          (SELECT duration FROM movies WHERE id = v_movie_id) + v_cleanup_time;
        
        INSERT INTO movie_shows (movie_id, hall_id, start_time, language)
        VALUES (v_movie_id, v_hall_id, v_test_start_time, 'English'::language_enum)
        RETURNING id INTO v_new_show_id;
        
        RAISE NOTICE 'Тест 2: Конец уборки = начало другого - УСПЕХ (сеанс %)', v_new_show_id;
        DELETE FROM movie_shows WHERE id = v_new_show_id;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 2: Конец уборки = начало другого - ОШИБКА: %', SQLERRM;
    END;
    
    RAISE NOTICE '=== Тестирование конфликтных сценариев ===';
    
    -- 3. Полное вложение интервала (ошибка)
    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_existing_show_id) + INTERVAL '30 minutes';
        
        INSERT INTO movie_shows (movie_id, hall_id, start_time, language)
        VALUES (v_movie_id, v_hall_id, v_test_start_time, 'English'::language_enum);
        
        RAISE NOTICE 'Тест 3: Полное вложение интервала - НЕПРОЙДЕН (ожидалась ошибка)';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 3: Полное вложение интервала - ОШИБКА (ожидаемо): %', SQLERRM;
    END;
    
    -- 4. Частичное перекрытие (левый край) (ошибка)
    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_existing_show_id) - INTERVAL '30 minutes';
        
        INSERT INTO movie_shows (movie_id, hall_id, start_time, language)
        VALUES (v_movie_id, v_hall_id, v_test_start_time, 'English'::language_enum);
        
        RAISE NOTICE 'Тест 4: Частичное перекрытие (левый край) - НЕПРОЙДЕН (ожидалась ошибка)';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 4: Частичное перекрытие (левый край) - ОШИБКА (ожидаемо): %', SQLERRM;
    END;
    
    -- 5. Частичное перекрытие (правый край) (ошибка)
    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_existing_show_id) + 
                         (SELECT duration FROM movies WHERE id = v_movie_id) - INTERVAL '30 minutes';
        
        INSERT INTO movie_shows (movie_id, hall_id, start_time, language)
        VALUES (v_movie_id, v_hall_id, v_test_start_time, 'English'::language_enum);
        
        RAISE NOTICE 'Тест 5: Частичное перекрытие (правый край) - НЕПРОЙДЕН (ожидалась ошибка)';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 5: Частичное перекрытие (правый край) - ОШИБКА (ожидаемо): %', SQLERRM;
    END;
    
    -- 6. Полное совпадение интервалов (ошибка)
    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_existing_show_id);
        
        INSERT INTO movie_shows (movie_id, hall_id, start_time, language)
        VALUES (v_movie_id, v_hall_id, v_test_start_time, 'English'::language_enum);
        
        RAISE NOTICE 'Тест 6: Полное совпадение интервалов - НЕПРОЙДЕН (ожидалась ошибка)';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 6: Полное совпадение интервалов - ОШИБКА (ожидаемо): %', SQLERRM;
    END;
    
    -- 7. Полное перекрытие интервалов (ошибка)
    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_existing_show_id) - INTERVAL '30 minutes';
        
        INSERT INTO movie_shows (movie_id, hall_id, start_time, language)
        VALUES (v_movie_id, v_hall_id, v_test_start_time, 'English'::language_enum);
        
        RAISE NOTICE 'Тест 7: Полное перекрытие интервалов - НЕПРОЙДЕН (ожидалась ошибка)';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 7: Полное перекрытие интервалов - ОШИБКА (ожидаемо): %', SQLERRM;
    END;
    
    RAISE NOTICE '=== Тестирование неконфликтных сценариев ===';
    
    -- 8. Раздельные интервалы (успех)
    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_existing_show_id) + 
                          (SELECT duration FROM movies WHERE id = v_movie_id) + v_cleanup_time + INTERVAL '1 hour';
        
        INSERT INTO movie_shows (movie_id, hall_id, start_time, language)
        VALUES (v_movie_id, v_hall_id, v_test_start_time, 'English'::language_enum)
        RETURNING id INTO v_new_show_id;
        
        RAISE NOTICE 'Тест 8: Раздельные интервалы - УСПЕХ (сеанс %)', v_new_show_id;
        DELETE FROM movie_shows WHERE id = v_new_show_id;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 8: Раздельные интервалы - ОШИБКА: %', SQLERRM;
    END;
    
    -- 9. Изменение невременных атрибутов (успех)
    BEGIN
        UPDATE movie_shows SET language = 'Русский'::language_enum WHERE id = v_existing_show_id;
        RAISE NOTICE 'Тест 9: Изменение невременных атрибутов - УСПЕХ';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 9: Изменение невременных атрибутов - ОШИБКА: %', SQLERRM;
    END;
    
    -- Очистка
    DELETE FROM tickets WHERE movie_show_id = v_existing_show_id;
    DELETE FROM movie_shows WHERE id = v_existing_show_id;
    DELETE FROM seats WHERE hall_id = v_hall_id;
    DELETE FROM halls WHERE id = v_hall_id;
    DELETE FROM screen_types WHERE id = v_screen_type_id;
    DELETE FROM seat_types WHERE id = v_seat_type_id;
    DELETE FROM users WHERE id = v_base_user_id;
    DELETE FROM movies WHERE id = v_movie_id;
END $$;


DO $$
DECLARE
    v_hall_id UUID;
    v_movie_id UUID;
    v_base_show_id UUID;
    v_test_show_id UUID;
    v_cleanup_time INTERVAL := INTERVAL '10 minutes';
    v_show_duration INTERVAL := INTERVAL '120 minutes';
    v_test_start_time TIMESTAMP;
    v_screen_type_id UUID;
    v_seat_type_id UUID;
    v_base_user_id UUID;
BEGIN
    -- Создаем тестовые данные
    INSERT INTO screen_types (name, description, price_modifier)
    VALUES ('Standard', 'Standard screen', 1.0)
    RETURNING id INTO v_screen_type_id;
    
    INSERT INTO halls (screen_type_id, name, description)
    VALUES (v_screen_type_id, 'Test Hall', 'Test Hall Description')
    RETURNING id INTO v_hall_id;
    
    INSERT INTO seat_types (name, description, price_modifier)
    VALUES ('Standard', 'Standard seat', 1.0)
    RETURNING id INTO v_seat_type_id;
    
    -- Добавляем тестовые места
    INSERT INTO seats (hall_id, seat_type_id, row_number, seat_number)
    VALUES 
        (v_hall_id, v_seat_type_id, 1, 1),
        (v_hall_id, v_seat_type_id, 1, 2);
    
    INSERT INTO users (name, email, password_hash, birth_date)
    VALUES ('Test User', 'test@example.com', 'hash', '1990-01-01')
    RETURNING id INTO v_base_user_id;
    
    INSERT INTO movies (title, duration, description, age_limit, release_date)
    VALUES ('Test Movie', '02:00:00', 'Test Description', 12, '2023-01-01')
    RETURNING id INTO v_movie_id;

    -- Создаем базовый сеанс для тестирования
    SELECT create_movie_show_with_tickets(
        v_movie_id, 
        v_hall_id, 
        '2023-01-01:12:12:12'::timestamp, 
        'English'::language_enum, 
        300.00
    ) INTO v_base_show_id;
    
    -- Создаем тестовый сеанс, который будем обновлять
    SELECT create_movie_show_with_tickets(
        v_movie_id, 
        v_hall_id, 
        '2023-01-01:12:12:12'::timestamp + INTERVAL '5 hours', 
        'English'::language_enum, 
        300.00
    ) INTO v_test_show_id;
    
    RAISE NOTICE '=== ТЕСТИРОВАНИЕ ОБНОВЛЕНИЯ СЕАНСОВ ===';
    
    RAISE NOTICE '=== Граничные условия ===';
    
    -- 1. Обновление: начало сразу после уборки (успех)
    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_base_show_id) + 
                          (SELECT duration FROM movies WHERE id = v_movie_id) + v_cleanup_time;
        
        UPDATE movie_shows 
        SET start_time = v_test_start_time 
        WHERE id = v_test_show_id;
        
        RAISE NOTICE 'Тест 1: Обновление - начало сразу после уборки - УСПЕХ';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 1: Обновление - начало сразу после уборки - ОШИБКА: %', SQLERRM;
    END;
    
    -- 2. Обновление: конец уборки = начало другого сеанса (успех)
    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_base_show_id) + 
                          (SELECT duration FROM movies WHERE id = v_movie_id) + v_cleanup_time;
        
        UPDATE movie_shows 
        SET start_time = v_test_start_time 
        WHERE id = v_test_show_id;
        
        RAISE NOTICE 'Тест 2: Обновление - конец уборки = начало другого - УСПЕХ';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 2: Обновление - конец уборки = начало другого - ОШИБКА: %', SQLERRM;
    END;
    
    RAISE NOTICE '=== Конфликтные сценарии ===';
    
    -- 3. Обновление: полное вложение интервала (ошибка)
    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_base_show_id) + INTERVAL '30 minutes';
        
        UPDATE movie_shows 
        SET start_time = v_test_start_time 
        WHERE id = v_test_show_id;
        
        RAISE NOTICE 'Тест 3: Обновление - полное вложение интервала - НЕПРОЙДЕН (ожидалась ошибка)';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 3: Обновление - полное вложение интервала - ОШИБКА (ожидаемо): %', SQLERRM;
    END;
    
    -- 4. Обновление: частичное перекрытие (левый край) (ошибка)
    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_base_show_id) - INTERVAL '30 minutes';
        
        UPDATE movie_shows 
        SET start_time = v_test_start_time 
        WHERE id = v_test_show_id;
        
        RAISE NOTICE 'Тест 4: Обновление - частичное перекрытие (левый край) - НЕПРОЙДЕН (ожидалась ошибка)';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 4: Обновление - частичное перекрытие (левый край) - ОШИБКА (ожидаемо): %', SQLERRM;
    END;
    
    -- 5. Обновление: частичное перекрытие (правый край) (ошибка)
    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_base_show_id) + 
                         (SELECT duration FROM movies WHERE id = v_movie_id) - INTERVAL '30 minutes';
        
        UPDATE movie_shows 
        SET start_time = v_test_start_time 
        WHERE id = v_test_show_id;
        
        RAISE NOTICE 'Тест 5: Обновление - частичное перекрытие (правый край) - НЕПРОЙДЕН (ожидалась ошибка)';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 5: Обновление - частичное перекрытие (правый край) - ОШИБКА (ожидаемо): %', SQLERRM;
    END;
    
    -- 6. Обновление: полное совпадение интервалов (ошибка)
    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_base_show_id);
        
        UPDATE movie_shows 
        SET start_time = v_test_start_time 
        WHERE id = v_test_show_id;
        
        RAISE NOTICE 'Тест 6: Обновление - полное совпадение интервалов - НЕПРОЙДЕН (ожидалась ошибка)';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 6: Обновление - полное совпадение интервалов - ОШИБКА (ожидаемо): %', SQLERRM;
    END;
    
    -- 7. Обновление: полное перекрытие интервалов (ошибка)
    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_base_show_id) - INTERVAL '30 minutes';
        
        UPDATE movie_shows 
        SET start_time = v_test_start_time 
        WHERE id = v_test_show_id;
        
        RAISE NOTICE 'Тест 7: Обновление - полное перекрытие интервалов - НЕПРОЙДЕН (ожидалась ошибка)';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 7: Обновление - полное перекрытие интервалов - ОШИБКА (ожидаемо): %', SQLERRM;
    END;
    
    RAISE NOTICE '=== Неконфликтные сценарии ===';
    
    -- 8. Обновление: раздельные интервалы (успех)
    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_base_show_id) + 
                          (SELECT duration FROM movies WHERE id = v_movie_id) + v_cleanup_time + INTERVAL '1 hour';
        
        UPDATE movie_shows 
        SET start_time = v_test_start_time 
        WHERE id = v_test_show_id;
        
        RAISE NOTICE 'Тест 8: Обновление - раздельные интервалы - УСПЕХ';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 8: Обновление - раздельные интервалы - ОШИБКА: %', SQLERRM;
    END;
    
    -- 9. Обновление: изменение невременных атрибутов (успех)
    BEGIN
        UPDATE movie_shows 
        SET language = 'Русский'::language_enum 
        WHERE id = v_test_show_id;
        
        RAISE NOTICE 'Тест 9: Обновление - изменение невременных атрибутов - УСПЕХ';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 9: Обновление - изменение невременных атрибутов - ОШИБКА: %', SQLERRM;
    END;
    
    -- Очистка
    DELETE FROM tickets WHERE movie_show_id IN (v_base_show_id, v_test_show_id);
    DELETE FROM movie_shows WHERE id IN (v_base_show_id, v_test_show_id);
    DELETE FROM seats WHERE hall_id = v_hall_id;
    DELETE FROM halls WHERE id = v_hall_id;
    DELETE FROM screen_types WHERE id = v_screen_type_id;
    DELETE FROM seat_types WHERE id = v_seat_type_id;
    DELETE FROM users WHERE id = v_base_user_id;
    DELETE FROM movies WHERE id = v_movie_id;
END $$;

RESET ROLE;


SET ROLE cinema_admin;

DO $$
DECLARE
    v_movie_id UUID;
    v_hall_id UUID;
    v_show_id UUID;
    v_tickets_count INT;
    v_expected_price DECIMAL(10,2);
    v_actual_price DECIMAL(10,2);
    v_screen_modifier DECIMAL(5,2);
    v_seat_modifier DECIMAL(5,2);
BEGIN
    -- Подготовка тестовых данных
    INSERT INTO movies (title, duration, description, age_limit, release_date)
    VALUES ('Test Movie', '02:00:00', 'Test Description', 12, '2023-01-01')
    RETURNING id INTO v_movie_id;
    
    INSERT INTO screen_types (name, description, price_modifier)
    VALUES ('Standard', 'Standard screen', 1.2)
    RETURNING id INTO v_screen_modifier;
    
    INSERT INTO halls (screen_type_id, name, description)
    VALUES (v_screen_modifier, 'Test Hall', 'Test Hall Description')
    RETURNING id INTO v_hall_id;
    
    INSERT INTO seat_types (name, description, price_modifier)
    VALUES ('Standard', 'Standard seat', 1.0),
           ('VIP', 'VIP seat', 1.5)
    RETURNING id INTO v_seat_modifier;
    
    -- Тест 1: Зал с 10 одинаковыми местами (ряды != места в ряду)
    RAISE NOTICE '=== Тест 1: Зал с 10 одинаковыми местами (ряды != места в ряду) ===';
    
    -- Добавляем 10 мест (2 ряда по 5 мест)
    INSERT INTO seats (hall_id, seat_type_id, row_number, seat_number)
    SELECT v_hall_id, v_seat_modifier, 
           (n-1)/5 + 1, -- 2 ряда (1 и 2)
           (n-1)%5 + 1  -- 5 мест в ряду
    FROM generate_series(1, 10) AS n;
    
    -- Вызываем функцию
    SELECT create_movie_show_with_tickets(
        v_movie_id, 
        v_hall_id, 
        NOW() + INTERVAL '1 hour', 
        'Russian'::language_enum, 
        300.00
    ) INTO v_show_id;
    
    -- Проверяем результаты
    SELECT COUNT(*) INTO v_tickets_count FROM tickets WHERE movie_show_id = v_show_id;
    SELECT price INTO v_actual_price FROM tickets WHERE movie_show_id = v_show_id LIMIT 1;
    v_expected_price := ROUND(300.00 * 1.2 * 1.0, 2); -- base_price * screen_mod * seat_mod
    
    RAISE NOTICE 'ID сеанса: %', v_show_id;
    RAISE NOTICE 'Создано билетов: % (ожидалось 10)', v_tickets_count;
    RAISE NOTICE 'Цена билета: % (ожидалось %)', v_actual_price, v_expected_price;
    
    -- Очистка
    DELETE FROM tickets WHERE movie_show_id = v_show_id;
    DELETE FROM movie_shows WHERE id = v_show_id;
    DELETE FROM seats WHERE hall_id = v_hall_id;
    
    
    -- Тест 2: Зал с 25 одинаковыми местами (ряды == места в ряду)
    RAISE NOTICE '=== Тест 2: Зал с 25 одинаковыми местами (ряды == места в ряду) ===';
    
    -- Добавляем 25 мест (5x5)
    INSERT INTO seats (hall_id, seat_type_id, row_number, seat_number)
    SELECT v_hall_id, v_seat_modifier, 
           (n-1)/5 + 1, -- 5 рядов
           (n-1)%5 + 1   -- 5 мест в ряду
    FROM generate_series(1, 25) AS n;
    
    -- Вызываем функцию
    SELECT create_movie_show_with_tickets(
        v_movie_id, 
        v_hall_id, 
        NOW() + INTERVAL '2 hours', 
        'English'::language_enum, 
        350.00
    ) INTO v_show_id;
    
    -- Проверяем результаты
    SELECT COUNT(*) INTO v_tickets_count FROM tickets WHERE movie_show_id = v_show_id;
    RAISE NOTICE 'Создано билетов: % (ожидалось 25)', v_tickets_count;
    
    -- Очистка
    DELETE FROM tickets WHERE movie_show_id = v_show_id;
    DELETE FROM movie_shows WHERE id = v_show_id;
    DELETE FROM seats WHERE hall_id = v_hall_id;
    
    
    -- Тест 3: Зал с разными типами мест
    RAISE NOTICE '=== Тест 3: Зал с разными типами мест ===';
    
    -- Добавляем 10 мест (5 стандартных, 5 VIP)
    INSERT INTO seats (hall_id, seat_type_id, row_number, seat_number)
    SELECT v_hall_id, 
           CASE WHEN n <= 5 THEN (SELECT id FROM seat_types WHERE name = 'Standard') 
                ELSE (SELECT id FROM seat_types WHERE name = 'VIP') END,
           1, -- все в одном ряду
           n
    FROM generate_series(1, 10) AS n;
    
    -- Вызываем функцию
    SELECT create_movie_show_with_tickets(
        v_movie_id, 
        v_hall_id, 
        NOW() + INTERVAL '3 hours', 
        'Russian'::language_enum, 
        400.00
    ) INTO v_show_id;
    
    -- Проверяем разные цены
    SELECT price INTO v_actual_price 
    FROM tickets 
    WHERE movie_show_id = v_show_id 
    AND seat_id IN (SELECT id FROM seats WHERE seat_type_id = (SELECT id FROM seat_types WHERE name = 'Standard'))
    LIMIT 1;
    
    v_expected_price := ROUND(400.00 * 1.2 * 1.0, 2);
    RAISE NOTICE 'Цена стандартного места: % (ожидалось %)', v_actual_price, v_expected_price;
    
    SELECT price INTO v_actual_price 
    FROM tickets 
    WHERE movie_show_id = v_show_id 
    AND seat_id IN (SELECT id FROM seats WHERE seat_type_id = (SELECT id FROM seat_types WHERE name = 'VIP'))
    LIMIT 1;
    
    v_expected_price := ROUND(400.00 * 1.2 * 1.5, 2);
    RAISE NOTICE 'Цена VIP места: % (ожидалось %)', v_actual_price, v_expected_price;
    
    -- Очистка
    DELETE FROM tickets WHERE movie_show_id = v_show_id;
    DELETE FROM movie_shows WHERE id = v_show_id;
    DELETE FROM seats WHERE hall_id = v_hall_id;
    
    
    -- Тест 4: Ошибочные сценарии
    RAISE NOTICE '=== Тест 4: Ошибочные сценарии ===';
    
    -- 4.1 Несуществующий movie_id
    BEGIN
        SELECT create_movie_show_with_tickets(
            gen_random_uuid(), -- случайный UUID
            v_hall_id, 
            NOW() + INTERVAL '4 hours', 
            'English'::language_enum, 
            300.00
        ) INTO v_show_id;
        
        RAISE NOTICE 'Тест 4.1: Несуществующий movie_id - НЕПРОЙДЕН (ожидалась ошибка)';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 4.1: Несуществующий movie_id - ОШИБКА (ожидаемо): %', SQLERRM;
    END;
    
    -- 4.2 Несуществующий hall_id
    BEGIN
        SELECT create_movie_show_with_tickets(
            v_movie_id, 
            gen_random_uuid(), -- случайный UUID
            NOW() + INTERVAL '4 hours', 
            'English'::language_enum, 
            300.00
        ) INTO v_show_id;
        
        RAISE NOTICE 'Тест 4.2: Несуществующий hall_id - НЕПРОЙДЕН (ожидалась ошибка)';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 4.2: Несуществующий hall_id - ОШИБКА (ожидаемо): %', SQLERRM;
    END;
    
    -- 4.3 Отрицательная базовая цена
    BEGIN
        SELECT create_movie_show_with_tickets(
            v_movie_id, 
            v_hall_id, 
            NOW() + INTERVAL '4 hours', 
            'English'::language_enum, 
            -10.00
        ) INTO v_show_id;
        
        RAISE NOTICE 'Тест 4.3: Отрицательная базовая цена - НЕПРОЙДЕН (ожидалась ошибка)';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 4.3: Отрицательная базовая цена - ОШИБКА (ожидаемо): %', SQLERRM;
    END;
    
    -- 4.4 Нулевая базовая цена
    BEGIN
        SELECT create_movie_show_with_tickets(
            v_movie_id, 
            v_hall_id, 
            NOW() + INTERVAL '4 hours', 
            'English'::language_enum, 
            0.00
        ) INTO v_show_id;
        
        RAISE NOTICE 'Тест 4.4: Нулевая базовая цена - НЕПРОЙДЕН (ожидалась ошибка)';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 4.4: Нулевая базовая цена - ОШИБКА (ожидаемо): %', SQLERRM;
    END;
    
    
    -- Тест 5: Специальные случаи
    RAISE NOTICE '=== Тест 5: Специальные случаи ===';
    
    -- 5.1 Зал без мест
    BEGIN
        -- Создаем пустой зал
        INSERT INTO halls (screen_type_id, name, description)
        VALUES (v_screen_modifier, 'Empty Hall', 'Empty Hall Description')
        RETURNING id INTO v_hall_id;
        
        SELECT create_movie_show_with_tickets(
            v_movie_id, 
            v_hall_id, 
            NOW() + INTERVAL '5 hours', 
            'English'::language_enum, 
            300.00
        ) INTO v_show_id;
        
        -- Проверяем
        SELECT COUNT(*) INTO v_tickets_count FROM tickets WHERE movie_show_id = v_show_id;
        RAISE NOTICE 'Тест 5.1: Зал без мест - сеанс создан с ID: %', v_show_id;
        RAISE NOTICE 'Создано билетов: % (ожидалось 0)', v_tickets_count;
        
        -- Очистка
        DELETE FROM movie_shows WHERE id = v_show_id;
        DELETE FROM halls WHERE id = v_hall_id;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Тест 5.1: Зал без мест - ОШИБКА: %', SQLERRM;
    END;
    
    -- Окончательная очистка
    DELETE FROM seat_types WHERE id IN (SELECT id FROM seat_types WHERE name IN ('Standard', 'VIP'));
    DELETE FROM halls WHERE id = v_hall_id;
    DELETE FROM screen_types WHERE id = v_screen_modifier;
    DELETE FROM movies WHERE id = v_movie_id;
END $$;

RESET ROLE;