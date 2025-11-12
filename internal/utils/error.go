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

	// Дополнительные проверки для pgx
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
	if err != nil {
		switch {
		case isUniqueViolation(err):
			{
				return NewConflict("Конфликт при работе с БД", err)
			}
		case isPermissionDenied(err):
			{
				return NewForbidden("Доступ запрещён", err)
			}
		case isNoRows(err):
			{
				return NewNotFound("Данные не найдены", err)
			}
		case isForeignKeyViolation(err):
			{
				return NewConflict("Ошибка внешнего ключа", err)
			}
		case isDataTypeMismatch(err):
			{
				return NewBadRequest("Неверный тип", err)
			}
		case isSyntaxError(err):
			{
				return NewInternal(fmt.Sprintf("Ошибка SQL запроса, %v\n", err), err)
			}
		case isNotNullViolation(err):
			{
				return NewBadRequest("Обязательное поле не заполнено", err)
			}
		case isCheckViolation(err):
			{
				// Анализируем какое именно check ограничение нарушено
				errorStr := strings.ToLower(err.Error())
				switch {
				case strings.Contains(errorStr, "valid_name"):
					return NewBadRequest("Некорректное имя", err)
				case strings.Contains(errorStr, "valid_description"):
					return NewBadRequest("Некорректное описание", err)
				case strings.Contains(errorStr, "email_format"):
					return NewBadRequest("Некорректный формат email", err)
				case strings.Contains(errorStr, "valid_birth_date"):
					return NewBadRequest("Некорректная дата рождения", err)
				case strings.Contains(errorStr, "valid_duration"):
					return NewBadRequest("Некорректная длительность", err)
				case strings.Contains(errorStr, "valid_age_limit"):
					return NewBadRequest("Некорректный возрастной рейтинг", err)
				case strings.Contains(errorStr, "box_office_revenue"):
					return NewBadRequest("Некорректная выручка", err)
				case strings.Contains(errorStr, "start_time"):
					return NewBadRequest("Некорректное время начала", err)
				case strings.Contains(errorStr, "price"):
					return NewBadRequest("Некорректная цена", err)
				case strings.Contains(errorStr, "rating"):
					return NewBadRequest("Некорректный рейтинг", err)
				default:
					return NewBadRequest("Нарушение ограничений данных", err)
				}
			}
		case isInvalidTextRepresentation(err):
			{
				return NewBadRequest("Некорректный формат данных", err)
			}
		case strings.Contains(err.Error(), "Невозможно запланировать показ"):
			{
				return NewConflict(err.Error(), err)
			}
		default:
			{
				// Дополнительные проверки по тексту ошибки для обратной совместимости
				errorStr := strings.ToLower(err.Error())
				if strings.Contains(errorStr, "check constraint") {
					return NewBadRequest("Нарушение ограничений данных", err)
				}
				if strings.Contains(errorStr, "invalid input") {
					return NewBadRequest("Некорректные входные данные", err)
				}
				return NewInternal("Неизвестная ошибка сервера", err)
			}
		}
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
	json.NewEncoder(w).Encode(dto.ErrorResponse{Message: message})
}
