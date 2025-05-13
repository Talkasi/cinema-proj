package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func isPermissionDenied(err error) bool {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return pgErr.Code == "42501"
	}

	return false
}

func isNoRows(err error) bool {
	return (err == pgx.ErrNoRows) || (errors.Is(err, pgx.ErrNoRows) || (err != nil && err.Error() == "no rows in result set"))
}

func isUniqueViolation(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "23505" // Код ошибки для уникального нарушения
	}
	return false
}

func isForeignKeyViolation(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "23503" // Код ошибки для нарушения внешнего ключа
	}
	return false
}

func isNotNullViolation(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "23502" // Код ошибки для нарушения NOT NULL
	}
	return false
}

func isSyntaxError(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "42601" // Код ошибки для синтаксической ошибки
	}
	return false
}

func isDataTypeMismatch(err error) bool {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return pgErr.Code == "42804"
	}

	return false
}

func isDeadlockDetected(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "40P01" // Код ошибки для взаимного блокирования
	}
	return false
}

func isTransactionIsolationViolation(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "40001" // Код ошибки для нарушения изоляции транзакции
	}
	return false
}

func ParseUUIDFromPath(w http.ResponseWriter, pathValue string) (uuid.UUID, bool) {
	id, err := uuid.Parse(pathValue)
	if err != nil || id.String() == "" {
		http.Error(w, "Неверный формат ID", http.StatusBadRequest)
		return uuid.Nil, false
	}
	return id, true
}

func DecodeJSONBody(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return false
	}
	return true
}

func HandleDatabaseError(w http.ResponseWriter, err error, entity string) bool {
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при работе с %s: %v", entity, err), http.StatusInternalServerError)
		return true
	}
	return false
}

func CheckRowsAffected(w http.ResponseWriter, rowsAffected int64) bool {
	if rowsAffected == 0 {
		http.Error(w, "Данные не найдены", http.StatusNotFound)
		return false
	}
	return true
}

func ValidateRequiredFields(w http.ResponseWriter, fields map[string]string) bool {
	for field, value := range fields {
		if value == "" {
			http.Error(w, fmt.Sprintf("Поле '%s' не может быть пустым", field), http.StatusBadRequest)
			return false
		}
	}
	return true
}

func IsError(w http.ResponseWriter, err error) bool {
	if err != nil {
		if isUniqueViolation(err) {
			http.Error(w, "Конфликт при работе с БД", http.StatusConflict)
			return true
		}
		if isPermissionDenied(err) {
			http.Error(w, "Доступ запрещён", http.StatusForbidden)
			return true
		}
		if isNoRows(err) {
			http.Error(w, "Данные не найдены", http.StatusNotFound)
			return true
		}
		if isForeignKeyViolation(err) {
			http.Error(w, "Ошибка внешнего ключа, скорее всего на удаляемый объект ссылаются записи другой таблицы", http.StatusFailedDependency)
			return true
		}
		if isDataTypeMismatch(err) {
			http.Error(w, "Неверный тип", http.StatusBadRequest)
			return true
		}
		if isSyntaxError(err) {
			http.Error(w, fmt.Sprintf("ОШИБКА SQL ЗАПРОСА, %v\n", err), http.StatusInternalServerError)
			return true
		}
		if isNotNullViolation(err) {
			http.Error(w, "Передан null в обязательный непустой параметр", http.StatusInternalServerError)
			return true
		}

		http.Error(w, fmt.Sprintf("Ошибка при вставке %v\n", err), http.StatusInternalServerError)
		return true
	}

	return false
}
