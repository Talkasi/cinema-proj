package pgerr

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestConvertErrorMapsUniqueViolation(t *testing.T) {
	t.Parallel()

	err := ConvertError(&pgconn.PgError{Code: "23505", Message: "duplicate key value violates unique constraint"})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Code != 409 {
		t.Fatalf("code = %d, want 409", err.Code)
	}
	if err.Message != "Database conflict" {
		t.Fatalf("message = %q, want %q", err.Message, "Database conflict")
	}
}

func TestConvertErrorMapsForeignKeyViolation(t *testing.T) {
	t.Parallel()

	err := ConvertError(&pgconn.PgError{Code: "23503", Message: "insert or update violates foreign key constraint"})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Code != 409 {
		t.Fatalf("code = %d, want 409", err.Code)
	}
	if err.Message != "Foreign key error" {
		t.Fatalf("message = %q, want %q", err.Message, "Foreign key error")
	}
}

func TestConvertErrorDefaultInternal(t *testing.T) {
	t.Parallel()

	err := ConvertError(errors.New("boom"))
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Code != 500 {
		t.Fatalf("code = %d, want 500", err.Code)
	}
	if err.Message != "Unknown server error" {
		t.Fatalf("message = %q", err.Message)
	}
}

func TestConvertErrorMapsNoRowsToNotFound(t *testing.T) {
	t.Parallel()

	err := ConvertError(pgx.ErrNoRows)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Code != 404 {
		t.Fatalf("code = %d, want 404", err.Code)
	}
}

func TestConvertErrorMapsCheckConstraintToBadRequest(t *testing.T) {
	t.Parallel()

	err := ConvertError(&pgconn.PgError{
		Code:    "23514",
		Message: "new row for relation violates check constraint \"valid_name\"",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Code != 400 {
		t.Fatalf("code = %d, want 400", err.Code)
	}
}
