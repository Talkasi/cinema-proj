package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func validateAllScreenTypeAdmin(w http.ResponseWriter, e ScreenTypeAdmin) bool {
	e.Name = PrepareString(e.Name)
	e.Description = PrepareString(e.Description)

	if err := validateScreenTypeName(e.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	if err := validateScreenTypeDesctiption(e.Description); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	if e.PriceModifier <= 0 {
		http.Error(w, "Наценка должна быть положительной", http.StatusBadRequest)
		return false
	}

	return true
}

func validateAllScreenTypeData(w http.ResponseWriter, e ScreenTypeData) bool {
	e.Name = PrepareString(e.Name)
	e.Description = PrepareString(e.Description)

	if err := validateScreenTypeName(e.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	if err := validateScreenTypeDesctiption(e.Description); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}

func validateScreenTypeName(name string) error {
	validNameRegex := regexp.MustCompile(`\S`)
	if !validNameRegex.MatchString(name) {
		return errors.New("имя не может быть пустым или состоять только из пробелов")
	}
	if len(name) > 100 {
		return errors.New("имя не может превышать 100 символов")
	}
	return nil
}

func validateScreenTypeDesctiption(description string) error {
	validDescriptionRegex := regexp.MustCompile(`\S`)
	if !validDescriptionRegex.MatchString(description) {
		return errors.New("описание не может быть пустым или состоять только из пробелов")
	}
	if len(description) > 1000 {
		return errors.New("описание не может превышать 1000 символов")
	}
	return nil
}

// @Summary Получить все типы экранов (guest | user | admin)
// @Description Возвращает список всех типов экранов, содержащихся в базе данных.
// @Tags Типы экранов
// @Produce json
// @Success 200 {array} ScreenType "Список типов экранов"
// @Failure 404 {object} ErrorResponse "Типы экранов не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /screen-types [get]
func GetScreenTypes(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(context.Background(), "SELECT id, name, description FROM screen_types")
		if HandleDatabaseError(w, err, "типами экранов") {
			return
		}
		defer rows.Close()

		var types []ScreenType
		for rows.Next() {
			var e ScreenType
			if err := rows.Scan(&e.ID, &e.Name, &e.Description); HandleDatabaseError(w, err, "типом экранов") {
				return
			}
			types = append(types, e)
		}

		if len(types) == 0 {
			http.Error(w, "Типы экранов не найдены", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(types)
	}
}

// @Summary Получить тип экрана по ID (guest | user | admin)
// @Description Возвращает тип экрана по ID.
// @Tags Типы экранов
// @Produce json
// @Param id path string true "ID типа экрана"
// @Success 200 {object} ScreenType "Тип экрана"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 404 {object} ErrorResponse "Тип экрана не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /screen-types/{id} [get]
func GetScreenTypeByID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var e ScreenType
		e.ID = id.String()
		err := db.QueryRow(context.Background(),
			"SELECT name, description FROM screen_types WHERE id = $1", id).
			Scan(&e.Name, &e.Description)

		if IsError(w, err) {
			return
		}

		json.NewEncoder(w).Encode(e)
	}
}

// @Summary Создать тип экрана (admin)
// @Description Создаёт новый тип экрана.
// @Tags Типы экранов
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param screen_type body ScreenTypeAdmin true "Данные типа экрана"
// @Success 201 {object} CreateResponse "ID созданного типа экрана"
// @Failure 400 {object} ErrorResponse "В запросе предоставлены неверные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /screen-types [post]
func CreateScreenType(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var e ScreenTypeAdmin
		if !DecodeJSONBody(w, r, &e) {
			return
		}
		if !validateAllScreenTypeAdmin(w, e) {
			return
		}

		id := uuid.New()
		_, err := db.Exec(context.Background(),
			"INSERT INTO screen_types (id, name, description, price_modifier) VALUES ($1, $2, $3, $4)",
			id, e.Name, e.Description, e.PriceModifier)

		if IsError(w, err) {
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(id.String())
	}
}

// @Summary Обновить тип экрана (admin)
// @Description Обновляет существующий тип экрана.
// @Tags Типы экранов
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID типа экрана"
// @Param screen_type body ScreenTypeAdmin true "Обновлённые данные типа экрана"
// @Success 200 "Данные о типе экрана успешно обновлены"
// @Failure 400 {object} ErrorResponse "В запросе предоставлены неверные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Тип экранов не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /screen-types/{id} [put]
func UpdateScreenType(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var e ScreenTypeAdmin
		if !DecodeJSONBody(w, r, &e) {
			return
		}
		if !validateAllScreenTypeAdmin(w, e) {
			return
		}

		res, err := db.Exec(context.Background(),
			"UPDATE screen_types SET name=$1, description=$2, price_modifier=$3 WHERE id=$4",
			e.Name, e.Description, e.PriceModifier, id)

		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// @Summary Удалить тип экрана (admin)
// @Description Удаляет тип экрана по ID.
// @Tags Типы экранов
// @Param id path string true "ID типа экрана"
// @Security BearerAuth
// @Success 204 "Данные о типе экрана успешно удалены"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Тип экрана не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /screen-types/{id} [delete]
func DeleteScreenType(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		res, err := db.Exec(context.Background(),
			"DELETE FROM screen_types WHERE id = $1", id)
		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// @Summary Поиск типов экранов по названию (guest | user | admin)
// @Description Возвращает типы экранов, название которых содержит указанную строку.
// @Tags Типы экранов
// @Produce json
// @Param query query string true "Поисковый запрос"
// @Success 200 {array} ScreenType "Список типов экранов"
// @Failure 400 {object} ErrorResponse "Строка поиска пуста"
// @Failure 404 {object} ErrorResponse "Типы экранов не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /screen-types/search [get]
func SearchScreenTypes(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		query = PrepareString(query)
		if query == "" {
			http.Error(w, "Строка поиска пуста", http.StatusBadRequest)
			return
		}

		rows, err := db.Query(context.Background(),
			"SELECT id, name, description FROM screen_types WHERE name ILIKE $1", "%"+query+"%")
		if IsError(w, err) {
			return
		}
		defer rows.Close()

		var types []ScreenType
		for rows.Next() {
			var e ScreenType
			if err := rows.Scan(&e.ID, &e.Name, &e.Description); IsError(w, err) {
				return
			}
			types = append(types, e)
		}

		if len(types) == 0 {
			http.Error(w, "Типы экранов не найдены", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(types)
	}
}
