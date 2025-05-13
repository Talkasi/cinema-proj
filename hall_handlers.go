package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func validateAllHallData(w http.ResponseWriter, h HallData) (bool, HallData) {
	var err error
	h.Name, err = PrepareString(h.Name, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка валидации имени: %v", err.Error()), http.StatusBadRequest)
		return false, h
	}

	h.Description, err = PrepareString(h.Description, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка валидации описания: %v", err.Error()), http.StatusBadRequest)
		return false, h
	}

	if err := validateHallName(h.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false, h
	}

	if err := validateHallCapacity(h.Capacity); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false, h
	}

	if err := validateScreenTypeID(h.ScreenTypeID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false, h
	}

	if err := validateHallDescription(h.Description); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false, h
	}

	return true, h
}

func validateHallName(name string) error {
	if !regexp.MustCompile(`^[a-zA-Zа-яА-Я0-9\s\.\-_#№]+$`).MatchString(name) {
		return errors.New("название зала содержит запрещённые символы. Разрешены буквы, цифры, пробелы, точки, дефисы, подчёркивания, # и №")
	}
	if len(name) > 100 {
		return errors.New("название зала не может превышать 100 символов")
	}
	return nil
}

func validateHallCapacity(capacity int) error {
	if capacity <= 0 {
		return errors.New("вместимость зала должна быть положительным числом")
	}
	return nil
}

func validateScreenTypeID(screenTypeID string) error {
	if _, err := uuid.Parse(screenTypeID); err != nil {
		return errors.New("неверный формат UUID типа экрана")
	}
	return nil
}

func validateHallDescription(description string) error {
	if description != "" && len(description) > 1000 {
		return errors.New("описание зала не может превышать 1000 символов")
	}
	return nil
}

// @Summary Получить все кинозалы
// @Description Возвращает список всех кинозалов, содержащихся в базе данных.
// @Tags Кинозалы
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Hall "Список кинозалов"
// @Failure 404 {object} ErrorResponse "Данные не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /halls [get]
func GetHalls(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(context.Background(),
			"SELECT id, name, capacity, screen_type_id, description FROM halls")
		if HandleDatabaseError(w, err, "залами") {
			return
		}
		defer rows.Close()

		var halls []Hall
		for rows.Next() {
			var h Hall
			if err := rows.Scan(&h.ID, &h.Name, &h.Capacity, &h.ScreenTypeID, &h.Description); HandleDatabaseError(w, err, "залом") {
				return
			}
			halls = append(halls, h)
		}

		if len(halls) == 0 {
			http.Error(w, "Залы не найдены", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(halls)
	}
}

// @Summary Получить кинозал по ID
// @Description Возвращает кинозал по его ID.
// @Tags Кинозалы
// @Produce json
// @Param id path string true "ID зала"
// @Security BearerAuth
// @Success 200 {object} Hall "Данные кинозала"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 404 {object} ErrorResponse "Данные не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /halls/{id} [get]
func GetHallByID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var h Hall
		h.ID = id.String()
		err := db.QueryRow(context.Background(),
			"SELECT name, capacity, screen_type_id, description FROM halls WHERE id = $1", id).
			Scan(&h.Name, &h.Capacity, &h.ScreenTypeID, &h.Description)

		if IsError(w, err) {
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(h)
	}
}

// @Summary Создать кинозал
// @Description Создаёт новый кинозал.
// @Tags Кинозалы
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param hall body HallData true "Данные кинозала"
// @Success 201 {object} CreateResponse "ID созданного кинозала"
// @Failure 400 {object} ErrorResponse "В запросе предоставлены неверные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /halls [post]
func CreateHall(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var h HallData
		if !DecodeJSONBody(w, r, &h) {
			return
		}

		var ok bool
		if ok, h = validateAllHallData(w, h); !ok {
			return
		}

		id := uuid.New()
		_, err := db.Exec(context.Background(),
			"INSERT INTO halls (id, name, capacity, screen_type_id, description) VALUES ($1, $2, $3, $4, $5)",
			id, h.Name, h.Capacity, h.ScreenTypeID, h.Description)

		if IsError(w, err) {
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(id.String())
	}
}

// @Summary Обновить кинозал
// @Description Обновляет существующий кинозал.
// @Tags Кинозалы
// @Accept json
// @Produce json
// @Param id path string true "ID зала"
// @Param hall body HallData true "Обновлённые данные зала"
// @Security BearerAuth
// @Success 200 "Данные о кинозале успешно обновлены"
// @Failure 400 {object} ErrorResponse "В запросе предоставлены неверные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Зал не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /halls/{id} [put]
func UpdateHall(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var h HallData
		if !DecodeJSONBody(w, r, &h) {
			return
		}

		if ok, h = validateAllHallData(w, h); !ok {
			return
		}

		res, err := db.Exec(context.Background(),
			"UPDATE halls SET name=$1, capacity=$2, screen_type_id=$3, description=$4 WHERE id=$5",
			h.Name, h.Capacity, h.ScreenTypeID, h.Description, id)

		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		updatedHall := Hall{
			ID:           id.String(),
			Name:         h.Name,
			Capacity:     h.Capacity,
			ScreenTypeID: h.ScreenTypeID,
			Description:  h.Description,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updatedHall)
	}
}

// @Summary Удалить кинозал
// @Description Удаляет кинозал по его ID.
// @Tags Кинозалы
// @Param id path string true "ID кинозала"
// @Security BearerAuth
// @Success 204 "Данные о кинозале успешно удалены"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Данные не найдены"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /halls/{id} [delete]
func DeleteHall(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		res, err := db.Exec(context.Background(),
			"DELETE FROM halls WHERE id = $1", id)

		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
