package main

import (
	"time"

	"github.com/google/uuid"
)

var GenresData = []Genre{
	{ID: uuid.New().String(), Name: "Драма", Description: "Фильмы, в которых основное внимание уделяется эмоциональному развитию персонажей и сложным жизненным ситуациям."},
	{ID: uuid.New().String(), Name: "Комедия", Description: "Фильмы с юмористическим сюжетом, целью которых является развлечение зрителей."},
	{ID: uuid.New().String(), Name: "Ужасы", Description: "Фильмы, направленные на создание у зрителя чувства страха, ужаса и тревоги."},
	{ID: uuid.New().String(), Name: "Научная фантастика", Description: "Фильмы научной фантастики, часто основанные на идеях о будущем и развитии технологий."},
	{ID: uuid.New().String(), Name: "Романтика", Description: "Фильмы, в которых основное внимание уделяется романтическим отношениям."},
	{ID: uuid.New().String(), Name: "Триллер", Description: "Фильмы с напряженным, захватывающим сюжетом, часто с элементами мистики или криминала."},
	{ID: uuid.New().String(), Name: "Фэнтези", Description: "Фильмы, в которых представлены вымышленные миры, магия и фантастические существа."},
	{ID: uuid.New().String(), Name: "Приключения", Description: "Фильмы, в которых главный акцент на приключениях и исследовании новых миров."},
	{ID: uuid.New().String(), Name: "Документальный", Description: "Фильмы, которые исследуют реальную жизнь, события или явления, часто с образовательной целью."},
	{ID: uuid.New().String(), Name: "Анимация", Description: "Фильмы, в которых используются различные методы анимации для создания визуального контента."},
	{ID: uuid.New().String(), Name: "Мистика", Description: "Фильмы с элементами расследования, разгадки тайны или преступления."},
}

var ScreenTypesData = []ScreenType{
	{ID: uuid.New().String(), Name: "LED", Description: "Современные экраны, использующие светодиоды для отображения изображения с высокой яркостью и контрастностью."},
	{ID: uuid.New().String(), Name: "LCD", Description: "Жидкокристаллические экраны, обеспечивающие хорошее качество изображения и энергоэффективность."},
	{ID: uuid.New().String(), Name: "DLP", Description: "Цифровые проекторы на основе технологии цифровой обработки света, обеспечивающие высокое качество изображения."},
	{ID: uuid.New().String(), Name: "OLED", Description: "Экраны с органическими светодиодами, обеспечивающие глубокие черные цвета и широкий угол обзора."},
	{ID: uuid.New().String(), Name: "Проекционный экран", Description: "Экран, на который проецируется изображение с проектора."},
	{ID: uuid.New().String(), Name: "Система 3D", Description: "Оборудование для показа фильмов в 3D-формате, включая специальные экраны и очки."},
	{ID: uuid.New().String(), Name: "IMAX", Description: "Специальные экраны и проекторы для показа фильмов в формате IMAX, обеспечивающие уникальный опыт просмотра."},
	{ID: uuid.New().String(), Name: "Сквозной экран", Description: "Экран, который позволяет зрителям видеть изображение с обеих сторон."},
	{ID: uuid.New().String(), Name: "Мобильный экран", Description: "Переносные экраны, используемые для временных показов или мероприятий на открытом воздухе."},
	{ID: uuid.New().String(), Name: "Система звука для экрана", Description: "Оборудование, интегрированное с экраном для обеспечения качественного звука."},
}

var SeatTypesData = []SeatType{
	{ID: uuid.New().String(), Name: "Стандартное", Description: "Обычное кресло для зрителей с комфортной посадкой."},
	{ID: uuid.New().String(), Name: "VIP", Description: "Комфортабельные кресла с дополнительными удобствами, такими как подставки для ног и увеличенное пространство."},
	{ID: uuid.New().String(), Name: "Люкс", Description: "Эксклюзивные места с повышенным комфортом, часто с возможностью заказа еды и напитков."},
	{ID: uuid.New().String(), Name: "Кресло-качалка", Description: "Кресло, которое может качаться, обеспечивая дополнительный комфорт."},
	{ID: uuid.New().String(), Name: "Сиденье для инвалидов", Description: "Специально оборудованные места для зрителей с ограниченными возможностями."},
	{ID: uuid.New().String(), Name: "Семейное", Description: "Места, предназначенные для семейных групп, часто расположенные рядом друг с другом."},
	{ID: uuid.New().String(), Name: "Балкон", Description: "Места, расположенные на верхнем уровне зала, обеспечивающие хороший обзор."},
	{ID: uuid.New().String(), Name: "Премиум", Description: "Места с лучшим расположением и дополнительными удобствами."},
	{ID: uuid.New().String(), Name: "Кресло с подогревом", Description: "Кресло, оснащенное функцией подогрева для дополнительного комфорта."},
	{ID: uuid.New().String(), Name: "Кресло с массажем", Description: "Кресло, которое предлагает функции массажа для расслабления зрителей."},
}

