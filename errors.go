package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx"
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

func isDivisionByZero(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "22012" // Код ошибки для деления на ноль
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

func IsError(w http.ResponseWriter, err error) bool {
	if err != nil {
		if isUniqueViolation(err) {
			http.Error(w, "Конфликт при работе с БД", http.StatusConflict)
			return true
		}
		if isPermissionDenied(err) {
			http.Error(w, "Доступ запрещен", http.StatusForbidden)
			return true
		}
		if isNoRows(err) {
			http.Error(w, "Данные не найдены", http.StatusNotFound)
			return true
		}
		if isForeignKeyViolation(err) {
			http.Error(w, "Неверный внешний ключ", http.StatusFailedDependency)
			return true
		}
		if isDataTypeMismatch(err) {
			http.Error(w, "Неверный тип", http.StatusBadRequest)
			return true
		}
		if isDivisionByZero(err) {
			http.Error(w, "Ошибка деления на ноль", http.StatusBadRequest)
			return true
		}
		if isSyntaxError(err) {
			http.Error(w, "ОШИБКА SQL ЗАПРОСА", http.StatusInternalServerError)
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
