package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func validateTicketData(w http.ResponseWriter, t TicketData) bool {
	if err := validateTicketStatus(t.Status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	if t.Price < 0 {
		http.Error(w, "Цена не может быть отрицательной", http.StatusBadRequest)
		return false
	}

	if err := uuid.Validate(t.MovieShowID); err != nil {
		http.Error(w, "Неверный формат ID сеанса", http.StatusBadRequest)
		return false
	}

	if err := uuid.Validate(t.SeatID); err != nil {
		http.Error(w, "Неверный формат ID места", http.StatusBadRequest)
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

// @Summary Получить все билеты для сеанса фильма по ID (guest | user | admin) ПОДУМАТЬ
// @Description Возвращает список всех билетов по ID сеанаса, содержащихся в базе данных.
// @Tags Билеты
// @Produce json
// @Param movie_show_id path string true "ID показа фильма"
// @Success 200 {array} Ticket
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /tickets/movie-show/{movie_show_id} [get]
func GetTicketsByMovieShowID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		movieShowID, ok := ParseUUIDFromPath(w, r.PathValue("movie_show_id"))
		if !ok {
			return
		}

		rows, err := db.Query(context.Background(), `
			SELECT t.id, t.movie_show_id, t.seat_id, t.ticket_status, t.price, t.user_id
			FROM tickets t
			WHERE t.movie_show_id = $1`, movieShowID)
		if HandleDatabaseError(w, err, "билетами") {
			return
		}
		defer rows.Close()

		var tickets []Ticket
		for rows.Next() {
			var t Ticket
			if err := rows.Scan(&t.ID, &t.MovieShowID, &t.SeatID, &t.Status, &t.Price, &t.UserID); HandleDatabaseError(w, err, "билетом") {
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

// @Summary Получить билет по ID (guest | user | admin) ПОДУМАТЬ
// @Description Возвращает билет по ID.
// @Tags Билеты
// @Produce json
// @Param id path string true "ID билета"
// @Success 200 {object} Ticket "Билет"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 404 {object} ErrorResponse "Билет не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /tickets/{id} [get]
func GetTicketByID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}
		var t Ticket
		err := db.QueryRow(context.Background(), "SELECT id, movie_show_id, seat_id, ticket_status, price, user_id FROM tickets WHERE id = $1", id).
			Scan(&t.ID, &t.MovieShowID, &t.SeatID, &t.Status, &t.Price, &t.UserID)

		if IsError(w, err) {
			return
		}
		json.NewEncoder(w).Encode(t)
	}
}

// @Summary Создать билет (admin)
// @Description Создаёт новый билет.
// @Tags Билеты
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ticket body TicketData true "Билет"
// @Success 201 {object} CreateResponse "ID созданного билета"
// @Failure 400 {object} ErrorResponse "В запросе предоставлены неверные данные"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /tickets [post]
func CreateTicket(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var t TicketData
		if !DecodeJSONBody(w, r, &t) {
			return
		}

		if !validateTicketData(w, t) {
			return
		}

		id := uuid.New()
		_, err := db.Exec(context.Background(), "INSERT INTO tickets (id, movie_show_id, seat_id, ticket_status, price, user_id) VALUES ($1, $2, $3, $4, $5, $6)",
			id, t.MovieShowID, t.SeatID, t.Status, t.Price, t.UserID)
		if IsError(w, err) {
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(id.String())
	}
}

// @Summary Обновить билет (user | admin)
// @Description Обновляет существующий билет.
// @Tags Билеты
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID билета"
// @Param ticket body TicketData true "Обновлённые данные билета"
// @Success 200 "Данные о билете успешно обновлены"
// @Failure 404 {object} ErrorResponse "Билет не найден"
// @Failure 400 {object} ErrorResponse "Неверный формат JSON"
// @Failure 500 {object} ErrorResponse "Ошибка"
// @Router /tickets/{id} [put]
func UpdateTicket(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var t TicketData
		if !DecodeJSONBody(w, r, &t) {
			return
		}

		if !validateTicketData(w, t) {
			return
		}

		res, err := db.Exec(context.Background(), "UPDATE tickets SET movie_show_id=$1, seat_id=$2, ticket_status=$3, price=$4, user_id=$5 WHERE id=$6",
			t.MovieShowID, t.SeatID, t.Status, t.Price, t.UserID, id)
		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}
		json.NewEncoder(w)
	}
}

// @Summary Удалить билет (admin)
// @Description Удаляет билет по ID.
// @Tags Билеты
// @Security BearerAuth
// @Param id path string true "ID билета"
// @Success 204 "Данные о билете успешно удалены"
// @Failure 400 {object} ErrorResponse "Неверный формат ID"
// @Failure 403 {object} ErrorResponse "Доступ запрещён"
// @Failure 404 {object} ErrorResponse "Билет не найден"
// @Failure 500 {object} ErrorResponse "Ошибка сервера"
// @Router /tickets/{id} [delete]
func DeleteTicket(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

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

// @Summary Получить билеты пользователя (user* | admin)
// @Tags Билеты
// @Produce json
// @Param user_id path string true "ID пользователя"
// @Success 200 {array} Ticket
// @Failure 500 {object} ErrorResponse "Ошибка"
// @Router /tickets/user/{user_id} [get]
func GetTicketsByUserID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := ParseUUIDFromPath(w, r.PathValue("user_id"))
		if !ok {
			return
		}

		role := r.Header.Get("Role")
		user_id := r.Header.Get("UserID")
		if (role != os.Getenv("CLAIM_ROLE_ADMIN")) && (userID.String() != user_id) {
			http.Error(w, "Доступ запрещён", http.StatusForbidden)
			return
		}

		rows, err := db.Query(context.Background(), `
			SELECT id, movie_show_id, seat_id, user_id, ticket_status, price
			FROM tickets
			WHERE user_id = $1`, userID)
		if IsError(w, err) {
			return
		}
		defer rows.Close()

		var tickets []Ticket
		for rows.Next() {
			var t Ticket
			if err := rows.Scan(&t.ID, &t.MovieShowID, &t.SeatID, &t.UserID, &t.Status, &t.Price); err != nil {
				println(err.Error())
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
