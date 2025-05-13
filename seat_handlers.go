package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Валидаторы для мест
func validateAllSeatData(w http.ResponseWriter, s SeatData) bool {
	if err := validateHallID(s.HallID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	if err := validateSeatTypeID(s.SeatTypeID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	if err := validateRowNumber(s.RowNumber); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	if err := validateSeatNumber(s.SeatNumber); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}

func validateHallID(hallID string) error {
	if _, err := uuid.Parse(hallID); err != nil {
		return errors.New("неверный формат ID зала")
	}
	return nil
}

func validateSeatTypeID(seatTypeID string) error {
	if _, err := uuid.Parse(seatTypeID); err != nil {
		return errors.New("неверный формат ID типа места")
	}
	return nil
}

func validateRowNumber(rowNumber int) error {
	if rowNumber <= 0 || rowNumber > 100 {
		return errors.New("номер ряда должен быть от 1 до 100")
	}
	return nil
}

func validateSeatNumber(seatNumber int) error {
	if seatNumber <= 0 || seatNumber > 100 {
		return errors.New("номер места должен быть от 1 до 100")
	}
	return nil
}

// @Summary Получить все места
// @Description Возвращает список всех мест, содержащихся в базе данных.
// @Tags Места
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Seat "Список мест"
// @Failure 404 {string} string "Места не найдены"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seats [get]
func GetSeats(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(context.Background(), `
			SELECT id, hall_id, seat_type_id, row_number, seat_number 
			FROM seats`)
		if HandleDatabaseError(w, err, "местами") {
			return
		}
		defer rows.Close()

		var seats []Seat
		for rows.Next() {
			var s Seat
			if err := rows.Scan(&s.ID, &s.HallID, &s.SeatTypeID, &s.RowNumber, &s.SeatNumber); HandleDatabaseError(w, err, "местом") {
				return
			}
			seats = append(seats, s)
		}

		if len(seats) == 0 {
			http.Error(w, "Места не найдены", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(seats)
	}
}

// @Summary Получить место по ID
// @Description Возвращает место по ID.
// @Tags Места
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID места"
// @Success 200 {object} Seat "Место"
// @Failure 400 {string} string "Неверный формат ID"
// @Failure 404 {string} string "Место не найдено"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seats/{id} [get]
func GetSeatByID(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var s Seat
		s.ID = id.String()
		err := db.QueryRow(context.Background(), `
			SELECT hall_id, seat_type_id, row_number, seat_number 
			FROM seats WHERE id = $1`, id).
			Scan(&s.HallID, &s.SeatTypeID, &s.RowNumber, &s.SeatNumber)

		if IsError(w, err) {
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s)
	}
}

// @Summary Создать место
// @Description Создаёт новое место.
// @Tags Места
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param seat body SeatData true "Данные места"
// @Success 201 {object} CreateResponse "ID созданного места"
// @Failure 400 {string} string "В запросе предоставлены неверные данные"
// @Failure 403 {string} string "Доступ запрещен"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seats [post]
func CreateSeat(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var s SeatData
		if !DecodeJSONBody(w, r, &s) {
			return
		}
		if !validateAllSeatData(w, s) {
			return
		}

		id := uuid.New()
		_, err := db.Exec(context.Background(), `
			INSERT INTO seats (id, hall_id, seat_type_id, row_number, seat_number) 
			VALUES ($1, $2, $3, $4, $5)`,
			id, s.HallID, s.SeatTypeID, s.RowNumber, s.SeatNumber)

		if IsError(w, err) {
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(id.String())
	}
}

// @Summary Обновить место
// @Description Обновляет существующее место.
// @Tags Места
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID места"
// @Param seat body SeatData true "Обновлённые данные места"
// @Success 200 "Данные о месте успешно обновлены"
// @Failure 400 {string} string "В запросе предоставлены неверные данные"
// @Failure 403 {string} string "Доступ запрещен"
// @Failure 404 {string} string "Место не найдено"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seats/{id} [put]
func UpdateSeat(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		var s SeatData
		if !DecodeJSONBody(w, r, &s) {
			return
		}
		if !validateAllSeatData(w, s) {
			return
		}

		res, err := db.Exec(context.Background(), `
			UPDATE seats 
			SET hall_id=$1, seat_type_id=$2, row_number=$3, seat_number=$4 
			WHERE id=$5`,
			s.HallID, s.SeatTypeID, s.RowNumber, s.SeatNumber, id)

		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w)
	}
}

// @Summary Удалить место
// @Description Удаляет место по его ID.
// @Tags Места
// @Param id path string true "ID места"
// @Security BearerAuth
// @Success 204 "Данные о месте успешно удалены"
// @Failure 400 {string} string "Неверный формат ID"
// @Failure 403 {string} string "Доступ запрещён"
// @Failure 404 {string} string "Место не найдено"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /seats/{id} [delete]
func DeleteSeat(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := ParseUUIDFromPath(w, r.PathValue("id"))
		if !ok {
			return
		}

		res, err := db.Exec(context.Background(),
			"DELETE FROM seats WHERE id = $1", id)

		if IsError(w, err) {
			return
		}

		if !CheckRowsAffected(w, res.RowsAffected()) {
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
