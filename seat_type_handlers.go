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

func validateAllSeatTypeData(w http.ResponseWriter, s SeatTypeData) bool {
	s.Name = PrepareString(s.Name)
	s.Description = PrepareString(s.Description)

	if err := validateSeatTypeName(s.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	if err := validateSeatTypeDescription(s.Description); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}

func validateSeatTypeName(name string) error {
	if len(name) == 0 || !regexp.MustCompile(`\S`).MatchString(name) {
		return errors.New("имя не может быть пустым или состоять только из пробелов")
	}
	if len(name) > 100 {
		return errors.New("имя не может превышать 100 символов")
	}
	return nil
}

func validateSeatTypeDescription(desc string) error {
	if len(desc) == 0 || !regexp.MustCompile(`\S`).MatchString(desc) {
		return errors.New("описание не может быть пустым или состоять только из пробелов")
	}
	if len(desc) > 1000 {
		return errors.New("описание не может превышать 1000 символов")
	}
	return nil
}

// @Summary Получить все типы мест
// @Description Возвращает список всех типов мест, содержащихся в базе данных.
// @Tags Типы мест
// @Produce json
// @Security BearerAuth
// @Success 200 {array} SeatType "Список типов мест"
// @Failure 404 {object} ErrorResponse "Типы мест не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /seat-types [get]
func GetSeatTypes(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(context.Background(), "SELECT id, name, description FROM seat_types")
		if HandleDatabaseError(w, err, "типами мест") {
			return
		}
		defer rows.Close()

		var types []SeatType
		for rows.Next() {
			var s SeatType
			if err := rows.Scan(&s.ID, &s.Name, &s.Description); HandleDatabaseError(w, err, "типом места") {
				return
			}
			types = append(types, s)
		}

		if len(types) == 0 {
			http.Error(w, "Типы мест не найдены", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(types)
	}
}

// @Summary Получить тип места по ID
// @Description Возвращает тип места по ID.
// @Tags Типы мест
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID типа места"
// @Success 200 {object} SeatType "Тип места"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 404 {object} ErrorResponse "Тип места не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /seat-types/{id} [get]
func GetSeatTypeByID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var s SeatType
		s.ID = id.String()
		err := db.QueryRow(context.Background(),
			"SELECT name, description FROM seat_types WHERE id = $1", id).
			Scan(&s.Name, &s.Description)

		if IsError(w, err) {
			return
		}

		json.NewEncoder(w).Encode(s)
	}
}

// @Summary Создать тип места
// @Description Создаёт новый тип места.
// @Tags Типы мест
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param seat_type body SeatTypeData true "Данные типа места"
// @Success 201 {object} CreateResponse "ID созданного типа места"
// @Failure 400 {object} ErrorResponse "В запросе предоставлены неверные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /seat-types [post]
func CreateSeatType(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var s SeatTypeData
		if !DecodeJSONBody(w, r, &s) {
			return
		}
		if !validateAllSeatTypeData(w, s) {
			return
		}

		id := uuid.New()
		_, err := db.Exec(context.Background(),
			"INSERT INTO seat_types (id, name, description) VALUES ($1, $2, $3)",
			id, s.Name, s.Description)

		if IsError(w, err) {
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(id.String())
	}
}

// @Summary Обновить тип места
// @Description Обновляет существующий тип места.
// @Tags Типы мест
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID типа места"
// @Param seat_type body SeatTypeData true "Обновлённые данные типа места"
// @Success 200 "Данные о типе места успешно обновлены"
// @Failure 400 {object} ErrorResponse "В запросе предоставлены неверные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Тип места не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /seat-types/{id} [put]
func UpdateSeatType(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var s SeatTypeData
		if !DecodeJSONBody(w, r, &s) {
			return
		}
		if !validateAllSeatTypeData(w, s) {
			return
		}

		res, err := db.Exec(context.Background(),
			"UPDATE seat_types SET name=$1, description=$2 WHERE id=$3",
			s.Name, s.Description, id)

		if IsError(w, err) {
			return
		}
		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// @Summary Удалить тип места
// @Description Удаляет тип места по его ID.
// @Tags Типы мест
// @Param id path string true "ID типа места"
// @Security BearerAuth
// @Success 204 "Данные о типе места успешно удалены"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Тип места не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /seat-types/{id} [delete]
func DeleteSeatType(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		res, err := db.Exec(context.Background(),
			"DELETE FROM seat_types WHERE id = $1", id)

		if IsError(w, err) {
			return
		}
		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// @Summary Поиск типов места по названию
// @Description Возвращает типы места, название которых содержит указанную строку.
// @Tags Типы мест
// @Produce json
// @Security BearerAuth
// @Param query query string true "Поисковый запрос"
// @Success 200 {array} SeatType "Список типов мест"
// @Failure 400 {object} ErrorResponse "Строка поиска пуста"
// @Failure 404 {object} ErrorResponse "Типы мест не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /seat-types/search [get]
func SearchSeatTypes(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		query = PrepareString(query)
		if query == "" {
			http.Error(w, "Строка поиска пуста", http.StatusBadRequest)
			return
		}

		rows, err := db.Query(context.Background(),
			"SELECT id, name, description FROM seat_types WHERE name ILIKE $1", "%"+query+"%")
		if IsError(w, err) {
			return
		}
		defer rows.Close()

		var types []SeatType
		for rows.Next() {
			var e SeatType
			if err := rows.Scan(&e.ID, &e.Name, &e.Description); IsError(w, err) {
				return
			}
			types = append(types, e)
		}

		if len(types) == 0 {
			http.Error(w, "Типы мест не найдены", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(types)
	}
}
