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

    SELECT box_office_revenue INTO initial_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);

    UPDATE tickets SET ticket_Status = 'Purchased'::ticket_Status_enum, user_id = new_user_id WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Test 1: Available → Purchased, Revenue izmenilsya s % na % (ozhidalos +%)', initial_revenue, new_revenue, ticket_price;

    initial_revenue := new_revenue;

    UPDATE tickets SET ticket_status = 'Available'::ticket_status_enum, user_id = NULL WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Test 2: Purchased → Available, Revenue izmenilsya s % na % (ozhidalos -%)', initial_revenue, new_revenue, ticket_price;

    initial_revenue := new_revenue;

    UPDATE tickets SET ticket_status = 'Reserved'::ticket_status_enum, user_id = new_user_id WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Test 3: Available → Reserved, Revenue izmenilsya s % na % (ozhidalos 0)', initial_revenue, new_revenue;

    initial_revenue := new_revenue;

    UPDATE tickets SET ticket_Status = 'Purchased'::ticket_Status_enum, user_id = new_user_id WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Test 4: Reserved → Purchased, Revenue izmenilsya s % na % (ozhidalos +%)', initial_revenue, new_revenue, ticket_price;

    initial_revenue := new_revenue;

    UPDATE tickets SET ticket_Status = 'Reserved'::ticket_Status_enum, user_id = new_user_id WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Test 5: Purchased → Reserved, Revenue izmenilsya s % na % (ozhidalos -%)', initial_revenue, new_revenue, ticket_price;

    initial_revenue := new_revenue;

    UPDATE tickets SET ticket_Status = 'Available'::ticket_Status_enum, user_id = NULL WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Test 6: Reserved → Available, Revenue izmenilsya s % na % (ozhidalos 0)', initial_revenue, new_revenue;

    initial_revenue := new_revenue;

    UPDATE tickets SET price = price + 100 WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Test 7: price + 100, Revenue izmenilsya s % na % (ozhidalos 0)', initial_revenue, new_revenue;

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

    SELECT box_office_revenue INTO initial_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);

    UPDATE tickets SET ticket_Status = 'Reserved'::ticket_Status_enum, user_id = new_user_id WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Test 1: Available → Reserved, Revenue izmenilsya s % na % (ozhidalos 0)', initial_revenue, new_revenue;

    initial_revenue := new_revenue;

    UPDATE tickets SET ticket_Status = 'Reserved'::ticket_Status_enum, user_id = new_user_id WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Test 2: Purchased → Reserved, Revenue izmenilsya s % na % (ozhidalos 0)', initial_revenue, new_revenue;

    initial_revenue := new_revenue;

    UPDATE tickets SET price = price + 100 WHERE id = ticket_id;
    SELECT box_office_revenue INTO new_revenue FROM movies WHERE id = (SELECT movie_id FROM movie_shows WHERE id = show_id);
    RAISE NOTICE 'Test 3: price + 100, Revenue izmenilsya s % na % (ozhidalos 0)', initial_revenue, new_revenue;
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

    INSERT INTO screen_types (name, description, price_modifier)
    VALUES ('Standard', 'Standard screen', 1.0)
    RETURNING id INTO v_screen_type_id;
    
    INSERT INTO halls (screen_type_id, name, description)
    VALUES (v_screen_type_id, 'Test Hall', 'Test Hall Description')
    RETURNING id INTO v_hall_id;
    
    INSERT INTO seat_types (name, description, price_modifier)
    VALUES ('Standard', 'Standard seat', 1.0)
    RETURNING id INTO v_seat_type_id;

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

    SELECT create_movie_show_with_tickets(
        v_movie_id, 
        v_hall_id, 
         '2023-01-01:12:12:12'::timestamp, 
        'English'::language_enum, 
        300.00
    ) INTO v_existing_show_id;
    
    RAISE NOTICE 'Message', v_existing_show_id;
    
    RAISE NOTICE '=== Testirovanie granichnykh usloviy ===';

    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_existing_show_id) + 
                          (SELECT duration FROM movies WHERE id = v_movie_id) + v_cleanup_time;
        
        INSERT INTO movie_shows (movie_id, hall_id, start_time, language)
        VALUES (v_movie_id, v_hall_id, v_test_start_time, 'English'::language_enum)
        RETURNING id INTO v_new_show_id;
        
        RAISE NOTICE 'Test 1: Nachalo srazu posle uborki - USPEKh (seans %)', v_new_show_id;
        DELETE FROM movie_shows WHERE id = v_new_show_id;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;

    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_existing_show_id) + 
                          (SELECT duration FROM movies WHERE id = v_movie_id) + v_cleanup_time;
        
        INSERT INTO movie_shows (movie_id, hall_id, start_time, language)
        VALUES (v_movie_id, v_hall_id, v_test_start_time, 'English'::language_enum)
        RETURNING id INTO v_new_show_id;
        
        RAISE NOTICE 'Test 2: Konets uborki = nachalo drugogo - USPEKh (seans %)', v_new_show_id;
        DELETE FROM movie_shows WHERE id = v_new_show_id;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;
    
    RAISE NOTICE '=== Testirovanie konfliktnykh stsenariev ===';

    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_existing_show_id) + INTERVAL '30 minutes';
        
        INSERT INTO movie_shows (movie_id, hall_id, start_time, language)
        VALUES (v_movie_id, v_hall_id, v_test_start_time, 'English'::language_enum);
        
        RAISE NOTICE 'Message';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;

    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_existing_show_id) - INTERVAL '30 minutes';
        
        INSERT INTO movie_shows (movie_id, hall_id, start_time, language)
        VALUES (v_movie_id, v_hall_id, v_test_start_time, 'English'::language_enum);
        
        RAISE NOTICE 'Message';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;

    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_existing_show_id) + 
                         (SELECT duration FROM movies WHERE id = v_movie_id) - INTERVAL '30 minutes';
        
        INSERT INTO movie_shows (movie_id, hall_id, start_time, language)
        VALUES (v_movie_id, v_hall_id, v_test_start_time, 'English'::language_enum);
        
        RAISE NOTICE 'Message';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;

    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_existing_show_id);
        
        INSERT INTO movie_shows (movie_id, hall_id, start_time, language)
        VALUES (v_movie_id, v_hall_id, v_test_start_time, 'English'::language_enum);
        
        RAISE NOTICE 'Message';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;

    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_existing_show_id) - INTERVAL '30 minutes';
        
        INSERT INTO movie_shows (movie_id, hall_id, start_time, language)
        VALUES (v_movie_id, v_hall_id, v_test_start_time, 'English'::language_enum);
        
        RAISE NOTICE 'Message';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;
    
    RAISE NOTICE '=== Testirovanie nekonfliktnykh stsenariev ===';

    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_existing_show_id) + 
                          (SELECT duration FROM movies WHERE id = v_movie_id) + v_cleanup_time + INTERVAL '1 hour';
        
        INSERT INTO movie_shows (movie_id, hall_id, start_time, language)
        VALUES (v_movie_id, v_hall_id, v_test_start_time, 'English'::language_enum)
        RETURNING id INTO v_new_show_id;
        
        RAISE NOTICE 'Test 8: Razdelnye intervaly - USPEKh (seans %)', v_new_show_id;
        DELETE FROM movie_shows WHERE id = v_new_show_id;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;

    BEGIN
        UPDATE movie_shows SET language = 'Russkiy'::language_enum WHERE id = v_existing_show_id;
        RAISE NOTICE 'Test 9: Izmenenie nevremennykh atributov - USPEKh';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;

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

    INSERT INTO screen_types (name, description, price_modifier)
    VALUES ('Standard', 'Standard screen', 1.0)
    RETURNING id INTO v_screen_type_id;
    
    INSERT INTO halls (screen_type_id, name, description)
    VALUES (v_screen_type_id, 'Test Hall', 'Test Hall Description')
    RETURNING id INTO v_hall_id;
    
    INSERT INTO seat_types (name, description, price_modifier)
    VALUES ('Standard', 'Standard seat', 1.0)
    RETURNING id INTO v_seat_type_id;

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

    SELECT create_movie_show_with_tickets(
        v_movie_id, 
        v_hall_id, 
        '2023-01-01:12:12:12'::timestamp, 
        'English'::language_enum, 
        300.00
    ) INTO v_base_show_id;

    SELECT create_movie_show_with_tickets(
        v_movie_id, 
        v_hall_id, 
        '2023-01-01:12:12:12'::timestamp + INTERVAL '5 hours', 
        'English'::language_enum, 
        300.00
    ) INTO v_test_show_id;
    
    RAISE NOTICE 'Message';
    
    RAISE NOTICE '=== Granichnye usloviya ===';

    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_base_show_id) + 
                          (SELECT duration FROM movies WHERE id = v_movie_id) + v_cleanup_time;
        
        UPDATE movie_shows 
        SET start_time = v_test_start_time 
        WHERE id = v_test_show_id;
        
        RAISE NOTICE 'Message';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;

    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_base_show_id) + 
                          (SELECT duration FROM movies WHERE id = v_movie_id) + v_cleanup_time;
        
        UPDATE movie_shows 
        SET start_time = v_test_start_time 
        WHERE id = v_test_show_id;
        
        RAISE NOTICE 'Message';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;
    
    RAISE NOTICE '=== Konfliktnye stsenarii ===';

    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_base_show_id) + INTERVAL '30 minutes';
        
        UPDATE movie_shows 
        SET start_time = v_test_start_time 
        WHERE id = v_test_show_id;
        
        RAISE NOTICE 'Message';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;

    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_base_show_id) - INTERVAL '30 minutes';
        
        UPDATE movie_shows 
        SET start_time = v_test_start_time 
        WHERE id = v_test_show_id;
        
        RAISE NOTICE 'Message';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;

    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_base_show_id) + 
                         (SELECT duration FROM movies WHERE id = v_movie_id) - INTERVAL '30 minutes';
        
        UPDATE movie_shows 
        SET start_time = v_test_start_time 
        WHERE id = v_test_show_id;
        
        RAISE NOTICE 'Message';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;

    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_base_show_id);
        
        UPDATE movie_shows 
        SET start_time = v_test_start_time 
        WHERE id = v_test_show_id;
        
        RAISE NOTICE 'Message';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;

    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_base_show_id) - INTERVAL '30 minutes';
        
        UPDATE movie_shows 
        SET start_time = v_test_start_time 
        WHERE id = v_test_show_id;
        
        RAISE NOTICE 'Message';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;
    
    RAISE NOTICE '=== Nekonfliktnye stsenarii ===';

    BEGIN
        v_test_start_time := (SELECT start_time FROM movie_shows WHERE id = v_base_show_id) + 
                          (SELECT duration FROM movies WHERE id = v_movie_id) + v_cleanup_time + INTERVAL '1 hour';
        
        UPDATE movie_shows 
        SET start_time = v_test_start_time 
        WHERE id = v_test_show_id;
        
        RAISE NOTICE 'Message';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;

    BEGIN
        UPDATE movie_shows 
        SET language = 'Russkiy'::language_enum 
        WHERE id = v_test_show_id;
        
        RAISE NOTICE 'Message';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;

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

    RAISE NOTICE '=== Test 1: Zal s 10 odinakovymi mestami (ryady != mesta v ryadu) ===';

    INSERT INTO seats (hall_id, seat_type_id, row_number, seat_number)
    SELECT v_hall_id, v_seat_modifier, 
           (n-1)/5 + 1, -- 2 ryada (1 i 2)
           (n-1)%5 + 1  -- 5 mest v ryadu
    FROM generate_series(1, 10) AS n;

    SELECT create_movie_show_with_tickets(
        v_movie_id, 
        v_hall_id, 
        NOW() + INTERVAL '1 hour', 
        'Russian'::language_enum, 
        300.00
    ) INTO v_show_id;

    SELECT COUNT(*) INTO v_tickets_count FROM tickets WHERE movie_show_id = v_show_id;
    SELECT price INTO v_actual_price FROM tickets WHERE movie_show_id = v_show_id LIMIT 1;
    v_expected_price := ROUND(300.00 * 1.2 * 1.0, 2); -- base_price * screen_mod * seat_mod
    
    RAISE NOTICE 'ID seansa: %', v_show_id;
    RAISE NOTICE 'Message', v_tickets_count;
    RAISE NOTICE 'Tsena bileta: % (ozhidalos %)', v_actual_price, v_expected_price;

    DELETE FROM tickets WHERE movie_show_id = v_show_id;
    DELETE FROM movie_shows WHERE id = v_show_id;
    DELETE FROM seats WHERE hall_id = v_hall_id;

    RAISE NOTICE '=== Test 2: Zal s 25 odinakovymi mestami (ryady == mesta v ryadu) ===';

    INSERT INTO seats (hall_id, seat_type_id, row_number, seat_number)
    SELECT v_hall_id, v_seat_modifier, 
           (n-1)/5 + 1, -- 5 ryadov
           (n-1)%5 + 1   -- 5 mest v ryadu
    FROM generate_series(1, 25) AS n;

    SELECT create_movie_show_with_tickets(
        v_movie_id, 
        v_hall_id, 
        NOW() + INTERVAL '2 hours', 
        'English'::language_enum, 
        350.00
    ) INTO v_show_id;

    SELECT COUNT(*) INTO v_tickets_count FROM tickets WHERE movie_show_id = v_show_id;
    RAISE NOTICE 'Message', v_tickets_count;

    DELETE FROM tickets WHERE movie_show_id = v_show_id;
    DELETE FROM movie_shows WHERE id = v_show_id;
    DELETE FROM seats WHERE hall_id = v_hall_id;

    RAISE NOTICE '=== Test 3: Zal s raznymi tipami mest ===';

    INSERT INTO seats (hall_id, seat_type_id, row_number, seat_number)
    SELECT v_hall_id, 
           CASE WHEN n <= 5 THEN (SELECT id FROM seat_types WHERE name = 'Standard') 
                ELSE (SELECT id FROM seat_types WHERE name = 'VIP') END,
           1, -- vse v odnom ryadu
           n
    FROM generate_series(1, 10) AS n;

    SELECT create_movie_show_with_tickets(
        v_movie_id, 
        v_hall_id, 
        NOW() + INTERVAL '3 hours', 
        'Russian'::language_enum, 
        400.00
    ) INTO v_show_id;

    SELECT price INTO v_actual_price 
    FROM tickets 
    WHERE movie_show_id = v_show_id 
    AND seat_id IN (SELECT id FROM seats WHERE seat_type_id = (SELECT id FROM seat_types WHERE name = 'Standard'))
    LIMIT 1;
    
    v_expected_price := ROUND(400.00 * 1.2 * 1.0, 2);
    RAISE NOTICE 'Tsena standartnogo mesta: % (ozhidalos %)', v_actual_price, v_expected_price;
    
    SELECT price INTO v_actual_price 
    FROM tickets 
    WHERE movie_show_id = v_show_id 
    AND seat_id IN (SELECT id FROM seats WHERE seat_type_id = (SELECT id FROM seat_types WHERE name = 'VIP'))
    LIMIT 1;
    
    v_expected_price := ROUND(400.00 * 1.2 * 1.5, 2);
    RAISE NOTICE 'Tsena VIP mesta: % (ozhidalos %)', v_actual_price, v_expected_price;

    DELETE FROM tickets WHERE movie_show_id = v_show_id;
    DELETE FROM movie_shows WHERE id = v_show_id;
    DELETE FROM seats WHERE hall_id = v_hall_id;

    RAISE NOTICE '=== Test 4: Oshibochnye stsenarii ===';

    BEGIN
        SELECT create_movie_show_with_tickets(
            gen_random_uuid(), -- sluchaynyy UUID
            v_hall_id, 
            NOW() + INTERVAL '4 hours', 
            'English'::language_enum, 
            300.00
        ) INTO v_show_id;
        
        RAISE NOTICE 'Message';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;

    BEGIN
        SELECT create_movie_show_with_tickets(
            v_movie_id, 
            gen_random_uuid(), -- sluchaynyy UUID
            NOW() + INTERVAL '4 hours', 
            'English'::language_enum, 
            300.00
        ) INTO v_show_id;
        
        RAISE NOTICE 'Message';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;

    BEGIN
        SELECT create_movie_show_with_tickets(
            v_movie_id, 
            v_hall_id, 
            NOW() + INTERVAL '4 hours', 
            'English'::language_enum, 
            -10.00
        ) INTO v_show_id;
        
        RAISE NOTICE 'Message';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;

    BEGIN
        SELECT create_movie_show_with_tickets(
            v_movie_id, 
            v_hall_id, 
            NOW() + INTERVAL '4 hours', 
            'English'::language_enum, 
            0.00
        ) INTO v_show_id;
        
        RAISE NOTICE 'Message';
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;

    RAISE NOTICE '=== Test 5: Spetsialnye sluchai ===';

    BEGIN

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

        SELECT COUNT(*) INTO v_tickets_count FROM tickets WHERE movie_show_id = v_show_id;
        RAISE NOTICE 'Message', v_show_id;
        RAISE NOTICE 'Message', v_tickets_count;

        DELETE FROM movie_shows WHERE id = v_show_id;
        DELETE FROM halls WHERE id = v_hall_id;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'Message', SQLERRM;
    END;

    DELETE FROM seat_types WHERE id IN (SELECT id FROM seat_types WHERE name IN ('Standard', 'VIP'));
    DELETE FROM halls WHERE id = v_hall_id;
    DELETE FROM screen_types WHERE id = v_screen_modifier;
    DELETE FROM movies WHERE id = v_movie_id;
END $$;

RESET ROLE;