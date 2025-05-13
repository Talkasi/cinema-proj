package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func validateAllScreenTypeData(w http.ResponseWriter, e ScreenTypeData) bool {
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

// @Summary Получить все типы оборудования
// @Description Возвращает список всех типов оборудования
// @Tags screen-types
// @Produce json
// @Security BearerAuth
// @Success 200 {array} ScreenType "Список типов оборудования"
// @Failure 404 {object} ErrorResponse "Типы оборудования не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /screen-types [get]
func GetScreenTypes(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(context.Background(), "SELECT id, name, description FROM screen_types")
		if HandleDatabaseError(w, err, "типами оборудования") {
			return
		}
		defer rows.Close()

		var types []ScreenType
		for rows.Next() {
			var e ScreenType
			if err := rows.Scan(&e.ID, &e.Name, &e.Description); HandleDatabaseError(w, err, "типом оборудования") {
				return
			}
			types = append(types, e)
		}

		if len(types) == 0 {
			http.Error(w, "Типы оборудования не найдены", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(types)
	}
}

// @Summary Получить тип оборудования по ID
// @Description Возвращает тип оборудования по его ID
// @Tags screen-types
// @Produce json
// @Param id path string true "UUID типа оборудования"
// @Security BearerAuth
// @Success 200 {object} ScreenType "Тип оборудования"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 404 {object} ErrorResponse "Тип оборудования не найден"
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

// @Summary Создать тип оборудования
// @Description Создаёт новый тип оборудования
// @Tags screen-types
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param screen_type body ScreenTypeData true "Данные типа оборудования"
// @Success 201 {object} string "UUID созданного типа оборудования"
// @Failure 400 {object} ErrorResponse "Неверный формат JSON"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /screen-types [post]
func CreateScreenType(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var e ScreenTypeData
		if !DecodeJSONBody(w, r, &e) {
			return
		}
		if !validateAllScreenTypeData(w, e) {
			return
		}

		id := uuid.New()
		_, err := db.Exec(context.Background(),
			"INSERT INTO screen_types (id, name, description) VALUES ($1, $2, $3)",
			id, e.Name, e.Description)

		if IsError(w, err) {
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(id.String())
	}
}

// @Summary Обновить тип оборудования
// @Description Обновляет существующий тип оборудования
// @Tags screen-types
// @Accept json
// @Produce json
// @Param id path string true "UUID типа оборудования"
// @Param screen_type body ScreenTypeData true "Обновлённые данные типа оборудования"
// @Security BearerAuth
// @Success 200 "Тип оборудования успешно обновлён"
// @Failure 400 {object} ErrorResponse "Неверный формат ID/JSON или пустые поля"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Тип оборудования не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /screen-types/{id} [put]
func UpdateScreenType(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var e ScreenTypeData
		if !DecodeJSONBody(w, r, &e) {
			return
		}
		if !validateAllScreenTypeData(w, e) {
			return
		}

		res, err := db.Exec(context.Background(),
			"UPDATE screen_types SET name=$1, description=$2 WHERE id=$3",
			e.Name, e.Description, id)

		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// @Summary Удалить тип оборудования
// @Description Удаляет тип оборудования по его ID
// @Tags screen-types
// @Param id path string true "UUID типа оборудования"
// @Security BearerAuth
// @Success 204 "Тип оборудования успешно удалён"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Тип оборудования не найден"
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
