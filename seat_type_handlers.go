package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// @Summary Получить все типы мест
// @Tags seat-types
// @Produce json
// @Success 200 {array} SeatType
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seat-types [get]
func GetSeatTypes(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(context.Background(), "SELECT id, name, description FROM seat_types")
		if err != nil {
			http.Error(w, "Ошибка при получении типов мест", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var seatTypes []SeatType
		for rows.Next() {
			var st SeatType
			if err := rows.Scan(&st.ID, &st.Name, &st.Description); err != nil {
				http.Error(w, "Ошибка при сканировании", http.StatusInternalServerError)
				return
			}
			seatTypes = append(seatTypes, st)
		}
		json.NewEncoder(w).Encode(seatTypes)
	}
}

// @Summary Получить тип места по ID
// @Tags seat-types
// @Produce json
// @Param id path string true "ID типа места"
// @Success 200 {object} SeatType
// @Failure 404 {string} string "Тип места не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seat-types/{id} [get]
func GetSeatTypeByID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		var st SeatType
		err := db.QueryRow(context.Background(), "SELECT id, name, description FROM seat_types WHERE id = $1", id).
			Scan(&st.ID, &st.Name, &st.Description)

		if err == sql.ErrNoRows {
			http.Error(w, "Тип места не найден", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Ошибка при получении", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(st)
	}
}

// @Summary Создать тип места
// @Tags seat-types
// @Accept json
// @Produce json
// @Param seat_type body SeatType true "Тип места"
// @Success 201 {object} SeatType
// @Failure 400 {string} string "Неверный запрос"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seat-types [post]
func CreateSeatType(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var st SeatType
		if err := json.NewDecoder(r.Body).Decode(&st); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		st.ID = uuid.New().String()

		_, err := db.Exec(context.Background(), "INSERT INTO seat_types (id, name, description) VALUES ($1, $2, $3)",
			st.ID, st.Name, st.Description)
		if err != nil {
			http.Error(w, "Ошибка при вставке", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(st)
	}
}

// @Summary Обновить тип места
// @Tags seat-types
// @Accept json
// @Produce json
// @Param id path string true "ID типа места"
// @Param seat_type body SeatType true "Обновлённые данные типа"
// @Success 200 {object} SeatType
// @Failure 400 {string} string "Неверный JSON"
// @Failure 404 {string} string "Тип не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seat-types/{id} [put]
func UpdateSeatType(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		var st SeatType
		if err := json.NewDecoder(r.Body).Decode(&st); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		st.ID = id

		res, err := db.Exec(context.Background(), "UPDATE seat_types SET name=$1, description=$2 WHERE id=$3",
			st.Name, st.Description, st.ID)
		if err != nil {
			http.Error(w, "Ошибка при обновлении", http.StatusInternalServerError)
			return
		}
		rows := res.RowsAffected()
		if rows == 0 {
			http.Error(w, "Тип места не найден", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(st)
	}
}

// @Summary Удалить тип места
// @Tags seat-types
// @Param id path string true "ID типа места"
// @Success 204 {string} string "Удалено"
// @Failure 404 {string} string "Тип не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seat-types/{id} [delete]
func DeleteSeatType(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		res, err := db.Exec(context.Background(), "DELETE FROM seat_types WHERE id = $1", id)
		if err != nil {
			http.Error(w, "Ошибка при удалении", http.StatusInternalServerError)
			return
		}
		rows := res.RowsAffected()
		if rows == 0 {
			http.Error(w, "Тип места не найден", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
