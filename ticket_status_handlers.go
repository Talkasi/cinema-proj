package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// @Summary Получить все статусы билетов
// @Tags ticket-statuses
// @Produce json
// @Success 200 {array} TicketStatus
// @Failure 500 {string} string "Ошибка сервера"
// @Router /ticket-statuses [get]
func GetTicketStatuses(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(context.Background(), "SELECT id, name FROM ticket_statuses")
		if err != nil {
			http.Error(w, "Ошибка при получении статусов билетов", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var statuses []TicketStatus
		for rows.Next() {
			var ts TicketStatus
			if err := rows.Scan(&ts.ID, &ts.Name); err != nil {
				http.Error(w, "Ошибка при сканировании", http.StatusInternalServerError)
				return
			}
			statuses = append(statuses, ts)
		}
		json.NewEncoder(w).Encode(statuses)
	}
}

// @Summary Получить статус билета по ID
// @Tags ticket-statuses
// @Produce json
// @Param id path string true "ID статуса билета"
// @Success 200 {object} TicketStatus
// @Failure 404 {string} string "Статус билета не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /ticket-statuses/{id} [get]
func GetTicketStatusByID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		var ts TicketStatus
		err := db.QueryRow(context.Background(), "SELECT id, name FROM ticket_statuses WHERE id = $1", id).
			Scan(&ts.ID, &ts.Name)

		if err == sql.ErrNoRows {
			http.Error(w, "Статус билета не найден", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Ошибка при получении", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(ts)
	}
}

// @Summary Создать статус билета
// @Tags ticket-statuses
// @Accept json
// @Produce json
// @Param ticket_status body TicketStatus true "Статус билета"
// @Success 201 {object} TicketStatus
// @Failure 400 {string} string "Неверный запрос"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /ticket-statuses [post]
func CreateTicketStatus(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ts TicketStatus
		if err := json.NewDecoder(r.Body).Decode(&ts); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		ts.ID = uuid.New().String()

		_, err := db.Exec(context.Background(), "INSERT INTO ticket_statuses (id, name) VALUES ($1, $2)",
			ts.ID, ts.Name)
		if err != nil {
			http.Error(w, "Ошибка при вставке", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(ts)
	}
}

// @Summary Обновить статус билета
// @Tags ticket-statuses
// @Accept json
// @Produce json
// @Param id path string true "ID статуса билета"
// @Param ticket_status body TicketStatus true "Обновлённые данные статуса"
// @Success 200 {object} TicketStatus
// @Failure 400 {string} string "Неверный JSON"
// @Failure 404 {string} string "Статус не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /ticket-statuses/{id} [put]
func UpdateTicketStatus(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		var ts TicketStatus
		if err := json.NewDecoder(r.Body).Decode(&ts); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		ts.ID = id

		res, err := db.Exec(context.Background(), "UPDATE ticket_statuses SET name=$1 WHERE id=$2",
			ts.Name, ts.ID)
		if err != nil {
			http.Error(w, "Ошибка при обновлении", http.StatusInternalServerError)
			return
		}
		rows := res.RowsAffected()
		if rows == 0 {
			http.Error(w, "Статус не найден", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(ts)
	}
}

// @Summary Удалить статус билета
// @Tags ticket-statuses
// @Param id path string true "ID статуса билета"
// @Success 204 {string} string "Удалено"
// @Failure 404 {string} string "Статус не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /ticket-statuses/{id} [delete]
func DeleteTicketStatus(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		res, err := db.Exec(context.Background(), "DELETE FROM ticket_statuses WHERE id = $1", id)
		if err != nil {
			http.Error(w, "Ошибка при удалении", http.StatusInternalServerError)
			return
		}
		rows := res.RowsAffected()
		if rows == 0 {
			http.Error(w, "Статус не найден", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