var HallsData = []Hall{
	{
		ID:           uuid.New().String(),
		Name:         "Основной кинозал",
		Capacity:     500,
		ScreenTypeID: ScreenTypesData[0].ID,
		Description:  "Главный кинозал кинотеатра с современным оборудованием",
	},
	{
		ID:           uuid.New().String(),
		Name:         "Малый кинозал",
		Capacity:     150,
		ScreenTypeID: ScreenTypesData[3].ID,
		Description:  "Небольшой уютный кинозал для камерных просмотров",
	},
	{
		ID:           uuid.New().String(),
		Name:         "VIP кинозал",
		Capacity:     50,
		ScreenTypeID: ScreenTypesData[2].ID,
		Description:  "Премиальный кинозал с креслами-реклайнерами и сервисом",
	},
	{
		ID:           uuid.New().String(),
		Name:         "IMAX кинозал",
		Capacity:     300,
		ScreenTypeID: ScreenTypesData[4].ID,
		Description:  "Зал с технологией IMAX Laser для максимального погружения",
	},
	{
		ID:           uuid.New().String(),
		Name:         "4DX кинозал",
		Capacity:     200,
		ScreenTypeID: ScreenTypesData[3].ID,
		Description:  "Зал с движущимися креслами и спецэффектами",
	},
}

var MoviesData = []Movie{
	{
		ID:               uuid.New().String(),
		Title:            "Интерстеллар",
		Duration:         "02:49:00",
		Rating:           8.6,
		Description:      "Фантастический эпос о путешествии группы исследователей, которые используют недавно обнаруженный червоточину, чтобы обойти ограничения космических путешествий человека и покорить огромные расстояния на межзвёздном корабле.",
		AgeLimit:         12,
		BoxOfficeRevenue: 701.7,
		ReleaseDate:      MustParseTime("2014-11-06"),
	},
	{
		ID:               uuid.New().String(),
		Title:            "Начало",
		Duration:         "02:28:00",
		Rating:           8.8,
		Description:      "Криминальный триллер о технологии проникновения в сны и краже идей из подсознания.",
		AgeLimit:         12,
		BoxOfficeRevenue: 836.8,
		ReleaseDate:      MustParseTime("2010-07-16"),
	},
	{
		ID:               uuid.New().String(),
		Title:            "Довод",
		Duration:         "02:30:00",
		Rating:           7.5,
		Description:      "Шпионский боевик о секретной технологии инверсии времени, которая может предотвратить Третью мировую войну.",
		AgeLimit:         16,
		BoxOfficeRevenue: 363.7,
		ReleaseDate:      MustParseTime("2020-08-26"),
	},
	{
		ID:               uuid.New().String(),
		Title:            "Темный рыцарь",
		Duration:         "02:32:00",
		Rating:           9.0,
		Description:      "Бэтмен, Джокер и Харви Дент вступают в смертельную схватку за душу Готэма.",
		AgeLimit:         16,
		BoxOfficeRevenue: 1004.6,
		ReleaseDate:      MustParseTime("2008-07-18"),
	},
	{
		ID:               uuid.New().String(),
		Title:            "Зеленая книга",
		Duration:         "02:10:00",
		Rating:           8.2,
		Description:      "История дружбы афроамериканского пианиста и его итальянского водителя во время турне по югу США в 1960-х.",
		AgeLimit:         12,
		BoxOfficeRevenue: 321.7,
		ReleaseDate:      MustParseTime("2018-11-21"),
	},
	{
		ID:               uuid.New().String(),
		Title:            "Черная книга",
		Duration:         "02:10:00",
		Rating:           8.2,
		Description:      "История дружбы",
		AgeLimit:         12,
		BoxOfficeRevenue: 321.7,
		ReleaseDate:      MustParseTime("2018-11-21"),
	},
}

