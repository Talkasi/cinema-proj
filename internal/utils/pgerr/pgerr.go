package pgerr

import (
	uerrors "cw/internal/utils/errors"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func isPermissionDenied(err error) bool {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return pgErr.Code == "42501"
	}

	return strings.Contains(err.Error(), "permission denied")
}

func isNoRows(err error) bool {
	if errors.Is(err, pgx.ErrNoRows) {
		return true
	}

	return err != nil && err.Error() == "no rows in result set"
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
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
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}

	return false
}

func isNotNullViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23502"
	}

	return false
}

func isCheckViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23514"
	}

	return false
}

func isInvalidTextRepresentation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "22P02"
	}

	return false
}

func isSyntaxError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "42601"
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

func ConvertError(err error) *uerrors.Error {
	if err == nil {
		return nil
	}

	if convertedErr := handleDatabaseErrors(err); convertedErr != nil {
		return convertedErr
	}

	if convertedErr := handleCheckConstraintErrors(err); convertedErr != nil {
		return convertedErr
	}

	if convertedErr := handleCustomErrors(err); convertedErr != nil {
		return convertedErr
	}

	return uerrors.NewInternal("Unknown server error", err)
}

func handleDatabaseErrors(err error) *uerrors.Error {
	switch {
	case isUniqueViolation(err):
		return uerrors.NewConflict("Database conflict", err)
	case isPermissionDenied(err):
		return uerrors.NewForbidden("Access denied", err)
	case isNoRows(err):
		return uerrors.NewNotFound("Data not found", err)
	case isForeignKeyViolation(err):
		return uerrors.NewConflict("Foreign key error", err)
	case isDataTypeMismatch(err):
		return uerrors.NewBadRequest("Invalid type", err)
	case isSyntaxError(err):
		return uerrors.NewInternal(fmt.Sprintf("SQL query error, %v\n", err), err)
	case isNotNullViolation(err):
		return uerrors.NewBadRequest("Required field is missing", err)
	case isInvalidTextRepresentation(err):
		return uerrors.NewBadRequest("Invalid data format", err)
	default:
		return nil
	}
}

func handleCheckConstraintErrors(err error) *uerrors.Error {
	if !isCheckViolation(err) {
		return nil
	}

	errorStr := strings.ToLower(err.Error())
	checkViolationMap := map[string]string{
		"valid_name":         "Invalid name",
		"valid_description":  "Invalid description",
		"email_format":       "Invalid email format",
		"valid_birth_date":   "Invalid birth date",
		"valid_duration":     "Invalid duration",
		"valid_age_limit":    "Invalid age rating",
		"box_office_revenue": "Invalid revenue",
		"start_time":         "Invalid start time",
		"price":              "Invalid price",
		"rating":             "Invalid rating",
	}

	for constraint, message := range checkViolationMap {
		if strings.Contains(errorStr, constraint) {
			return uerrors.NewBadRequest(message, err)
		}
	}

	return uerrors.NewBadRequest("Data constraint violation", err)
}

func handleCustomErrors(err error) *uerrors.Error {
	if strings.Contains(err.Error(), "Unable to schedule show") {
		return uerrors.NewConflict(err.Error(), err)
	}

	errorStr := strings.ToLower(err.Error())
	if strings.Contains(errorStr, "check constraint") {
		return uerrors.NewBadRequest("Data constraint violation", err)
	}
	if strings.Contains(errorStr, "invalid input") {
		return uerrors.NewBadRequest("Invalid input", err)
	}

	return nil
}
