package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// @Summary Получить все залы
// @Description Возвращает список всех залов
// @Tags halls
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Hall "Список залов"
// @Failure 404 {string} string "Залы не найдены"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /halls [get]
func GetHalls(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(context.Background(),
			"SELECT id, name, capacity, equipment_type_id, description FROM halls")
		if HandleDatabaseError(w, err, "залами") {
			return
		}
		defer rows.Close()

		var halls []Hall
		for rows.Next() {
			var h Hall
			if err := rows.Scan(&h.ID, &h.Name, &h.Capacity, &h.EquipmentTypeID, &h.Description); HandleDatabaseError(w, err, "залом") {
				return
			}
			halls = append(halls, h)
		}

		if len(halls) == 0 {
			http.Error(w, "Залы не найдены", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(halls)
	}
}

// @Summary Получить зал по ID
// @Description Возвращает зал по его UUID
// @Tags halls
// @Produce json
// @Param id path string true "UUID зала"
// @Security BearerAuth
// @Success 200 {object} Hall "Данные зала"
// @Failure 400 {string} string "Неверный формат UUID"
// @Failure 404 {string} string "Зал не найден"
// @Failure 500 {string} string "Ошибка сервера"
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
			"SELECT name, capacity, equipment_type_id, description FROM halls WHERE id = $1", id).
			Scan(&h.Name, &h.Capacity, &h.EquipmentTypeID, &h.Description)

		if IsError(w, err) {
			return
		}

		json.NewEncoder(w).Encode(h)
	}
}

// @Summary Создать зал
// @Description Создает новый зал
// @Tags halls
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param hall body HallData true "Данные зала"
// @Success 201 {object} string "UUID созданного зала"
// @Failure 400 {string} string "Неверный формат JSON"
// @Failure 403 {string} string "Доступ запрещен"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /halls [post]
func CreateHall(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var h HallData
		if !DecodeJSONBody(w, r, &h) {
			return
		}

		if !ValidateRequiredFields(w, map[string]string{
			"name":              h.Name,
			"capacity":          fmt.Sprint(h.Capacity),
			"description":       h.Description,
			"equipment_type_id": h.EquipmentTypeID,
		}) {
			return
		}

		id := uuid.New().String()
		_, err := db.Exec(context.Background(),
			"INSERT INTO halls (id, name, capacity, equipment_type_id, description) VALUES ($1, $2, $3, $4, $5)",
			id, h.Name, h.Capacity, h.EquipmentTypeID, h.Description)

		if IsError(w, err) {
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(id)
	}
}

// @Summary Обновить зал
// @Description Обновляет существующий зал
// @Tags halls
// @Accept json
// @Produce json
// @Param id path string true "UUID зала"
// @Param hall body HallData true "Обновленные данные зала"
// @Security BearerAuth
// @Success 200 "Данные зала успешно обновлены"
// @Failure 400 {string} string "Неверный формат UUID/JSON или пустые поля"
// @Failure 403 {string} string "Доступ запрещен"
// @Failure 404 {string} string "Зал не найден"
// @Failure 500 {string} string "Ошибка сервера"
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

		if !ValidateRequiredFields(w, map[string]string{
			"name":              h.Name,
			"capacity":          fmt.Sprint(h.Capacity),
			"description":       h.Description,
			"equipment_type_id": h.EquipmentTypeID,
		}) {
			return
		}

		res, err := db.Exec(context.Background(),
			"UPDATE halls SET name=$1, capacity=$2, equipment_type_id=$3, description=$4 WHERE id=$5",
			h.Name, h.Capacity, h.EquipmentTypeID, h.Description, id)

		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		json.NewEncoder(w)
	}
}

// @Summary Удалить зал
// @Description Удаляет зал по его UUID
// @Tags halls
// @Param id path string true "UUID зала"
// @Security BearerAuth
// @Success 204 "Зал успешно удален"
// @Failure 400 {string} string "Неверный формат UUID"
// @Failure 403 {string} string "Доступ запрещен"
// @Failure 404 {string} string "Зал не найден"
// @Failure 500 {string} string "Ошибка сервера"
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
