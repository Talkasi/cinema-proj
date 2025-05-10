package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// @Summary Получить все места
// @Tags seats
// @Produce json
// @Success 200 {array} Seat
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seats [get]
func GetSeats(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, hall_id, seat_type_id, row_number, seat_number FROM seats")
		if err != nil {
			http.Error(w, "Ошибка при получении мест", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var seats []Seat
		for rows.Next() {
			var s Seat
			if err := rows.Scan(&s.ID, &s.HallID, &s.SeatTypeID, &s.RowNumber, &s.SeatNumber); err != nil {
				http.Error(w, "Ошибка при сканировании", http.StatusInternalServerError)
				return
			}
			seats = append(seats, s)
		}
		json.NewEncoder(w).Encode(seats)
	}
}

// @Summary Получить место по ID
// @Tags seats
// @Produce json
// @Param id path string true "ID места"
// @Success 200 {object} Seat
// @Failure 404 {string} string "Место не найдено"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seats/{id} [get]
func GetSeatByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		var s Seat
		err := db.QueryRow("SELECT id, hall_id, seat_type_id, row_number, seat_number FROM seats WHERE id = $1", id).
			Scan(&s.ID, &s.HallID, &s.SeatTypeID, &s.RowNumber, &s.SeatNumber)

		if err == sql.ErrNoRows {
			http.Error(w, "Место не найдено", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Ошибка при получении", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(s)
	}
}

// @Summary Создать место
// @Tags seats
// @Accept json
// @Produce json
// @Param seat body Seat true "Новое место"
// @Success 201 {object} Seat
// @Router /seats [post]
func CreateSeat(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var s Seat
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, "Неверный JSON", http.StatusBadRequest)
			return
		}
		s.ID = uuid.New().String()
		_, err := db.Exec(`INSERT INTO seats (id, hall_id, seat_type_id, row_number, seat_number) 
			VALUES ($1, $2, $3, $4, $5)`, s.ID, s.HallID, s.SeatTypeID, s.RowNumber, s.SeatNumber)
		if err != nil {
			http.Error(w, "Ошибка при создании", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(s)
	}
}

// @Summary Обновить место
// @Tags seats
// @Accept json
// @Produce json
// @Param id path string true "ID места"
// @Param seat body Seat true "Обновлённые данные места"
// @Success 200 {object} Seat
// @Failure 400 {string} string "Неверный JSON"
// @Failure 404 {string} string "Место не найдено"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seats/{id} [put]
func UpdateSeat(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		var s Seat

		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		s.ID = id

		res, err := db.Exec("UPDATE seats SET hall_id=$1, seat_type_id=$2, row_number=$3, seat_number=$4 WHERE id=$5",
			s.HallID, s.SeatTypeID, s.RowNumber, s.SeatNumber, s.ID)
		if err != nil {
			http.Error(w, "Ошибка при обновлении", http.StatusInternalServerError)
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			http.Error(w, "Место не найдено", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(s)
	}
}

// @Summary Удалить место
// @Tags seats
// @Param id path string true "ID места"
// @Success 204 {string} string "Удалено"
// @Failure 404 {string} string "Место не найдено"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seats/{id} [delete]
func DeleteSeat(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		res, err := db.Exec("DELETE FROM seats WHERE id = $1", id)
		if err != nil {
			http.Error(w, "Ошибка при удалении", http.StatusInternalServerError)
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			http.Error(w, "Место не найдено", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
