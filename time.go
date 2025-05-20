package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	maxSize    = 6                   // Максимальный размер таблицы (10^6 записей)
	numSamples = 50                  // Количество замеров для усреднения
	minTime    = 1 * time.Nanosecond // Минимальное время для отображения
	dbConnStr  = "host=localhost port=5432 user=postgres password=postgres dbname=cinema sslmode=disable"
)

func time_all() {
	// Генерация размеров таблиц
	sizes := generateSizes(maxSize)

	// Результаты
	timesWithoutIndex := make([]float64, len(sizes))
	timesWithIndex := make([]float64, len(sizes))

	// Подключение к БД
	db, err := sql.Open("pgx", dbConnStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	println(len(sizes), sizes)
	// Запуск тестов
	for i, size := range sizes {
		println("start", i, size)
		log.Printf("Testing with size: %d\n", size)

		// Подготовка тестовых данных
		if err := prepareTestData(db, size); err != nil {
			log.Fatal(err)
		}

		// Тест без индекса
		avgTime, err := testQueryPerformance(db, size, false)
		if err != nil {
			log.Fatal(err)
		}
		timesWithoutIndex[i] = avgTime.Seconds()

		// Создание индекса
		if _, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_movie_shows_hall_time ON movie_shows(hall_id, start_time)"); err != nil {
			log.Fatal(err)
		}

		// Тест с индексом
		avgTime, err = testQueryPerformance(db, size, true)
		if err != nil {
			log.Fatal(err)
		}
		timesWithIndex[i] = avgTime.Seconds()

		// Очистка данных
		if _, err := db.Exec("TRUNCATE TABLE movie_shows, movies, halls CASCADE"); err != nil {
			log.Fatal(err)
		}

		println("end", i, size)
	}

	// Построение графика
	if err := plotResults(sizes, timesWithoutIndex, timesWithIndex); err != nil {
		log.Fatal(err)
	}
}

func generateSizes(maxSize int) []int {
	var sizes []int
	for i := 1; i <= maxSize; i++ {
		power := 1
		for j := 0; j < i; j++ {
			power *= 10
		}
		sizes = append(sizes, power)

		// Добавляем промежуточные значения
		if i < maxSize {
			for j := 2; j <= 9; j++ {
				sizes = append(sizes, power*j)
			}
		}
	}
	return sizes
}

func prepareTestData(db *sql.DB, size int) error {
	// Создание тестовых фильмов и залов
	if _, err := db.Exec(`
        INSERT INTO movies (id, title, duration, description, age_limit, release_date)
        SELECT 
            gen_random_uuid(),
            'Movie ' || seq,
            '02:30:00'::time,
            'Description for movie ' || seq,
            12,
            NOW() - (random() * 365 * 5 || ' days')::interval
        FROM generate_series(1, 100) seq
        ON CONFLICT DO NOTHING
    `); err != nil {
		return fmt.Errorf("inserting movies: %w", err)
	}

	// Создание screen_types
	if _, err := db.Exec(`
        INSERT INTO screen_types (id, name, description)
        VALUES 
            (gen_random_uuid(), 'Standard', 'Standard screen'),
            (gen_random_uuid(), 'IMAX', 'IMAX screen')
        ON CONFLICT DO NOTHING
    `); err != nil {
		return fmt.Errorf("inserting screen types: %w", err)
	}

	// Создание залов
	if _, err := db.Exec(`
        INSERT INTO halls (id, name, screen_type_id, description)
        SELECT 
            gen_random_uuid(),
            'Hall ' || seq,
            (SELECT id FROM screen_types ORDER BY random() LIMIT 1),
            'Description for hall ' || seq
        FROM generate_series(1, 10) seq
        ON CONFLICT DO NOTHING
    `); err != nil {
		return fmt.Errorf("inserting halls: %w", err)
	}

	// Отключаем триггер для массовой вставки
	if _, err := db.Exec("ALTER TABLE movie_shows DISABLE TRIGGER check_movie_show_conflict_before_insert_or_update"); err != nil {
		return fmt.Errorf("disabling trigger: %w", err)
	}

	// Создание сеансов с фиксированным интервалом, чтобы избежать конфликтов
	if _, err := db.Exec(`
        INSERT INTO movie_shows (id, movie_id, hall_id, start_time, language)
        WITH hall_movies AS (
            SELECT 
                h.id as hall_id,
                m.id as movie_id,
                m.duration as duration,
                (ARRAY['English', 'Spanish', 'French', 'German', 'Italian', 'Русский'])[1 + (random() * 5)::int]::language_enum as language
            FROM 
                (SELECT id FROM halls LIMIT 10) h
                CROSS JOIN (SELECT id, duration FROM movies LIMIT 100) m
            ORDER BY random()
            LIMIT $1
        )
        SELECT
            gen_random_uuid(),
            hm.movie_id,
            hm.hall_id,
            NOW() + (seq * interval '4 hours'), -- Фиксированный интервал 4 часа между сеансами
            hm.language
        FROM 
            hall_movies hm,
            generate_series(0, $1-1) seq
    `, size); err != nil {
		return fmt.Errorf("inserting movie shows: %w", err)
	}

	// Включаем триггер обратно
	if _, err := db.Exec("ALTER TABLE movie_shows ENABLE TRIGGER check_movie_show_conflict_before_insert_or_update"); err != nil {
		return fmt.Errorf("enabling trigger: %w", err)
	}

	return nil
}

