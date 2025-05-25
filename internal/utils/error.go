package utils

import (
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
				return NewInternal("Передан null в обязательный непустой параметр", err)
			}
		case strings.Contains(err.Error(), "Невозможно запланировать показ"):
			{
				return NewConflict(err.Error(), err)
			}
		default:
			{
				return NewInternal("Неизвестная сервера", err)
			}
		}
	}

	return nil
}