var UsersData = []User{
	{
		ID:           uuid.New().String(),
		Name:         "Иван Иванов",
		Email:        "ivan@example.com",
		PasswordHash: "$2a$10$xS.xH8z3bJ1J5hNtGvXZfez7v6JQY9W7kZf3JvYbW6cXrV1nYd2E3C",
		BirthDate:    "1990-05-15",
		IsBlocked:    false,
		IsAdmin:      false,
	},
	{
		ID:           uuid.New().String(),
		Name:         "Петр Петров",
		Email:        "petr@example.com",
		PasswordHash: "$2a$10$yT.9H7v2cK2J4mNtHvWZfez7v6JQY9W7kZf3JvYbW6cXrV1nYd2E3C",
		BirthDate:    "1985-10-20",
		IsBlocked:    false,
		IsAdmin:      true,
	},
	{
		ID:           uuid.New().String(),
		Name:         "Сергей Сергеев",
		Email:        "sergey@example.com",
		PasswordHash: "$2a$10$zU.8H6w1bL3K5nNtGvXZfez7v6JQY9W7kZf3JvYbW6cXrV1nYd2E3C",
		BirthDate:    "1995-03-10",
		IsBlocked:    true,
		IsAdmin:      false,
	},
}

var MovieShowsData = []MovieShow{
	{
		ID:        uuid.New().String(),
		MovieID:   MoviesData[0].ID,
		HallID:    HallsData[0].ID,
		StartTime: time.Now().Add(24 * time.Hour),
		Language:  "Русский",
	},
	{
		ID:        uuid.New().String(),
		MovieID:   MoviesData[1].ID,
		HallID:    HallsData[1].ID,
		StartTime: time.Now().Add(26 * time.Hour),
		Language:  "English",
	},
	{
		ID:        uuid.New().String(),
		MovieID:   MoviesData[2].ID,
		HallID:    HallsData[2].ID,
		StartTime: time.Now().Add(48 * time.Hour),
		Language:  "Русский",
	},
}

var SeatsData = []Seat{
	{
		ID:         uuid.New().String(),
		HallID:     HallsData[0].ID,
		SeatTypeID: SeatTypesData[0].ID,
		RowNumber:  1,
		SeatNumber: 1,
	},
	{
		ID:         uuid.New().String(),
		HallID:     HallsData[0].ID,
		SeatTypeID: SeatTypesData[1].ID,
		RowNumber:  2,
		SeatNumber: 5,
	},
	{
		ID:         uuid.New().String(),
		HallID:     HallsData[1].ID,
		SeatTypeID: SeatTypesData[2].ID,
		RowNumber:  3,
		SeatNumber: 10,
	},
}

var TicketsData = []Ticket{
	{
		ID:          uuid.New().String(),
		MovieShowID: MovieShowsData[0].ID,
		SeatID:      SeatsData[0].ID,
		Status:      "Purchased",
		Price:       500.00,
	},
	{
		ID:          uuid.New().String(),
		MovieShowID: MovieShowsData[1].ID,
		SeatID:      SeatsData[1].ID,
		Status:      "Reserved",
		Price:       750.00,
	},
	{
		ID:          uuid.New().String(),
		MovieShowID: MovieShowsData[2].ID,
		SeatID:      SeatsData[2].ID,
		Status:      "Available",
		Price:       1000.00,
	},
}

var ReviewsData = []Review{
	{
		UserID:  UsersData[0].ID,
		MovieID: MoviesData[0].ID,
		Rating:  9.5,
		Comment: "Отличный фильм с глубоким смыслом и потрясающей графикой!",
	},
	{
		UserID:  UsersData[1].ID,
		MovieID: MoviesData[1].ID,
		Rating:  8.0,
		Comment: "Интересный сюжет, но сложный для восприятия с первого раза.",
	},
	{
		UserID:  UsersData[0].ID,
		MovieID: MoviesData[2].ID,
		Rating:  7.5,
		Comment: "Хороший боевик, но слишком много нелогичных моментов.",
	},
}

var MoviesGenresData = [][2]string{
	{MoviesData[0].ID, GenresData[3].ID}, // Интерстеллар - Научная фантастика
	{MoviesData[0].ID, GenresData[0].ID}, // Интерстеллар - Драма
	{MoviesData[1].ID, GenresData[5].ID}, // Начало - Триллер
	{MoviesData[1].ID, GenresData[3].ID}, // Начало - Научная фантастика
	{MoviesData[2].ID, GenresData[5].ID}, // Довод - Триллер
	{MoviesData[3].ID, GenresData[5].ID}, // Темный рыцарь - Триллер
	{MoviesData[3].ID, GenresData[0].ID}, // Темный рыцарь - Драма
	{MoviesData[4].ID, GenresData[0].ID}, // Зеленая книга - Драма
	{MoviesData[4].ID, GenresData[1].ID}, // Зеленая книга - Комедия
}
