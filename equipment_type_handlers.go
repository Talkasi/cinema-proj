package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// @Summary Получить все типы оборудования
// @Description Возвращает список всех типов оборудования
// @Tags equipment-types
// @Produce json
// @Security BearerAuth
// @Success 200 {array} EquipmentType "Список типов оборудования"
// @Failure 404 {string} string "Типы оборудования не найдены"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /equipment-types [get]
func GetEquipmentTypes(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(context.Background(), "SELECT id, name, description FROM equipment_types")
		if HandleDatabaseError(w, err, "типами оборудования") {
			return
		}
		defer rows.Close()

		var types []EquipmentType
		for rows.Next() {
			var e EquipmentType
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
// @Description Возвращает тип оборудования по его UUID
// @Tags equipment-types
// @Produce json
// @Param id path string true "UUID типа оборудования"
// @Security BearerAuth
// @Success 200 {object} EquipmentType "Тип оборудования"
// @Failure 400 {string} string "Неверный формат UUID"
// @Failure 404 {string} string "Тип оборудования не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /equipment-types/{id} [get]
func GetEquipmentTypeByID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var e EquipmentType
		e.ID = id.String()
		err := db.QueryRow(context.Background(),
			"SELECT name, description FROM equipment_types WHERE id = $1", id).
			Scan(&e.Name, &e.Description)

		if IsError(w, err) {
			return
		}

		json.NewEncoder(w).Encode(e)
	}
}

// @Summary Создать тип оборудования
// @Description Создает новый тип оборудования
// @Tags equipment-types
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param equipment_type body EquipmentTypeData true "Данные типа оборудования"
// @Success 201 {object} string "UUID созданного типа оборудования"
// @Failure 400 {string} string "Неверный формат JSON"
// @Failure 403 {string} string "Доступ запрещен"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /equipment-types [post]
func CreateEquipmentType(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var e EquipmentTypeData
		if !DecodeJSONBody(w, r, &e) {
			return
		}
		if !ValidateRequiredFields(w, map[string]string{
			"name":        e.Name,
			"description": e.Description,
		}) {
			return
		}

		id := uuid.New()
		_, err := db.Exec(context.Background(),
			"INSERT INTO equipment_types (id, name, description) VALUES ($1, $2, $3)",
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
// @Tags equipment-types
// @Accept json
// @Produce json
// @Param id path string true "UUID типа оборудования"
// @Param equipment_type body EquipmentTypeData true "Обновленные данные типа оборудования"
// @Security BearerAuth
// @Success 200 "Тип оборудования успешно обновлен"
// @Failure 400 {string} string "Неверный формат UUID/JSON или пустые поля"
// @Failure 403 {string} string "Доступ запрещен"
// @Failure 404 {string} string "Тип оборудования не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /equipment-types/{id} [put]
func UpdateEquipmentType(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var e EquipmentTypeData
		if !DecodeJSONBody(w, r, &e) {
			return
		}
		if !ValidateRequiredFields(w, map[string]string{
			"name":        e.Name,
			"description": e.Description,
		}) {
			return
		}

		res, err := db.Exec(context.Background(),
			"UPDATE equipment_types SET name=$1, description=$2 WHERE id=$3",
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
// @Description Удаляет тип оборудования по его UUID
// @Tags equipment-types
// @Param id path string true "UUID типа оборудования"
// @Security BearerAuth
// @Success 204 "Тип оборудования успешно удален"
// @Failure 400 {string} string "Неверный формат UUID"
// @Failure 403 {string} string "Доступ запрещен"
// @Failure 404 {string} string "Тип оборудования не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /equipment-types/{id} [delete]
func DeleteEquipmentType(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		res, err := db.Exec(context.Background(),
			"DELETE FROM equipment_types WHERE id = $1", id)
		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
