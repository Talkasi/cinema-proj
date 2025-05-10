package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// @Summary Создать билет
// @Tags tickets
// @Accept json
// @Produce json
// @Param ticket body Ticket true "Билет"
// @Success 201 {object} Ticket
// @Failure 400 {string} string "Неверный JSON"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /tickets [post]
func CreateTicket(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var t Ticket
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		t.ID = uuid.New().String()

		_, err := db.Exec("INSERT INTO tickets (id, movie_show_id, seat_id, ticket_status_id, price) VALUES ($1, $2, $3, $4, $5)",
			t.ID, t.MovieShowID, t.SeatID, t.StatusID, t.Price)
		if err != nil {
			http.Error(w, "Ошибка при создании", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(t)
	}
}

// @Summary Получить все билеты для показа фильма по ID
// @Tags seats
// @Produce json
// @Param movie_show_id path string true "ID показа фильма"
// @Success 200 {array} Ticket
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seats/{movie_show_id} [get]
func GetTicketsByMovieShowID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		movieShowID := r.PathValue("movie_show_id")

		rows, err := db.Query(`
			SELECT t.id, t.movie_show_id, t.seat_id, t.ticket_status_id, t.price
			FROM tickets t
			WHERE t.movie_show_id = ?`, movieShowID)
		if err != nil {
			http.Error(w, "Ошибка при получении билетов", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var tickets []Ticket
		for rows.Next() {
			var t Ticket
			if err := rows.Scan(&t.ID, &t.MovieShowID, &t.SeatID, &t.StatusID, &t.Price); err != nil {
				http.Error(w, "Ошибка при сканировании билета", http.StatusInternalServerError)
				return
			}
			tickets = append(tickets, t)
		}
		json.NewEncoder(w).Encode(tickets)
	}
}

// @Summary Получить билет по ID
// @Tags tickets
// @Produce json
// @Param id path string true "ID билета"
// @Success 200 {object} Ticket
// @Failure 404 {string} string "Билет не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /tickets/{id} [get]
func GetTicketByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var t Ticket
		err := db.QueryRow("SELECT id, movie_show_id, seat_id, ticket_status_id, price FROM tickets WHERE id = $1", id).
			Scan(&t.ID, &t.MovieShowID, &t.SeatID, &t.StatusID, &t.Price)

		if err == sql.ErrNoRows {
			http.Error(w, "Билет не найден", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Ошибка при получении", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(t)
	}
}

// @Summary Обновить билет
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path string true "ID билета"
// @Param ticket body Ticket true "Обновлённый билет"
// @Success 200 {object} Ticket
// @Failure 404 {string} string "Билет не найден"
// @Failure 500 {string} string "Ошибка"
// @Router /tickets/{id} [put]
func UpdateTicket(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var t Ticket
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		t.ID = id

		res, err := db.Exec("UPDATE tickets SET movie_show_id=$1, seat_id=$2, ticket_status_id=$3, price=$4 WHERE id=$5",
			t.MovieShowID, t.SeatID, t.StatusID, t.Price, t.ID)
		if err != nil {
			http.Error(w, "Ошибка при обновлении", http.StatusInternalServerError)
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			http.Error(w, "Билет не найден", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(t)
	}
}

// @Summary Удалить билет
// @Tags tickets
// @Param id path string true "ID билета"
// @Success 204 {string} string "Удалено"
// @Failure 404 {string} string "Не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /tickets/{id} [delete]
func DeleteTicket(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		res, err := db.Exec("DELETE FROM tickets WHERE id = $1", id)
		if err != nil {
			http.Error(w, "Ошибка при удалении", http.StatusInternalServerError)
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			http.Error(w, "Билет не найден", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
