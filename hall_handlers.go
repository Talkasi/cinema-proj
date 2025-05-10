package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// @Summary Получить все залы
// @Tags halls
// @Produce json
// @Success 200 {array} Hall
// @Failure 500 {string} string "Ошибка сервера"
// @Router /halls [get]
func GetHalls(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, capacity, equipment_type_id, description FROM halls")
		if err != nil {
			http.Error(w, "Ошибка при получении залов", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var halls []Hall
		for rows.Next() {
			var h Hall
			if err := rows.Scan(&h.ID, &h.Name, &h.Capacity, &h.EquipmentType, &h.Description); err != nil {
				http.Error(w, "Ошибка при сканировании", http.StatusInternalServerError)
				return
			}
			halls = append(halls, h)
		}
		json.NewEncoder(w).Encode(halls)
	}
}

// @Summary Получить зал по ID
// @Tags halls
// @Produce json
// @Param id path string true "ID зала"
// @Success 200 {object} Hall
// @Failure 404 {string} string "Зал не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /halls/{id} [get]
func GetHallByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var h Hall
		err := db.QueryRow("SELECT id, name, capacity, equipment_type_id, description FROM halls WHERE id = $1", id).
			Scan(&h.ID, &h.Name, &h.Capacity, &h.EquipmentType, &h.Description)

		if err == sql.ErrNoRows {
			http.Error(w, "Зал не найден", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Ошибка при получении зала", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(h)
	}
}

// @Summary Создать зал
// @Tags halls
// @Accept json
// @Produce json
// @Param hall body Hall true "Новый зал"
// @Success 201 {object} Hall
// @Failure 400 {string} string "Неверный JSON"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /halls [post]
func CreateHall(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var h Hall
		if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		h.ID = uuid.New().String()

		_, err := db.Exec("INSERT INTO halls (id, name, capacity, equipment_type_id, description) VALUES ($1, $2, $3, $4, $5)",
			h.ID, h.Name, h.Capacity, h.EquipmentType, h.Description)
		if err != nil {
			http.Error(w, "Ошибка при создании зала", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(h)
	}
}

// @Summary Обновить зал
// @Tags halls
// @Accept json
// @Produce json
// @Param id path string true "ID зала"
// @Param hall body Hall true "Обновлённые данные"
// @Success 200 {object} Hall
// @Failure 404 {string} string "Зал не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /halls/{id} [put]
func UpdateHall(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var h Hall
		if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		h.ID = id

		res, err := db.Exec("UPDATE halls SET name=$1, capacity=$2, equipment_type_id=$3, description=$4 WHERE id=$5",
			h.Name, h.Capacity, h.EquipmentType, h.Description, h.ID)
		if err != nil {
			http.Error(w, "Ошибка при обновлении", http.StatusInternalServerError)
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			http.Error(w, "Зал не найден", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(h)
	}
}

// @Summary Удалить зал
// @Tags halls
// @Param id path string true "ID зала"
// @Success 204 {string} string "Удалено"
// @Failure 404 {string} string "Зал не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /halls/{id} [delete]
func DeleteHall(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		res, err := db.Exec("DELETE FROM halls WHERE id = $1", id)
		if err != nil {
			http.Error(w, "Ошибка при удалении", http.StatusInternalServerError)
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			http.Error(w, "Зал не найден", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
