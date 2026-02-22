package utils

import (
	dto "cw/internal/dto/models"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Error struct {
	Message string
	Code    int
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func NewBadRequest(msg string, err error) *Error {
	return &Error{Message: msg, Code: http.StatusBadRequest, Err: err}
}

func NewNotFound(msg string, err error) *Error {
	return &Error{Message: msg, Code: http.StatusNotFound, Err: err}
}

func NewConflict(msg string, err error) *Error {
	return &Error{Message: msg, Code: http.StatusConflict, Err: err}
}

func NewForbidden(msg string, err error) *Error {
	return &Error{Message: msg, Code: http.StatusForbidden, Err: err}
}

func NewInternal(msg string, err error) *Error {
	return &Error{Message: msg, Code: http.StatusInternalServerError, Err: err}
}

func isPermissionDenied(err error) bool {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return pgErr.Code == "42501"
	}

	if strings.Contains(err.Error(), "permission denied") {
		return true
	}

	return false
}

func isNoRows(err error) bool {
	if errors.Is(err, pgx.ErrNoRows) {
		return true
	}

	if err != nil && err.Error() == "no rows in result set" {
		return true
	}

	return false
}

func isUniqueViolation(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "23505" // Код ошибки для уникального нарушения
	}
	return false
}

func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := strings.ToLower(err.Error())
	return strings.Contains(errorStr, "duplicate") ||
		strings.Contains(errorStr, "unique") ||
		strings.Contains(errorStr, "23505")
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

func isCheckViolation(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "23514" // Код ошибки для нарушения CHECK ограничения
	}
	return false
}

func isInvalidTextRepresentation(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "22P02" // Код ошибки для неверного формата текста
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
		return pgErr.Code == "42804" // Код ошибки несоответствия типов
	}

	return false
}

func ConvertError(err error) *Error {
	if err == nil {
		return nil
	}

	// Check for specific database-related errors first
	if convertedErr := handleDatabaseErrors(err); convertedErr != nil {
		return convertedErr
	}

	// Check for check constraint violations
	if convertedErr := handleCheckConstraintErrors(err); convertedErr != nil {
		return convertedErr
	}

	// Check for other custom error patterns
	if convertedErr := handleCustomErrors(err); convertedErr != nil {
		return convertedErr
	}

	// Default error
	return NewInternal("Неизвестная ошибка сервера", err)
}

// handleDatabaseErrors checks for common database errors
func handleDatabaseErrors(err error) *Error {
	switch {
	case isUniqueViolation(err):
		return NewConflict("Конфликт при работе с БД", err)
	case isPermissionDenied(err):
		return NewForbidden("Доступ запрещён", err)
	case isNoRows(err):
		return NewNotFound("Данные не найдены", err)
	case isForeignKeyViolation(err):
		return NewConflict("Ошибка внешнего ключа", err)
	case isDataTypeMismatch(err):
		return NewBadRequest("Неверный тип", err)
	case isSyntaxError(err):
		return NewInternal(fmt.Sprintf("Ошибка SQL запроса, %v\n", err), err)
	case isNotNullViolation(err):
		return NewBadRequest("Обязательное поле не заполнено", err)
	case isInvalidTextRepresentation(err):
		return NewBadRequest("Некорректный формат данных", err)
	default:
		return nil
	}
}

// handleCheckConstraintErrors handles specific check constraint violations
func handleCheckConstraintErrors(err error) *Error {
	if !isCheckViolation(err) {
		return nil
	}

	errorStr := strings.ToLower(err.Error())
	checkViolationMap := map[string]string{
		"valid_name":         "Некорректное имя",
		"valid_description":  "Некорректное описание",
		"email_format":       "Некорректный формат email",
		"valid_birth_date":   "Некорректная дата рождения",
		"valid_duration":     "Некорректная длительность",
		"valid_age_limit":    "Некорректный возрастной рейтинг",
		"box_office_revenue": "Некорректная выручка",
		"start_time":         "Некорректное время начала",
		"price":              "Некорректная цена",
		"rating":             "Некорректный рейтинг",
	}

	for constraint, message := range checkViolationMap {
		if strings.Contains(errorStr, constraint) {
			return NewBadRequest(message, err)
		}
	}

	return NewBadRequest("Нарушение ограничений данных", err)
}

// handleCustomErrors handles special custom error cases
func handleCustomErrors(err error) *Error {
	if strings.Contains(err.Error(), "Невозможно запланировать показ") {
		return NewConflict(err.Error(), err)
	}

	errorStr := strings.ToLower(err.Error())
	if strings.Contains(errorStr, "check constraint") {
		return NewBadRequest("Нарушение ограничений данных", err)
	}
	if strings.Contains(errorStr, "invalid input") {
		return NewBadRequest("Некорректные входные данные", err)
	}

	return nil
}

func WriteError(w http.ResponseWriter, err *Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)

	message := err.Message
	if err.Err != nil {
		message += " " + err.Err.Error()
	}
	if err := json.NewEncoder(w).Encode(dto.ErrorResponse{Message: message}); err != nil {
		// If we can't write the error response, log it and continue
		fmt.Printf("Failed to encode error response: %v\n", err)
	}
}