func testQueryPerformance(db *sql.DB, size int, withIndex bool) (time.Duration, error) {
	var totalTime time.Duration

	// Получаем список всех hall_id для тестирования
	rows, err := db.Query("SELECT id FROM halls")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var hallIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		hallIDs = append(hallIDs, id)
	}

	for i := 0; i < numSamples; i++ {
		// Выбираем случайный hall_id
		hallID := hallIDs[rand.Intn(len(hallIDs))]

		// Генерируем случайный временной диапазон (2 часа)
		randomHours := rand.Intn(48) - 24 // от -24 до +24 часов от текущего времени
		startTime := time.Now().Add(time.Duration(randomHours) * time.Hour)
		endTime := startTime.Add(2 * time.Hour)

		// Измеряем время выполнения запроса
		start := time.Now()

		rows, err := db.Query(`
			SELECT id, movie_id, hall_id, start_time, language 
			FROM movie_shows 
			WHERE hall_id = $1 
			AND start_time BETWEEN $2 AND $3
		`, hallID, startTime, endTime)

		elapsed := time.Since(start)
		if elapsed < minTime {
			elapsed = minTime
		}

		if err != nil {
			return 0, err
		}
		rows.Close()

		totalTime += elapsed
	}

	return totalTime / time.Duration(numSamples), nil
}

func plotResults(sizes []int, timesWithoutIndex, timesWithIndex []float64) error {
	p := plot.New()

	p.Title.Text = "Производительность запросов к таблице movie_shows"
	p.X.Label.Text = "Размер таблицы (количество сеансов)"
	p.Y.Label.Text = "Среднее время выполнения (сек)"
	p.X.Scale = plot.LogScale{}
	p.Y.Scale = plot.LogScale{}
	p.X.Tick.Marker = plot.LogTicks{}
	p.Y.Tick.Marker = plot.LogTicks{}

	// Создаем точки для графиков
	withoutIndexPoints := make(plotter.XYs, len(sizes))
	withIndexPoints := make(plotter.XYs, len(sizes))

	for i := range sizes {
		withoutIndexPoints[i].X = float64(sizes[i])
		withoutIndexPoints[i].Y = timesWithoutIndex[i]

		withIndexPoints[i].X = float64(sizes[i])
		withIndexPoints[i].Y = timesWithIndex[i]
	}

	// Добавляем линии на график
	err := plotutil.AddLinePoints(p,
		"Без индекса", withoutIndexPoints,
		"С индексом", withIndexPoints,
	)
	if err != nil {
		return err
	}

	// Сохраняем график в файл
	if err := p.Save(10*vg.Inch, 6*vg.Inch, "movie_shows_performance.png"); err != nil {
		return err
	}

	fmt.Println("График сохранен в movie_shows_performance.png")
	return nil
}
