package utils

import (
	validation "cw/internal/utils/validation"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteErrorHidesValidationDetails(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	validationErr := &validation.Error{
		Msg: "Invalid input",
		Fields: []validation.FieldError{
			{Field: "email", Rule: "email", Reason: "invalid email format"},
		},
		Err: nil,
	}

	WriteError(rr, NewBadRequest("Invalid input", validationErr))

	if rr.Code != 400 {
		t.Fatalf("status = %d, want 400", rr.Code)
	}

	var resp struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Message != "Invalid input" {
		t.Fatalf("message = %q, want %q", resp.Message, "Invalid input")
	}
	if strings.Contains(strings.ToLower(resp.Message), "email") {
		t.Fatalf("unexpected validation details in message: %q", resp.Message)
	}
}
