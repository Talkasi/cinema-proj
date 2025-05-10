package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// @Summary Получить все типы оборудования
// @Tags equipment-types
// @Produce json
// @Security BearerAuth
// @Success 200 {array} EquipmentType
// @Failure 500 {string} string "Ошибка сервера"
// @Router /equipment-types [get]
func GetEquipmentTypes(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, description FROM equipment_types")
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка при получении типов оборудования %v", err), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var types []EquipmentType
		for rows.Next() {
			var e EquipmentType
			if err := rows.Scan(&e.ID, &e.Name, &e.Description); err != nil {
				http.Error(w, "Ошибка при сканировании", http.StatusInternalServerError)
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
// @Tags equipment-types
// @Produce json
// @Param id path string true "ID типа оборудования"
// @Security BearerAuth
// @Success 200 {object} EquipmentType
// @Failure 404 {string} string "Тип оборудования не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /equipment-types/{id} [get]
func GetEquipmentTypeByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if _, err := uuid.Parse(id); err != nil {
			http.Error(w, "Неверный формат UUID", http.StatusBadRequest)
			return
		}

		var e EquipmentType
		err := db.QueryRow("SELECT id, name, description FROM equipment_types WHERE id = $1", id).
			Scan(&e.ID, &e.Name, &e.Description)

		if err == sql.ErrNoRows {
			http.Error(w, "Тип оборудования не найден", http.StatusNotFound)
			return
		} else if err != nil {
			if strings.Contains(err.Error(), "permission denied") {
				http.Error(w, "Доступ запрещен", http.StatusForbidden)
				return
			}

			http.Error(w, "Ошибка при получении", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(e)
	}
}

// @Summary Создать тип оборудования
// @Tags equipment-types
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param equipment_type body EquipmentType true "Тип оборудования"
// @Success 201 {object} EquipmentType
// @Failure 400 {string} string "Неверный запрос"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /equipment-types [post]
func CreateEquipmentType(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var e EquipmentType
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		e.ID = uuid.New().String()

		_, err := db.Exec("INSERT INTO equipment_types (id, name, description) VALUES ($1, $2, $3)",
			e.ID, e.Name, e.Description)

		if err != nil {
			if strings.Contains(err.Error(), "permission denied") {
				http.Error(w, "Доступ запрещен", http.StatusForbidden)
				return
			}

			http.Error(w, "Ошибка при вставке", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(e)
	}
}

// @Summary Обновить тип оборудования
// @Tags equipment-types
// @Accept json
// @Produce json
// @Param id path string true "ID типа оборудования"
// @Param equipment_type body EquipmentType true "Обновлённые данные типа"
// @Security BearerAuth
// @Success 200 {object} EquipmentType
// @Failure 400 {string} string "Неверный JSON"
// @Failure 404 {string} string "Тип не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /equipment-types/{id} [put]
func UpdateEquipmentType(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if _, err := uuid.Parse(id); err != nil {
			http.Error(w, "Неверный формат UUID", http.StatusBadRequest)
			return
		}

		var e EquipmentType
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		e.ID = id

		res, err := db.Exec("UPDATE equipment_types SET name=$1, description=$2 WHERE id=$3", e.Name, e.Description, e.ID)
		if err != nil {
			if strings.Contains(err.Error(), "permission denied") {
				http.Error(w, "Доступ запрещен", http.StatusForbidden)
				return
			}

			http.Error(w, "Ошибка при обновлении", http.StatusInternalServerError)
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			http.Error(w, "Тип оборудования не найден", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(e)
	}
}

// @Summary Удалить тип оборудования
// @Tags equipment-types
// @Param id path string true "ID типа оборудования"
// @Security BearerAuth
// @Success 204 {string} string "Удалено"
// @Failure 404 {string} string "Тип не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /equipment-types/{id} [delete]
func DeleteEquipmentType(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if _, err := uuid.Parse(id); err != nil {
			http.Error(w, "Неверный формат UUID", http.StatusBadRequest)
			return
		}

		res, err := db.Exec("DELETE FROM equipment_types WHERE id = $1", id)
		if err != nil {
			if strings.Contains(err.Error(), "permission denied") {
				http.Error(w, "Доступ запрещен", http.StatusForbidden)
				return
			}

			http.Error(w, "Ошибка при удалении", http.StatusInternalServerError)
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			http.Error(w, "Тип оборудования не найден", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
