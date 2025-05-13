package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func validateTicketData(w http.ResponseWriter, t Ticket) bool {
	if err := validateTicketStatus(t.Status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	if t.Price < 0 {
		http.Error(w, "Цена не может быть отрицательной", http.StatusBadRequest)
		return false
	}
	return true
}

func validateTicketStatus(status TicketStatusEnumType) error {
	if !status.IsValid() {
		return errors.New("недопустимый статус билета")
	}
	return nil
}

// @Summary Создать билет
// @Tags tickets
// @Accept json
// @Produce json
// @Param ticket body Ticket true "Билет"
// @Success 201 {object} Ticket
// @Failure 400 {object} ErrorResponse "Неверный JSON"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /tickets [post]
func CreateTicket(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var t Ticket
		if !DecodeJSONBody(w, r, &t) {
			return
		}
		t.ID = uuid.New().String()

		if !validateTicketData(w, t) {
			return
		}

		_, err := db.Exec(context.Background(), "INSERT INTO tickets (id, movie_show_id, seat_id, ticket_status_id, price) VALUES ($1, $2, $3, $4, $5)",
			t.ID, t.MovieShowID, t.SeatID, t.Status, t.Price)
		if IsError(w, err) {
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(t)
	}
}

// @Summary Получить все билеты для показа фильма по ID
// @Tags tickets
// @Produce json
// @Param movie_show_id path string true "ID показа фильма"
// @Success 200 {array} Ticket
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /seats/{movie_show_id} [get]
func GetTicketsByMovieShowID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		movieShowID := r.PathValue("movie_show_id")

		rows, err := db.Query(context.Background(), `
			SELECT t.id, t.movie_show_id, t.seat_id, t.ticket_status_id, t.price
			FROM tickets t
			WHERE t.movie_show_id = $1`, movieShowID)
		if HandleDatabaseError(w, err, "билетами") {
			return
		}
		defer rows.Close()

		var tickets []Ticket
		for rows.Next() {
			var t Ticket
			if err := rows.Scan(&t.ID, &t.MovieShowID, &t.SeatID, &t.Status, &t.Price); HandleDatabaseError(w, err, "билетом") {
				return
			}
			tickets = append(tickets, t)
		}

		if len(tickets) == 0 {
			http.Error(w, "Билеты не найдены", http.StatusNotFound)
		}

		json.NewEncoder(w).Encode(tickets)
	}
}

// @Summary Получить билет по ID
// @Tags tickets
// @Produce json
// @Param id path string true "ID билета"
// @Success 200 {object} Ticket
// @Failure 404 {object} ErrorResponse "Билет не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /tickets/{id} [get]
func GetTicketByID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var t Ticket
		err := db.QueryRow(context.Background(), "SELECT id, movie_show_id, seat_id, ticket_status_id, price FROM tickets WHERE id = $1", id).
			Scan(&t.ID, &t.MovieShowID, &t.SeatID, &t.Status, &t.Price)

		if IsError(w, err) {
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
// @Failure 404 {object} ErrorResponse "Билет не найден"
// @Failure 400 {object} ErrorResponse "Неверный формат JSON"
// @Failure 500 {object} ErrorResponse "Ошибка"
// @Router /tickets/{id} [put]
func UpdateTicket(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var t Ticket
		if !DecodeJSONBody(w, r, &t) {
			return
		}
		t.ID = id

		if !validateTicketData(w, t) {
			return
		}

		res, err := db.Exec(context.Background(), "UPDATE tickets SET movie_show_id=$1, seat_id=$2, ticket_status_id=$3, price=$4 WHERE id=$5",
			t.MovieShowID, t.SeatID, t.Status, t.Price, t.ID)
		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}
		json.NewEncoder(w).Encode(t)
	}
}

// @Summary Удалить билет
// @Tags tickets
// @Param id path string true "ID билета"
// @Success 204 {string} string "Удалено"
// @Failure 404 {object} ErrorResponse "Не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /tickets/{id} [delete]
func DeleteTicket(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		res, err := db.Exec(context.Background(), "DELETE FROM tickets WHERE id = $1", id)
		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// @Summary Получить билеты пользователя
// @Tags tickets
// @Produce json
// @Param user_id path string true "ID пользователя"
// @Success 200 {array} Ticket
// @Failure 500 {object} ErrorResponse "Ошибка"
// @Router /tickets/user/{user_id} [get]
func GetTicketsByUserID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("user_id")

		rows, err := db.Query(context.Background(), `
			SELECT id, movie_show_id, seat_id, ticket_status_id, price
			FROM tickets
			WHERE user_id = $1`, userID)
		if IsError(w, err) {
			return
		}
		defer rows.Close()

		var tickets []Ticket
		for rows.Next() {
			var t Ticket
			if err := rows.Scan(&t.ID, &t.MovieShowID, &t.SeatID, &t.Status, &t.Price); err != nil {
				http.Error(w, "Ошибка при сканировании", http.StatusInternalServerError)
				return
			}
			tickets = append(tickets, t)
		}

		if len(tickets) == 0 {
			http.Error(w, "Билеты не найдены", http.StatusNotFound)
		}

		json.NewEncoder(w).Encode(tickets)
	}
}
