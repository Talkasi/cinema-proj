package main

import "github.com/google/uuid"

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

var EquipmentTypesData = []EquipmentType{
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

var TicketStatusesData = []TicketStatus{
	{ID: uuid.New().String(), Name: "Забронирован"},
	{ID: uuid.New().String(), Name: "Куплен"},
	{ID: uuid.New().String(), Name: "Возвращен"},
}

var HallsData = []Hall{
	{
		ID:              uuid.New().String(),
		Name:            "Основной зал",
		Capacity:        500,
		EquipmentTypeID: EquipmentTypesData[0].ID,
		Description:     "Главный зал кинотеатра с современным оборудованием",
	},
	{
		ID:              uuid.New().String(),
		Name:            "Малый зал",
		Capacity:        150,
		EquipmentTypeID: EquipmentTypesData[3].ID,
		Description:     "Небольшой уютный зал для камерных просмотров",
	},
	{
		ID:              uuid.New().String(),
		Name:            "VIP зал",
		Capacity:        50,
		EquipmentTypeID: EquipmentTypesData[2].ID,
		Description:     "Премиальный зал с креслами-реклайнерами и сервисом",
	},
	{
		ID:              uuid.New().String(),
		Name:            "IMAX зал",
		Capacity:        300,
		EquipmentTypeID: EquipmentTypesData[4].ID,
		Description:     "Зал с технологией IMAX Laser для максимального погружения",
	},
	{
		ID:              uuid.New().String(),
		Name:            "4DX зал",
		Capacity:        200,
		EquipmentTypeID: EquipmentTypesData[3].ID,
		Description:     "Зал с движущимися креслами и спецэффектами",
	},
}
